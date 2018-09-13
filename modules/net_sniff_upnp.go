package modules

import (
	"fmt"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func upnpParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	if data := packets.UPNPGetMeta(pkt); data != nil && len(data) > 0 {
		s := ""
		for name, value := range data {
			s += fmt.Sprintf("%s:%s ", core.Blue(name), core.Yellow(value))
		}

		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"upnp",
			ip.SrcIP.String(),
			ip.DstIP.String(),
			nil,
			"%s %s -> %s : %s",
			core.W(core.BG_RED+core.FG_BLACK, "upnp"),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			core.Trim(s),
		).Push()

		return true
	}

	return false
}
