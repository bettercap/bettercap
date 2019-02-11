package prober

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
	} else if err := p.Session.Queue.Send(raw); err != nil {
		log.Error("error sending mdns packet: %s", err)
	} else {
		log.Debug("sent %d bytes of MDNS probe", len(raw))
	}
}
