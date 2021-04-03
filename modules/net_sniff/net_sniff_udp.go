package net_sniff

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var udpParsers = []func(net.IP, net.IP, []byte, gopacket.Packet, *layers.UDP) bool{
	dnsParser,
	mdnsParser,
	krb5Parser,
	upnpParser,
}

func onUDP(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, verbose bool) {
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)
	for _, parser := range udpParsers {
		if parser(srcIP, dstIP, payload, pkt, udp) {
			return
		}
	}

	if verbose {
		sz := len(payload)
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"udp",
			fmt.Sprintf("%s:%s", srcIP, vPort(udp.SrcPort)),
			fmt.Sprintf("%s:%s", dstIP, vPort(udp.DstPort)),
			SniffData{
				"Size": sz,
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "udp"),
			vIP(srcIP),
			vPort(udp.SrcPort),
			vIP(dstIP),
			vPort(udp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", sz)),
		).Push()
	}
}
