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

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    from,
		DstIP:    to,
	}

	udp := layers.UDP{
		SrcPort: layers.UDPPort(12345),
		DstPort: layers.UDPPort(port),
	}
	udp.Payload = []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}

	udp.SetNetworkLayerForChecksum(&ip4)

	return Serialize(&eth, &ip4, &udp)
}
