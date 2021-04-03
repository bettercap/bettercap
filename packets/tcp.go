package packets

import (
	"github.com/google/gopacket/layers"
	"net"
)

func NewTCPSyn(from net.IP, from_hw net.HardwareAddr, to net.IP, to_hw net.HardwareAddr, srcPort int, dstPort int) (error, []byte) {
	from4 := from.To4()
	to4 := to.To4()

	if from4 != nil && to4 != nil {
		eth := layers.Ethernet{
			SrcMAC:       from_hw,
			DstMAC:       to_hw,
			EthernetType: layers.EthernetTypeIPv4,
		}
		ip4 := layers.IPv4{
			Protocol: layers.IPProtocolTCP,
			Version:  4,
			TTL:      64,
			SrcIP:    from,
			DstIP:    to,
		}
		tcp := layers.TCP{
			SrcPort: layers.TCPPort(srcPort),
			DstPort: layers.TCPPort(dstPort),
			SYN:     true,
		}
		tcp.SetNetworkLayerForChecksum(&ip4)

		return Serialize(&eth, &ip4, &tcp)
	} else {
		eth := layers.Ethernet{
			SrcMAC:       from_hw,
			DstMAC:       to_hw,
			EthernetType: layers.EthernetTypeIPv6,
		}
		ip6 := layers.IPv6{
			Version: 6,
			NextHeader: layers.IPProtocolTCP,
			HopLimit:   64,
			SrcIP:   from,
			DstIP:   to,
		}
		tcp := layers.TCP{
			SrcPort: layers.TCPPort(srcPort),
			DstPort: layers.TCPPort(dstPort),
			SYN:     true,
		}
		tcp.SetNetworkLayerForChecksum(&ip6)

		return Serialize(&eth, &ip6, &tcp)
	}
}
