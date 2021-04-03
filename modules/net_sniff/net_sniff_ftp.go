package net_sniff

import (
	"net"
	"regexp"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

var (
	ftpRe = regexp.MustCompile(`^(USER|PASS) (.+)[\n\r]+$`)
)

func ftpParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := string(tcp.Payload)

	if matches := ftpRe.FindAllStringSubmatch(data, -1); matches != nil {
		what := str.Trim(matches[0][1])
		cred := str.Trim(matches[0][2])
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"ftp",
			srcIP.String(),
			dstIP.String(),
			nil,
			"%s %s > %s:%s - %s %s",
			tui.Wrap(tui.BACKYELLOW+tui.FOREWHITE, "ftp"),
			vIP(srcIP),
			vIP(dstIP),
			vPort(tcp.DstPort),
			tui.Bold(what),
			tui.Yellow(cred),
		).Push()

		return true
	}

	return false
}
