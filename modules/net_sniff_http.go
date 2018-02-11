package modules

import (
	"fmt"

	"github.com/evilsocket/bettercap-ng/core"
	"regexp"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var httpRe = regexp.MustCompile("(?s).*(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH) (.+) HTTP/\\d\\.\\d.+Host: ([^\\s]+)")
var uaRe = regexp.MustCompile("(?s).*User-Agent: ([^\\n]+).+")
var authRe = regexp.MustCompile("(?s).*Authorization: ([^\\n]+).+")

func httpParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	dataSize := len(data)

	if dataSize < 20 {
		return false
	}

	m := httpRe.FindSubmatch(data)
	if len(m) != 4 {
		return false
	}

	method := string(m[1])
	hostname := string(m[3])
	path := string(m[2])
	ua := ""
	mu := uaRe.FindSubmatch(data)
	if len(mu) == 2 {
		ua = string(mu[1])
	}
	auth := ""
	authDesc := ""
	mauth := authRe.FindSubmatch(data)
	if len(mauth) == 2 {
		auth = string(mauth[1])
		authDesc = fmt.Sprintf(" auth=%s", core.Red(auth))
	}

	url := fmt.Sprintf("%s", core.Yellow(hostname))
	if tcp.DstPort != 80 {
		url += fmt.Sprintf(":%s", vPort(tcp.DstPort))
	}
	url += fmt.Sprintf("%s", path)

	NewSnifferEvent(
		pkt.Metadata().Timestamp,
		"http",
		ip.SrcIP.String(),
		hostname,
		SniffData{
			"method": method,
			"host":   hostname,
			"path":   url,
			"agent":  ua,
			"auth":   auth,
		},
		"%s %s %s %s %s%s",
		core.W(core.BG_RED+core.FG_BLACK, "http"),
		vIP(ip.SrcIP),
		core.W(core.BG_LBLUE+core.FG_BLACK, method),
		vURL(url),
		core.Dim(ua),
		authDesc,
	).Push()

	return true
}
