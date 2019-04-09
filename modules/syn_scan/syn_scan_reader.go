package syn_scan

import (
	"net"
	"sync/atomic"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type OpenPort struct {
	Proto   string `json:"proto"`
	Banner  string `json:"banner"`
	Service string `json:"service"`
	Port    int    `json:"port"`
}

func (mod *SynScanner) isAddressInRange(ip net.IP) bool {
	for _, a := range mod.addresses {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}

func (mod *SynScanner) onPacket(pkt gopacket.Packet) {
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

	if mod.isAddressInRange(ip.SrcIP) && tcp.DstPort == synSourcePort && tcp.SYN && tcp.ACK {
		atomic.AddUint64(&mod.stats.openPorts, 1)

		from := ip.SrcIP.String()
		port := int(tcp.SrcPort)

		var host *network.Endpoint
		if ip.SrcIP.Equal(mod.Session.Interface.IP) {
			host = mod.Session.Interface
		} else if ip.SrcIP.Equal(mod.Session.Gateway.IP) {
			host = mod.Session.Gateway
		} else {
			host = mod.Session.Lan.GetByIp(from)
		}

		if host != nil {
			ports := host.Meta.GetOr("ports", map[int]OpenPort{}).(map[int]OpenPort)
			if _, found := ports[port]; !found {
				ports[port] = OpenPort{
					Proto:   "tcp",
					Port:    port,
					Service: network.GetServiceByPort(port, "tcp"),
				}
			}

			host.Meta.Set("ports", ports)
		}

		NewSynScanEvent(from, host, port).Push()
	}
}
