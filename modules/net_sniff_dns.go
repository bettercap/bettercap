package modules

import (
	"fmt"

	"github.com/evilsocket/bettercap-ng/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func dnsParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if parsed == false {
		return false
	}

	if dns.OpCode != layers.DNSOpCodeQuery || len(dns.Answers) == 0 {
		return false
	}

	for _, a := range dns.Answers {
		if a.IP == nil {
			continue
		}
		fmt.Printf("[%s] %s %s > %s : %s is %s\n",
			vTime(pkt.Metadata().Timestamp),
			core.W(core.BG_DGRAY+core.FG_WHITE, "dns"),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			core.Yellow(string(a.Name)),
			vIP(a.IP))
	}

	return true
}
