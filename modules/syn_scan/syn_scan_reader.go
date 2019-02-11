package syn_scan

import (
	"net"
	"sync/atomic"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (s *SynScanner) isAddressInRange(ip net.IP) bool {
	for _, a := range s.addresses {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}

func (s *SynScanner) onPacket(pkt gopacket.Packet) {
	var eth layers.Ethernet
	var ip layers.IPv4
	var tcp layers.TCP
	foundLayerTypes := []gopacket.LayerType{}

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth,
		&ip,
		&tcp,
	)

	err := parser.DecodeLayers(pkt.Data(), &foundLayerTypes)
	if err != nil {
		return
	}

	if s.isAddressInRange(ip.SrcIP) && tcp.DstPort == synSourcePort && tcp.SYN && tcp.ACK {
		atomic.AddUint64(&s.stats.openPorts, 1)

		from := ip.SrcIP.String()
		port := int(tcp.SrcPort)

		var host *network.Endpoint
		if ip.SrcIP.Equal(s.Session.Interface.IP) {
			host = s.Session.Interface
		} else if ip.SrcIP.Equal(s.Session.Gateway.IP) {
			host = s.Session.Gateway
		} else {
			host = s.Session.Lan.GetByIp(from)
		}

		if host != nil {
			ports := host.Meta.GetIntsWith("tcp-ports", port, true)
			host.Meta.SetInts("tcp-ports", ports)
		}

		NewSynScanEvent(from, host, port).Push()
	}
}
