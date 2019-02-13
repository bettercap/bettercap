package net_probe

import (
	"net"

	"github.com/bettercap/bettercap/packets"
)

func (mod *Prober) sendProbeMDNS(from net.IP, from_hw net.HardwareAddr) {
	err, raw := packets.NewMDNSProbe(from, from_hw)
	if err != nil {
		mod.Error("error while sending mdns probe: %v", err)
		return
	} else if err := mod.Session.Queue.Send(raw); err != nil {
		mod.Error("error sending mdns packet: %s", err)
	} else {
		mod.Debug("sent %d bytes of MDNS probe", len(raw))
	}
}
