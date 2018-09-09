package modules

import (
	"net"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
)

func (p *Prober) sendProbeUPNP(from net.IP, from_hw net.HardwareAddr) {
	err, raw := packets.NewUPNPProbe(from, from_hw)
	if err != nil {
		log.Error("error while sending upnp probe: %v", err)
		return
	} else if err := p.Session.Queue.Send(raw); err != nil {
		log.Error("error sending upnp packet: %s", err)
	} else {
		log.Debug("sent %d bytes of UPNP probe", len(raw))
	}
}
