package modules

import (
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func dnsParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if parsed == false {
		return false
	}

	if dns.OpCode != layers.DNSOpCodeQuery {
		return false
	}

	m := make(map[string][]string, 0)
	for _, a := range dns.Answers {
		if a.IP == nil {
			continue
		}

		hostname := string(a.Name)
		if _, found := m[hostname]; found == false {
			m[hostname] = make([]string, 0)
		}

		m[hostname] = append(m[hostname], vIP(a.IP))
	}

	for hostname, ips := range m {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"dns",
			ip.SrcIP.String(),
			ip.DstIP.String(),
			nil,
			"%s %s > %s : %s is %s",
			core.W(core.BG_DGRAY+core.FG_WHITE, "dns"),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			core.Yellow(hostname),
			core.Dim(strings.Join(ips, ", ")),
		).Push()
	}

	return true
}
