package net_sniff

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var tcpParsers = []func(*layers.IPv4, gopacket.Packet, *layers.TCP) bool{
	sniParser,
	ntlmParser,
	httpParser,
	ftpParser,
	teamViewerParser,
}

func onTCP(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	tcp := pkt.Layer(layers.LayerTypeTCP).(*layers.TCP)
	for _, parser := range tcpParsers {
		if parser(ip, pkt, tcp) {
			return
		}
	}

	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"tcp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(tcp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(tcp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKLIGHTBLUE+tui.FOREBLACK, "tcp"),
			vIP(ip.SrcIP),
			vPort(tcp.SrcPort),
			vIP(ip.DstIP),
			vPort(tcp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}
