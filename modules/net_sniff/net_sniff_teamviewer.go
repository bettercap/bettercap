package net_sniff

import (
	"net"

	"github.com/bettercap/bettercap/v2/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func teamViewerParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, tcp *layers.TCP) bool {
	if tcp.SrcPort == packets.TeamViewerPort || tcp.DstPort == packets.TeamViewerPort {
		if tv := packets.ParseTeamViewer(tcp.Payload); tv != nil {
			NewSnifferEvent(
				pkt.Metadata().Timestamp,
				"teamviewer",
				srcIP.String(),
				dstIP.String(),
				nil,
				"%s %s %s > %s",
				tui.Wrap(tui.BACKYELLOW+tui.FOREWHITE, "teamviewer"),
				vIP(srcIP),
				tui.Yellow(tv.Command),
				vIP(dstIP),
			).Push()
			return true
		}
	}

	return false
}
