package prober

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/packets"
)

func (p *Prober) sendProbeUPNP(from net.IP, from_hw net.HardwareAddr) {
	name := fmt.Sprintf("%s:%d", packets.UPNPDestIP, packets.UPNPPort)
	if addr, err := net.ResolveUDPAddr("udp", name); err != nil {
		p.Debug("could not resolve %s.", name)
	} else if con, err := net.DialUDP("udp", nil, addr); err != nil {
		p.Debug("could not dial %s.", name)
	} else {
		defer con.Close()
		if wrote, _ := con.Write(packets.UPNPDiscoveryPayload); wrote > 0 {
			p.Session.Queue.TrackSent(uint64(wrote))
		} else {
			p.Session.Queue.TrackError()
		}
	}

}
