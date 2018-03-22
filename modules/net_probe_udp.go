package modules

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/log"
)

func (p *Prober) sendProbeUDP(from net.IP, from_hw net.HardwareAddr, ip net.IP) {
	name := fmt.Sprintf("%s:137", ip)
	if addr, err := net.ResolveUDPAddr("udp", name); err != nil {
		log.Error("Could not resolve %s.", name)
	} else if con, err := net.DialUDP("udp", nil, addr); err != nil {
		log.Error("Could not dial %s.", name)
	} else {
		// log.Debug("UDP connection to %s enstablished.", name)
		defer con.Close()
		wrote, _ := con.Write([]byte{0x00})

		if wrote > 0 {
			p.Session.Queue.TrackSent(uint64(wrote))
		} else {
			p.Session.Queue.TrackError()
		}
	}
}
