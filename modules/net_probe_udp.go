package modules

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/log"
)

// NBNS port
const NBNSPort = 137

// NBNS hostname resolution request buffer.
var NBNSRequest = []byte{
	0x82, 0x28, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x20, 0x43, 0x4B, 0x41, 0x41, 0x41, 0x41,
	0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
	0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
	0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x0,
	0x0, 0x21, 0x0, 0x1,
}

func (p *Prober) sendProbeUDP(from net.IP, from_hw net.HardwareAddr, ip net.IP) {
	name := fmt.Sprintf("%s:%d", ip, NBNSPort)
	if addr, err := net.ResolveUDPAddr("udp", name); err != nil {
		log.Debug("could not resolve %s.", name)
	} else if con, err := net.DialUDP("udp", nil, addr); err != nil {
		log.Debug("could not dial %s.", name)
	} else {
		log.Debug("udp connection to %s enstablished.", name)

		buffer := make([]byte, 0xff)
		defer con.Close()
		wrote, _ := con.Write(NBNSRequest)

		log.Info("wrote %d bytes", len(NBNSRequest))

		read, _, _ := con.ReadFrom(buffer)

		log.Info("got %d bytes of buffer", len(buffer))

		if wrote > 0 {
			p.Session.Queue.TrackSent(uint64(wrote))
		} else {
			p.Session.Queue.TrackError()
		}

		if read > 0 {
			p.Session.Queue.TrackPacket(uint64(read))
		}
	}
}
