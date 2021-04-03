package packets

import (
	"github.com/google/gopacket/layers"
	"net"
)

func ICMP6NeighborAdvertisement(srcHW net.HardwareAddr, srcIP net.IP, dstHW net.HardwareAddr, dstIP net.IP, routerIP net.IP) (error, []byte) {
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
		Flags:         0x20 | 0x40, // solicited && override
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

var macIpv6Multicast = net.HardwareAddr([]byte{0x33, 0x33, 0x00, 0x00, 0x00, 0x01})
var ipv6Multicast = net.ParseIP("ff02::1")

func ICMP6RouterAdvertisement(ip net.IP, hw net.HardwareAddr, prefix string, prefixLength uint8) (error, []byte) {
	eth := layers.Ethernet{
		SrcMAC:       hw,
		DstMAC:       macIpv6Multicast,
		EthernetType: layers.EthernetTypeIPv6,
	}
	ip6 := layers.IPv6{
		NextHeader:   layers.IPProtocolICMPv6,
		TrafficClass: 224,
		Version:      6,
		HopLimit:     255,
		SrcIP:        ip,
		DstIP:        ipv6Multicast,
	}
	icmp6 := layers.ICMPv6{
		TypeCode: layers.ICMPv6TypeRouterAdvertisement << 8,
	}
	prefixData := []byte{
		prefixLength,
		0x0c,                   // flags
		0x00, 0x27, 0x8d, 0x00, // valid lifetime (2592000)
		0x00, 0x09, 0x3a, 0x80, // preferred lifetime (604800)
		0x00, 0x00, 0x00, 0x00, // reserved
	}
	prefixData = append(prefixData, []byte(net.ParseIP(prefix))...)

	adv := layers.ICMPv6RouterAdvertisement{
		HopLimit:       255,
		Flags:          0x08, // prf
		RouterLifetime: 1800,
		Options: []layers.ICMPv6Option{
			{
				Type: layers.ICMPv6OptSourceAddress,
				Data: hw,
			},
			{
				Type: layers.ICMPv6OptMTU,
				Data: []byte{0x00, 0x00, 0x00, 0x00, 0x05, 0xdc}, // 1500
			},
			{
				Type: layers.ICMPv6OptPrefixInfo,
				Data: prefixData,
			},
		},
	}
	icmp6.SetNetworkLayerForChecksum(&ip6)

	return Serialize(&eth, &ip6, &icmp6, &adv)
}
