package net_sniff

import (
	"net"
	"strings"

	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func mdnsParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, udp *layers.UDP) bool {
	if udp.SrcPort == packets.MDNSPort && udp.DstPort == packets.MDNSPort {
		dns := layers.DNS{}
		if err := dns.DecodeFromBytes(udp.Payload, gopacket.NilDecodeFeedback); err == nil && dns.OpCode == layers.DNSOpCodeQuery {
			for _, q := range dns.Questions {
				NewSnifferEvent(
					pkt.Metadata().Timestamp,
					"mdns",
					srcIP.String(),
					dstIP.String(),
					nil,
					"%s %s : %s query for %s",
					tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "mdns"),
					vIP(srcIP),
					tui.Dim(q.Type.String()),
					tui.Yellow(string(q.Name)),
				).Push()
			}

			m := make(map[string][]string)
			answers := append(dns.Answers, dns.Additionals...)
			answers = append(answers, dns.Authorities...)
			for _, answer := range answers {
				if answer.Type == layers.DNSTypeA || answer.Type == layers.DNSTypeAAAA {
					hostname := string(answer.Name)
					if _, found := m[hostname]; !found {
						m[hostname] = make([]string, 0)
					}
					m[hostname] = append(m[hostname], answer.IP.String())
				}
			}

			for hostname, ips := range m {
				for _, ip := range ips {
					if endpoint := session.I.Lan.GetByIp(ip); endpoint != nil {
						endpoint.OnMeta(map[string]string{
							"mdns:hostname": hostname,
						})
					}
				}

				NewSnifferEvent(
					pkt.Metadata().Timestamp,
					"mdns",
					srcIP.String(),
					dstIP.String(),
					nil,
					"%s %s : %s is %s",
					tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "mdns"),
					vIP(srcIP),
					tui.Yellow(hostname),
					tui.Dim(strings.Join(ips, ", ")),
				).Push()
			}

			return true
		}
	}
	return false
}
