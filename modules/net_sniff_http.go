package modules

import (
	"fmt"
	"regexp"

	"github.com/evilsocket/bettercap-ng/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var httpRe = regexp.MustCompile("(?s).*(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH) (.+) HTTP/\\d\\.\\d.+Host: ([^\\s]+)")
var uaRe = regexp.MustCompile("(?s).*User-Agent: ([^\\n]+).+")

func httpParser(
	ip *layers.IPv4,
	pkt gopacket.Packet,
	tcp *layers.TCP,
	truncateURLs bool,
) bool {
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

	url := fmt.Sprintf("%s", core.Yellow(hostname))
	if tcp.DstPort != 80 {
		url += fmt.Sprintf(":%s", vPort(tcp.DstPort))
	}
	url += fmt.Sprintf("%s", path)

	// shorten / truncate long URLs if needed
	formattedURL := string(url)
	if truncateURLs {
		formattedURL = vURL(url)
	}

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
		},
		"[%s] %s %s %s %s %s",
		vTime(pkt.Metadata().Timestamp),
		core.W(core.BG_RED+core.FG_BLACK, "http"),
		vIP(ip.SrcIP),
		core.W(core.BG_LBLUE+core.FG_BLACK, method),
		formattedURL,
		core.Dim(ua),
	).Push()

	return true
}
