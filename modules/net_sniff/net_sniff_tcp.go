package net_sniff

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var tcpParsers = []func(net.IP, net.IP, []byte, gopacket.Packet, *layers.TCP) bool{
	sniParser,
	ntlmParser,
	httpParser,
	ftpParser,
	teamViewerParser,
}

func onTCP(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, verbose bool) {
	tcp := pkt.Layer(layers.LayerTypeTCP).(*layers.TCP)
	for _, parser := range tcpParsers {
		if parser(srcIP, dstIP, payload, pkt, tcp) {
			return
		}
	}

	if verbose {
		sz := len(payload)
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"tcp",
			fmt.Sprintf("%s:%s", srcIP, vPort(tcp.SrcPort)),
			fmt.Sprintf("%s:%s", dstIP, vPort(tcp.DstPort)),
			SniffData{
				"Size": len(payload),
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKLIGHTBLUE+tui.FOREBLACK, "tcp"),
			vIP(srcIP),
			vPort(tcp.SrcPort),
			vIP(dstIP),
			vPort(tcp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", sz)),
		).Push()
	}
}
