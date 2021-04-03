package net_sniff

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func onUNK(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, verbose bool) {
	if verbose {
		sz := len(payload)
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			pkt.TransportLayer().LayerType().String(),
			vIP(srcIP),
			vIP(dstIP),
			SniffData{
				"Size": sz,
			},
			"%s %s > %s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, pkt.TransportLayer().LayerType().String()),
			vIP(srcIP),
			vIP(dstIP),
			tui.Dim(fmt.Sprintf("%d bytes", sz)),
		).Push()
	}
}

func mainParser(pkt gopacket.Packet, verbose bool) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Warning("error while parsing packet: %v", err)
		}
	}()

	// simple networking sniffing mode?
	nlayer := pkt.NetworkLayer()
	if nlayer != nil {
		isIPv4 := nlayer.LayerType() == layers.LayerTypeIPv4
		isIPv6 := nlayer.LayerType() == layers.LayerTypeIPv6

		if !isIPv4 && !isIPv6 {
			log.Debug("Unexpected layer type %s, skipping packet.", nlayer.LayerType())
			log.Debug("%s", pkt.Dump())
			return false
		}

		var srcIP, dstIP net.IP
		var basePayload []byte

		if isIPv4 {
			ip := nlayer.(*layers.IPv4)
			srcIP = ip.SrcIP
			dstIP = ip.DstIP
			basePayload = ip.Payload
		} else {
			ip := nlayer.(*layers.IPv6)
			srcIP = ip.SrcIP
			dstIP = ip.DstIP
			basePayload = ip.Payload
		}

		tlayer := pkt.TransportLayer()
		if tlayer == nil {
			log.Debug("Missing transport layer skipping packet.")
			log.Debug("%s", pkt.Dump())
			return false
		}

		if tlayer.LayerType() == layers.LayerTypeTCP {
			onTCP(srcIP, dstIP, basePayload, pkt, verbose)
		} else if tlayer.LayerType() == layers.LayerTypeUDP {
			onUDP(srcIP, dstIP, basePayload, pkt, verbose)
		} else {
			onUNK(srcIP, dstIP, basePayload, pkt, verbose)
		}
		return true
	} else if ok, radiotap, dot11 := packets.Dot11Parse(pkt); ok {
		// are we sniffing in monitor mode?
		onDOT11(radiotap, dot11, pkt, verbose)
		return true
	}
	return false
}
