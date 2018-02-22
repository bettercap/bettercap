package modules

import (
	"fmt"
	"net"
	"regexp"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket/layers"
)

var (
	reMAC = regexp.MustCompile(`^([0-9a-fA-F]{2}[:-]){5}([0-9a-fA-F]{2})$`)
)

type WOL struct {
	session.SessionModule
}

func NewWOL(s *session.Session) *WOL {
	w := &WOL{
		SessionModule: session.NewSessionModule("wol", s),
	}

	w.AddHandler(session.NewModuleHandler("wol.eth MAC", "wol.eth(\\s.+)?",
		"Send a WOL as a raw ethernet packet of type 0x0847 (if no MAC is specified, ff:ff:ff:ff:ff:ff will be used).",
		func(args []string) error {
			if mac, err := parseMAC(args); err != nil {
				return err
			} else {
				return w.wolETH(mac)
			}
		}))

	w.AddHandler(session.NewModuleHandler("wol.udp MAC", "wol.udp(\\s.+)?",
		"Send a WOL as an IPv4 broadcast packet to UDP port 9 (if no MAC is specified, ff:ff:ff:ff:ff:ff will be used).",
		func(args []string) error {
			if mac, err := parseMAC(args); err != nil {
				return err
			} else {
				return w.wolUDP(mac)
			}
		}))

	return w
}

func parseMAC(args []string) (string, error) {
	mac := "ff:ff:ff:ff:ff:ff"
	if len(args) == 1 {
		tmp := core.Trim(args[0])
		if tmp != "" {
			if reMAC.MatchString(tmp) == false {
				return "", fmt.Errorf("%s is not a valid MAC address.", tmp)
			} else {
				mac = tmp
			}
		}
	}

	return mac, nil
}

func (w *WOL) Name() string {
	return "wol"
}

func (w *WOL) Description() string {
	return "A module to send Wake On LAN packets in broadcast or to a specific MAC."
}

func (w *WOL) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (w *WOL) Configure() error {
	return nil
}

func (w *WOL) Start() error {
	return nil
}

func (w *WOL) Stop() error {
	return nil
}

func buildPayload(mac string) []byte {
	raw, _ := net.ParseMAC(mac)
	payload := make([]byte, 0)

	payload = append(payload, layers.EthernetBroadcast...)
	for i := 0; i < 16; i++ {
		payload = append(payload, raw...)
	}
	return payload
}

func (w *WOL) wolETH(mac string) error {
	w.SetRunning(true, nil)
	defer w.SetRunning(false, nil)

	payload := buildPayload(mac)
	log.Info("Sending %d bytes of ethernet WOL packet to %s", len(payload), core.Bold(mac))
	eth := layers.Ethernet{
		SrcMAC:       w.Session.Interface.HW,
		DstMAC:       layers.EthernetBroadcast,
		EthernetType: 0x0842,
	}

	err, raw := packets.Serialize(&eth)
	if err != nil {
		return err
	}

	raw = append(raw, payload...)
	if err := w.Session.Queue.Send(raw); err != nil {
		return err
	}

	return nil
}

func (w *WOL) wolUDP(mac string) error {
	w.SetRunning(true, nil)
	defer w.SetRunning(false, nil)

	payload := buildPayload(mac)
	log.Info("Sending %d bytes of UDP WOL packet to %s", len(payload), core.Bold(mac))

	eth := layers.Ethernet{
		SrcMAC:       w.Session.Interface.HW,
		DstMAC:       layers.EthernetBroadcast,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    w.Session.Interface.IP,
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
	if err := w.Session.Queue.Send(raw); err != nil {
		return err
	}

	return nil
}
