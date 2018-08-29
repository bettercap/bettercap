package packets

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/miekg/dns"
)

const MDNSPort = 5353

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
