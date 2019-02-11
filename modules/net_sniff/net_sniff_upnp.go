package net_sniff

import (
	"fmt"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func upnpParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	if data := packets.UPNPGetMeta(pkt); data != nil && len(data) > 0 {
		s := ""
		for name, value := range data {
			s += fmt.Sprintf("%s:%s ", tui.Blue(name), tui.Yellow(value))
		}

		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"upnp",
			ip.SrcIP.String(),
			ip.DstIP.String(),
			nil,
			"%s %s -> %s : %s",
			tui.Wrap(tui.BACKRED+tui.FOREBLACK, "upnp"),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			str.Trim(s),
		).Push()

		return true
	}

	return false
}
