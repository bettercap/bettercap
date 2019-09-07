package net_sniff

import (
	"fmt"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func onUNK(ip *layers.IPv4, pkt gopacket.Packet, verbose bool) {
	if verbose {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			pkt.TransportLayer().LayerType().String(),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			SniffData{
				"Size": len(ip.Payload),
			},
			"%s %s > %s %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, pkt.TransportLayer().LayerType().String()),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			tui.Dim(fmt.Sprintf("%d bytes", len(ip.Payload))),
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
		if nlayer.LayerType() != layers.LayerTypeIPv4 {
			log.Debug("Unexpected layer type %s, skipping packet.", nlayer.LayerType())
			log.Debug("%s", pkt.Dump())
			return false
		}

		ip := nlayer.(*layers.IPv4)

		tlayer := pkt.TransportLayer()
		if tlayer == nil {
			log.Debug("Missing transport layer skipping packet.")
			log.Debug("%s", pkt.Dump())
			return false
		}

		if tlayer.LayerType() == layers.LayerTypeTCP {
			onTCP(ip, pkt, verbose)
		} else if tlayer.LayerType() == layers.LayerTypeUDP {
			onUDP(ip, pkt, verbose)
		} else {
			onUNK(ip, pkt, verbose)
		}
		return true
	} else if ok, radiotap, dot11 := packets.Dot11Parse(pkt); ok {
		// are we sniffing in monitor mode?
		onDOT11(radiotap, dot11, pkt, verbose)
		return true
	}
	return false
}
