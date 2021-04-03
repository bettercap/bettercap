package syn_scan

import (
	"sync/atomic"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/async"
)

type OpenPort struct {
	Proto   string `json:"proto"`
	Banner  string `json:"banner"`
	Service string `json:"service"`
	Port    int    `json:"port"`
}

func (mod *SynScanner) onPacket(pkt gopacket.Packet) {
	if pkt == nil || pkt.Data() == nil {
		return
	}

	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP



	isIPv6 := false
	foundLayerTypes := []gopacket.LayerType{}
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth,
		&ip4,
		&tcp,
	)

	err := parser.DecodeLayers(pkt.Data(), &foundLayerTypes)
	if err != nil {
		// try ipv6
		parser := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&eth,
			&ip6,
			&tcp,
		)
		err = parser.DecodeLayers(pkt.Data(), &foundLayerTypes)
		if err != nil {
			return
		}
		isIPv6 = true
	}

	if tcp.DstPort == synSourcePort && tcp.SYN && tcp.ACK {
		atomic.AddUint64(&mod.stats.openPorts, 1)

		port := int(tcp.SrcPort)

		openPort := &OpenPort{
			Proto:   "tcp",
			Port:    port,
			Service: network.GetServiceByPort(port, "tcp"),
		}

		var host *network.Endpoint

		from := ""

		if isIPv6 {
			from = ip6.SrcIP.String()
			if ip6.SrcIP.Equal(mod.Session.Interface.IPv6) {
				host = mod.Session.Interface
			} else if ip6.SrcIP.Equal(mod.Session.Gateway.IPv6) {
				host = mod.Session.Gateway
			} else {
				host = mod.Session.Lan.GetByIp(from)
			}
		} else {
			from = ip4.SrcIP.String()

			if ip4.SrcIP.Equal(mod.Session.Interface.IP) {
				host = mod.Session.Interface
			} else if ip4.SrcIP.Equal(mod.Session.Gateway.IP) {
				host = mod.Session.Gateway
			} else {
				host = mod.Session.Lan.GetByIp(from)
			}
		}

		if host != nil {
			ports := host.Meta.GetOr("ports", map[int]*OpenPort{}).(map[int]*OpenPort)
			if _, found := ports[port]; !found {
				ports[port] = openPort
			}
			host.Meta.Set("ports", ports)
		}

		mod.bannerQueue.Add(async.Job(grabberJob{from, openPort}))

		NewSynScanEvent(from, host, port).Push()
	}
}
