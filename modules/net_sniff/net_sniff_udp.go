package net_sniff

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var udpParsers = []func(*layers.IPv4, gopacket.Packet, *layers.UDP) bool{
	dnsParser,
	mdnsParser,
	krb5Parser,
	upnpParser,
}

func onUDP(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)
	for _, parser := range udpParsers {
		if parser(ip, pkt, udp) {
			return
		}
	}

	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"udp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(udp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(udp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "udp"),
			vIP(ip.SrcIP),
			vPort(udp.SrcPort),
			vIP(ip.DstIP),
			vPort(udp.DstPort),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}
