package modules

import (
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func mdnsParser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
	if udp.SrcPort == packets.MDNSPort && udp.DstPort == packets.MDNSPort {
		dns := layers.DNS{}
		if err := dns.DecodeFromBytes(udp.Payload, gopacket.NilDecodeFeedback); err == nil && dns.OpCode == layers.DNSOpCodeQuery {
			m := make(map[string][]string)
			answers := append(dns.Answers, dns.Additionals...)
			answers = append(answers, dns.Authorities...)
			for _, answer := range answers {
				if answer.Type == layers.DNSTypeA || answer.Type == layers.DNSTypeAAAA {
					hostname := string(answer.Name)
					if _, found := m[hostname]; found == false {
						m[hostname] = make([]string, 0)
					}
					m[hostname] = append(m[hostname], answer.IP.String())
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
