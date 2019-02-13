package wol

import (
	"fmt"
	"net"
	"regexp"

	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

var (
	reMAC = regexp.MustCompile(`^([0-9a-fA-F]{2}[:-]){5}([0-9a-fA-F]{2})$`)
)

type WOL struct {
	session.SessionModule
}

func NewWOL(s *session.Session) *WOL {
	mod := &WOL{
		SessionModule: session.NewSessionModule("wol", s),
	}

	mod.AddHandler(session.NewModuleHandler("wol.eth MAC", "wol.eth(\\s.+)?",
		"Send a WOL as a raw ethernet packet of type 0x0847 (if no MAC is specified, ff:ff:ff:ff:ff:ff will be used).",
		func(args []string) error {
			if mac, err := parseMAC(args); err != nil {
				return err
			} else {
				return mod.wolETH(mac)
			}
		}))

	mod.AddHandler(session.NewModuleHandler("wol.udp MAC", "wol.udp(\\s.+)?",
		"Send a WOL as an IPv4 broadcast packet to UDP port 9 (if no MAC is specified, ff:ff:ff:ff:ff:ff will be used).",
		func(args []string) error {
			if mac, err := parseMAC(args); err != nil {
				return err
			} else {
				return mod.wolUDP(mac)
			}
		}))

	return mod
}

func parseMAC(args []string) (string, error) {
	mac := "ff:ff:ff:ff:ff:ff"
	if len(args) == 1 {
		tmp := str.Trim(args[0])
		if tmp != "" {
			if !reMAC.MatchString(tmp) {
				return "", fmt.Errorf("%s is not a valid MAC address.", tmp)
			} else {
				mac = tmp
			}
		}
	}

	return mac, nil
}

func (mod *WOL) Name() string {
	return "wol"
}

func (mod *WOL) Description() string {
	return "A module to send Wake On LAN packets in broadcast or to a specific MAC."
}

func (mod *WOL) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *WOL) Configure() error {
	return nil
}

func (mod *WOL) Start() error {
	return nil
}

func (mod *WOL) Stop() error {
	return nil
}

func buildPayload(mac string) []byte {
	raw, _ := net.ParseMAC(mac)
	payload := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for i := 0; i < 16; i++ {
		payload = append(payload, raw...)
	}
	return payload
}

func (mod *WOL) wolETH(mac string) error {
	mod.SetRunning(true, nil)
	defer mod.SetRunning(false, nil)

	payload := buildPayload(mac)
	mod.Info("sending %d bytes of ethernet WOL packet to %s", len(payload), tui.Bold(mac))
	eth := layers.Ethernet{
		SrcMAC:       mod.Session.Interface.HW,
		DstMAC:       layers.EthernetBroadcast,
		EthernetType: 0x0842,
	}

	err, raw := packets.Serialize(&eth)
	if err != nil {
		return err
	}

	raw = append(raw, payload...)
	return mod.Session.Queue.Send(raw)
}

func (mod *WOL) wolUDP(mac string) error {
	mod.SetRunning(true, nil)
	defer mod.SetRunning(false, nil)

	payload := buildPayload(mac)
	mod.Info("sending %d bytes of UDP WOL packet to %s", len(payload), tui.Bold(mac))

	eth := layers.Ethernet{
		SrcMAC:       mod.Session.Interface.HW,
		DstMAC:       layers.EthernetBroadcast,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    mod.Session.Interface.IP,
		DstIP:    net.ParseIP("255.255.255.255"),
	}

	udp := layers.UDP{
		SrcPort: layers.UDPPort(32767),
		DstPort: layers.UDPPort(9),
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	err, raw := packets.Serialize(&eth, &ip4, &udp)
	if err != nil {
		return err
	}

	raw = append(raw, payload...)
	return mod.Session.Queue.Send(raw)
}
