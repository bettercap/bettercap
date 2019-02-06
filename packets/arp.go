package packets

import (
	"net"

	"github.com/google/gopacket/layers"
)

func NewARPTo(from net.IP, from_hw net.HardwareAddr, to net.IP, to_hw net.HardwareAddr, req uint16) (layers.Ethernet, layers.ARP) {
	eth := layers.Ethernet{
		SrcMAC:       from_hw,
		DstMAC:       to_hw,
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         req,
		SourceHwAddress:   from_hw,
		SourceProtAddress: from.To4(),
		DstHwAddress:      to_hw,
		DstProtAddress:    to.To4(),
	}

	return eth, arp
}

func NewARP(from net.IP, from_hw net.HardwareAddr, to net.IP, req uint16) (layers.Ethernet, layers.ARP) {
	return NewARPTo(from, from_hw, to, []byte{0, 0, 0, 0, 0, 0}, req)
}

func NewARPRequest(from net.IP, from_hw net.HardwareAddr, to net.IP) (error, []byte) {
	eth, arp := NewARP(from, from_hw, to, layers.ARPRequest)
	return Serialize(&eth, &arp)
}

func NewARPReply(from net.IP, from_hw net.HardwareAddr, to net.IP, to_hw net.HardwareAddr) (error, []byte) {
	eth, arp := NewARPTo(from, from_hw, to, to_hw, layers.ARPReply)
	return Serialize(&eth, &arp)
}
