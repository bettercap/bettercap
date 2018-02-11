package modules

import (
	"regexp"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	ntlmRe  = regexp.MustCompile("(WWW-|Proxy-|)(Authenticate|Authorization): (NTLM|Negotiate)")
	challRe = regexp.MustCompile("(WWW-|Proxy-|)(Authenticate): (NTLM|Negotiate)")
	respRe  = regexp.MustCompile("(WWW-|Proxy-|)(Authorization): (NTLM|Negotiate)")
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

func ntlmParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	for _, line := range strings.Split(string(data), "\r\n") {
		if isNtlm(line) {
			tokens := strings.Split(line, " ")
			if len(tokens) != 3 {
				continue
			}
			what := "?"
			if isChallenge(line) {
				what = "challenge"
			} else if isResponse(line) {
				what = "response"
			}

			NewSnifferEvent(
				pkt.Metadata().Timestamp,
				"ntlm."+what,
				ip.SrcIP.String(),
				ip.DstIP.String(),
				SniffData{
					what: tokens[2],
				},
				"%s %s > %s | %s",
				core.W(core.BG_DGRAY+core.FG_WHITE, "ntlm."+what),
				vIP(ip.SrcIP),
				vIP(ip.DstIP),
				tokens[2],
			).Push()
		}
	}
	return true
}
