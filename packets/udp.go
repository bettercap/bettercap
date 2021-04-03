package packets

import (
	"github.com/google/gopacket/layers"
	"net"
)

func NewUDPProbe(from net.IP, from_hw net.HardwareAddr, to net.IP, port int) (error, []byte) {
	eth := layers.Ethernet{
		SrcMAC:       from_hw,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeIPv4,
	}

	udp := layers.UDP{
		SrcPort: layers.UDPPort(12345),
		DstPort: layers.UDPPort(port),
	}
	udp.Payload = []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}

	if to.To4() == nil {
		ip6 := layers.IPv6{
			NextHeader: layers.IPProtocolUDP,
			Version:    6,
			SrcIP:      from,
			DstIP:      to,
			HopLimit:   64,
		}

		udp.SetNetworkLayerForChecksum(&ip6)

		return Serialize(&eth, &ip6, &udp)
	} else {
		ip4 := layers.IPv4{
			Protocol: layers.IPProtocolUDP,
			Version:  4,
			TTL:      64,
			SrcIP:    from,
			DstIP:    to,
		}

		udp.SetNetworkLayerForChecksum(&ip4)

		return Serialize(&eth, &ip4, &udp)
	}
}
