package net_sniff

import (
	"net"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var (
	ntlmRe  = regexp.MustCompile("(WWW-|Proxy-|)(Authenticate|Authorization): (NTLM|Negotiate)")
	challRe = regexp.MustCompile("(WWW-|Proxy-|)(Authenticate): (NTLM|Negotiate)")
	respRe  = regexp.MustCompile("(WWW-|Proxy-|)(Authorization): (NTLM|Negotiate)")
	ntlm    = packets.NewNTLMState()
)

func isNtlm(s string) bool {
	return ntlmRe.FindString(s) != ""
}

func isChallenge(s string) bool {
	return challRe.FindString(s) != ""
}

func isResponse(s string) bool {
	return respRe.FindString(s) != ""
}

func ntlmParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	ok := false

	for _, line := range strings.Split(string(data), "\r\n") {
		if isNtlm(line) {
			tokens := strings.Split(line, " ")
			if len(tokens) != 3 {
				continue
			}
			if isChallenge(line) {
				ok = true
				ntlm.AddServerResponse(tcp.Ack, tokens[2])
			} else if isResponse(line) {
				ok = true
				ntlm.AddClientResponse(tcp.Seq, tokens[2], func(data packets.NTLMChallengeResponseParsed) {
					NewSnifferEvent(
						pkt.Metadata().Timestamp,
						"ntlm.response",
						srcIP.String(),
						dstIP.String(),
						nil,
						"%s %s > %s | %s",
						tui.Wrap(tui.BACKDARKGRAY+tui.FOREWHITE, "ntlm.response"),
						vIP(srcIP),
						vIP(dstIP),
						data.LcString(),
					).Push()
				})
			}
		}
	}
	return ok
}
