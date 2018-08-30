package modules

import (
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/miekg/dns"
)

func mdnsCollectHostname(m map[string][]string, hostname string, address string) {
	if _, found := m[hostname]; found == false {
		m[hostname] = make([]string, 0)
	}
	m[hostname] = append(m[hostname], address)
}

func mdnsParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	if udp.SrcPort == packets.MDNSPort && udp.DstPort == packets.MDNSPort {
		var msg dns.Msg
		if err := msg.Unpack(udp.Payload); err == nil && msg.Opcode == dns.OpcodeQuery && len(msg.Answer) > 0 {
			m := make(map[string][]string)
			for _, answer := range append(msg.Answer, msg.Extra...) {
				switch rr := answer.(type) {
				case *dns.A:
					mdnsCollectHostname(m, rr.Header().Name, answer.(*dns.A).A.String())

				case *dns.AAAA:
					mdnsCollectHostname(m, rr.Header().Name, answer.(*dns.AAAA).AAAA.String())
				}
			}

			for hostname, ips := range m {
				NewSnifferEvent(
					pkt.Metadata().Timestamp,
					"mdns",
					ip.SrcIP.String(),
					ip.DstIP.String(),
					nil,
					"%s %s : %s is %s",
					core.W(core.BG_DGRAY+core.FG_WHITE, "mdns"),
					vIP(ip.SrcIP),
					core.Yellow(hostname),
					core.Dim(strings.Join(ips, ", ")),
				).Push()
			}

			return true
		}
	}
	return false
}
