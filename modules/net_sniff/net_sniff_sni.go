package net_sniff

import (
	"fmt"

	"regexp"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

// poor man's TLS Client Hello with SNI extension parser :P
var sniRe = regexp.MustCompile("\x00\x00.{4}\x00.{2}([a-z0-9A-Z]+([\\-.]{1}[a-z0-9A-Z]+)*\\.[a-zA-Z]{2,6})\x00")

func sniParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	dataSize := len(data)

	if dataSize < 2 || data[0] != 0x16 || data[1] != 0x03 {
		return false
	}

	m := sniRe.FindSubmatch(data)
	if len(m) < 2 {
		return false
	}

	domain := string(m[1])
	if tcp.DstPort != 443 {
		domain = fmt.Sprintf("%s:%d", domain, tcp.DstPort)
	}

	NewSnifferEvent(
		pkt.Metadata().Timestamp,
		"https",
		ip.SrcIP.String(),
		domain,
		nil,
		"%s %s > %s",
		tui.Wrap(tui.BACKYELLOW+tui.FOREWHITE, "sni"),
		vIP(ip.SrcIP),
		tui.Yellow("https://"+domain),
	).Push()

	return true
}
