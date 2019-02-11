package net_sniff

import (
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func teamViewerParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	if tcp.SrcPort == packets.TeamViewerPort || tcp.DstPort == packets.TeamViewerPort {
		if tv := packets.ParseTeamViewer(tcp.Payload); tv != nil {
			NewSnifferEvent(
				pkt.Metadata().Timestamp,
				"teamviewer",
				ip.SrcIP.String(),
				ip.DstIP.String(),
				nil,
				"%s %s %s > %s",
				tui.Wrap(tui.BACKYELLOW+tui.FOREWHITE, "teamviewer"),
				vIP(ip.SrcIP),
				tui.Yellow(tv.Command),
				vIP(ip.DstIP),
			).Push()
			return true
		}
	}

	return false
}
