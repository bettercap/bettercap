package net_sniff

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"strings"

	"github.com/evilsocket/islazy/tui"
)

func dnsParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, udp *layers.UDP) bool {
	dns, parsed := pkt.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !parsed {
		return false
	}

	if dns.OpCode != layers.DNSOpCodeQuery {
		return false
	}

	m := make(map[string][]string)
	answers := [][]layers.DNSResourceRecord{
		dns.Answers,
		dns.Authorities,
		dns.Additionals,
	}

	for _, list := range answers {
		for _, a := range list {
			if a.IP == nil {
				continue
			}

			hostname := string(a.Name)
			if _, found := m[hostname]; !found {
				m[hostname] = make([]string, 0)
			}

			m[hostname] = append(m[hostname], vIP(a.IP))
		}
	}

	if len(m) == 0 && dns.ResponseCode != layers.DNSResponseCodeNoErr {
		for _, a := range dns.Questions {
			m[string(a.Name)] = []string{tui.Red(dns.ResponseCode.String())}
		}
	}

	for hostname, ips := range m {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"dns",
			srcIP.String(),
			dstIP.String(),
			nil,
			"%s %s > %s : %s is %s",
			tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "dns"),
			vIP(srcIP),
			vIP(dstIP),
			tui.Yellow(hostname),
			tui.Dim(strings.Join(ips, ", ")),
		).Push()
	}

	return true
}
