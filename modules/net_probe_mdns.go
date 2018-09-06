package modules

import (
	"net"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
)

func (p *Prober) sendProbeMDNS(from net.IP, from_hw net.HardwareAddr) {
	err, raw := packets.NewMDNSProbe(from, from_hw)
	if err != nil {
		log.Error("error while sending mdns probe: %v", err)
		return
	}

	log.Debug("sending %d bytes of mdns probe query", len(raw))
	if err := p.Session.Queue.Send(raw); err != nil {
		log.Error("error sending mdns packet: %s", err)
	}
}
