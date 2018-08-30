package packets

import (
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const MDNSPort = 5353

func MDNSGetMeta(pkt gopacket.Packet) map[string]string {
	if ludp := pkt.Layer(layers.LayerTypeUDP); ludp != nil {
		if udp := ludp.(*layers.UDP); udp != nil && udp.SrcPort == MDNSPort && udp.DstPort == MDNSPort {
			dns := layers.DNS{}
			if err := dns.DecodeFromBytes(udp.Payload, gopacket.NilDecodeFeedback); err == nil {
				answers := append(dns.Answers, dns.Additionals...)
				answers = append(answers, dns.Authorities...)
				for _, answer := range answers {
					switch answer.Type {
					case layers.DNSTypeTXT:
						meta := make(map[string]string)
						for _, raw := range answer.TXTs {
							if value := string(raw); strings.Contains(value, "=") {
								parts := strings.SplitN(value, "=", 2)
								meta[core.Trim(parts[0])] = core.Trim(parts[1])
							}
						}
						if len(meta) > 0 {
							return meta
						}
					}
				}
			}
		}
	}
	return nil
}

func MDNSGetHostname(pkt gopacket.Packet) string {
	if ludp := pkt.Layer(layers.LayerTypeUDP); ludp != nil {
		if udp := ludp.(*layers.UDP); udp != nil && udp.SrcPort == MDNSPort && udp.DstPort == MDNSPort {
			dns := layers.DNS{}
			if err := dns.DecodeFromBytes(udp.Payload, gopacket.NilDecodeFeedback); err == nil {
				answers := append(dns.Answers, dns.Additionals...)
				answers = append(answers, dns.Authorities...)
				for _, answer := range answers {
					switch answer.Type {
					case layers.DNSTypePTR:
					case layers.DNSTypeA:
					case layers.DNSTypeAAAA:
						return string(answer.Name)
					}
				}
			}
		}
	}
	return ""
}
