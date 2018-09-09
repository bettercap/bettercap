package packets

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	UPNPPort = 1900
)

var (
	UPNPDestMac = net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0xfb}
	UPNPDestIP  = net.ParseIP("239.255.255.250")
)

func NewUPNPProbe(from net.IP, from_hw net.HardwareAddr) (error, []byte) {
	eth := layers.Ethernet{
		SrcMAC:       from_hw,
		DstMAC:       UPNPDestMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    from,
		DstIP:    UPNPDestIP,
	}

	udp := layers.UDP{
		SrcPort: layers.UDPPort(12345),
		DstPort: layers.UDPPort(UPNPPort),
	}

	payload := []byte("M-SEARCH * HTTP/1.1\r\n" +
		fmt.Sprintf("Host: %s:%d\r\n", UPNPDestIP, UPNPPort) +
		"Man: ssdp:discover\r\n" +
		"ST: ssdp:all\r\n" +
		"MX: 2\r\n" +
		"\r\n")

	if err := udp.SetNetworkLayerForChecksum(&ip4); err != nil {
		return err, nil
	}

	return Serialize(&eth, &ip4, &udp, gopacket.Payload(payload))
}
