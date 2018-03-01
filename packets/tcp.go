package packets

import (
	"github.com/google/gopacket/layers"
	"net"
)

func NewTCPSyn(from net.IP, from_hw net.HardwareAddr, to net.IP, to_hw net.HardwareAddr, srcPort int, dstPort int) (error, []byte) {
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
}
