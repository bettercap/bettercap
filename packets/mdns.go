package packets

import (
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/miekg/dns"
)

const MDNSPort = 5353

func MDNSGetMeta(pkt gopacket.Packet) map[string]string {
	if ludp := pkt.Layer(layers.LayerTypeUDP); ludp != nil {
		if udp := ludp.(*layers.UDP); udp != nil && udp.SrcPort == MDNSPort && udp.DstPort == MDNSPort {
			var msg dns.Msg
			if err := msg.Unpack(udp.Payload); err == nil && msg.Opcode == dns.OpcodeQuery && len(msg.Answer) > 0 {
				for _, answer := range append(msg.Answer, msg.Extra...) {
					switch answer.(type) {
					case *dns.TXT:
						meta := make(map[string]string)
						txt := answer.(*dns.TXT)
						for _, value := range txt.Txt {
							if strings.Contains(value, "=") {
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
			var msg dns.Msg
			if err := msg.Unpack(udp.Payload); err == nil && msg.Opcode == dns.OpcodeQuery && len(msg.Answer) > 0 {
				for _, answer := range append(msg.Answer, msg.Extra...) {
					switch rr := answer.(type) {
					case *dns.PTR:
					case *dns.A:
					case *dns.AAAA:
						return rr.Header().Name
					}
				}
			}
		}
	}
	return ""
}
