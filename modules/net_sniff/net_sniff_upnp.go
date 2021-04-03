package net_sniff

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func upnpParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, udp *layers.UDP) bool {
	if data := packets.UPNPGetMeta(pkt); len(data) > 0 {
		s := ""
		for name, value := range data {
			s += fmt.Sprintf("%s:%s ", tui.Blue(name), tui.Yellow(value))
		}

		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"upnp",
			srcIP.String(),
			dstIP.String(),
			nil,
			"%s %s -> %s : %s",
			tui.Wrap(tui.BACKRED+tui.FOREBLACK, "upnp"),
			vIP(srcIP),
			vIP(dstIP),
			str.Trim(s),
		).Push()

		return true
	}

	return false
}
