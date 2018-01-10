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
	}

	if verbose == true {
		fmt.Printf("[%s] %s %s:%s > %s:%s %s\n",
			vTime(pkt.Metadata().Timestamp),
			core.W(core.BG_LBLUE+core.FG_BLACK, "tcp"),
			vIP(ip.SrcIP),
			vPort(tcp.SrcPort),
			vIP(ip.DstIP),
			vPort(tcp.DstPort),
			core.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))))
	}
}

func udpParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	udp := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP)

	if dnsParser(ip, pkt, udp) {
		return
	}

	if verbose == true {
		fmt.Printf("[%s] %s %s:%s > %s:%s %s\n",
			vTime(pkt.Metadata().Timestamp),
			core.W(core.BG_DGRAY+core.FG_WHITE, "udp"),
			vIP(ip.SrcIP),
			vPort(udp.SrcPort),
			vIP(ip.DstIP),
			vPort(udp.DstPort),
			core.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))))
	}
}

func unkParser(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	if verbose == true {
		fmt.Printf("[%s] [%s] %s > %s (%d bytes)\n",
			vTime(pkt.Metadata().Timestamp),
			pkt.TransportLayer().LayerType(),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			len(ip.Payload))
	}
}

func mainParser(pkt gopacket.Packet, verbose bool) bool {
	nlayer := pkt.NetworkLayer()
	if nlayer == nil {
		log.Warning("Missing network layer skipping packet.")
		return false
	}

	if nlayer.LayerType() != layers.LayerTypeIPv4 {
		log.Warning("Unexpected layer type %s, skipping packet.", nlayer.LayerType())
		return false
	}

	ip := nlayer.(*layers.IPv4)

	tlayer := pkt.TransportLayer()
	if tlayer == nil {
		log.Warning("Missing transport layer skipping packet.")
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
