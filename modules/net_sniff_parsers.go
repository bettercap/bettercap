package modules

import (
	"fmt"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func tcpParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	tcp := pkt.Layer(layers.LayerTypeTCP).(*layers.TCP)

	if sniParser(ip, pkt, tcp) {
		return
	} else if ntlmParser(ip, pkt, tcp) {
		return
	} else if httpParser(ip, pkt, tcp) {
		return
	} else if verbose == true {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"tcp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(tcp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(tcp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			core.W(core.BG_LBLUE+core.FG_BLACK, "tcp"),
			vIP(ip.SrcIP),
			vPort(tcp.SrcPort),
			vIP(ip.DstIP),
			vPort(tcp.DstPort),
			core.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func udpParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)

	if dnsParser(ip, pkt, udp) {
		return
	} else if krb5Parser(ip, pkt, udp) {
		return
	} else if verbose == true {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"udp",
			fmt.Sprintf("%s:%s", ip.SrcIP, vPort(udp.SrcPort)),
			fmt.Sprintf("%s:%s", ip.DstIP, vPort(udp.DstPort)),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s:%s > %s:%s %s",
			core.W(core.BG_DGRAY+core.FG_WHITE, "udp"),
			vIP(ip.SrcIP),
			vPort(udp.SrcPort),
			vIP(ip.DstIP),
			vPort(udp.DstPort),
			core.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func unkParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	if verbose == true {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			pkt.TransportLayer().LayerType().String(),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s > %s %s",
			core.W(core.BG_DGRAY+core.FG_WHITE, pkt.TransportLayer().LayerType().String()),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			core.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
		).Push()
	}
}

func mainParser(pkt gopacket.Packet, verbose bool) bool {
	nlayer := pkt.NetworkLayer()
	if nlayer == nil {
		log.Debug("Missing network layer skipping packet.")
		return false
	}

	if nlayer.LayerType() != layers.LayerTypeIPv4 {
		log.Debug("Unexpected layer type %s, skipping packet.", nlayer.LayerType())
		return false
	}

	ip := nlayer.(*layers.IPv4)

	tlayer := pkt.TransportLayer()
	if tlayer == nil {
		log.Debug("Missing transport layer skipping packet.")
		return false
	}

	if tlayer.LayerType() == layers.LayerTypeTCP {
		tcpParser(ip, pkt, verbose)
	} else if tlayer.LayerType() == layers.LayerTypeUDP {
		udpParser(ip, pkt, verbose)
	} else {
		unkParser(ip, pkt, verbose)
	}

	return true
}
