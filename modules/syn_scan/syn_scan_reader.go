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

	if tcp.DstPort == synSourcePort && tcp.SYN && tcp.ACK {
		atomic.AddUint64(&mod.stats.openPorts, 1)

		from := ip.SrcIP.String()
		port := int(tcp.SrcPort)

		openPort := &OpenPort{
			Proto:   "tcp",
			Port:    port,
			Service: network.GetServiceByPort(port, "tcp"),
		}

		var host *network.Endpoint
		if ip.SrcIP.Equal(mod.Session.Interface.IP) {
			host = mod.Session.Interface
		} else if ip.SrcIP.Equal(mod.Session.Gateway.IP) {
			host = mod.Session.Gateway
		} else {
			host = mod.Session.Lan.GetByIp(from)
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
