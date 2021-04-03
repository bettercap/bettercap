package packets

import (
	"github.com/google/gopacket/layers"
	"net"
)

func ICMP6RouterAdvertisement(srcHW net.HardwareAddr, srcIP net.IP, dstHW net.HardwareAddr, dstIP net.IP, routerIP net.IP) (error, []byte) {
	eth := layers.Ethernet{
		SrcMAC:       srcHW,
		DstMAC:       dstHW,
		EthernetType: layers.EthernetTypeIPv6,
	}
	ip6 := layers.IPv6{
		NextHeader:   layers.IPProtocolICMPv6,
		TrafficClass: 224,
		Version:      6,
		HopLimit:     255,
		SrcIP:        srcIP,
		DstIP:        dstIP,
	}
	icmp6 := layers.ICMPv6{
		TypeCode: layers.ICMPv6TypeNeighborAdvertisement << 8,
	}
	adv := layers.ICMPv6NeighborAdvertisement{
		Flags: 0x20 | 0x40, // solicited && override
		TargetAddress: routerIP,
		Options: []layers.ICMPv6Option{
			{
				Type: layers.ICMPv6OptTargetAddress,
				Data: srcHW,
			},
		},
	}
	icmp6.SetNetworkLayerForChecksum(&ip6)

	return Serialize(&eth, &ip6, &icmp6, &adv)
}
