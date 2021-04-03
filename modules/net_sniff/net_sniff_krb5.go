package net_sniff

import (
	"encoding/asn1"
	"net"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func krb5Parser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, udp *layers.UDP) bool {
	if udp.DstPort != 88 {
		return false
	}

	var req packets.Krb5Request
	_, err := asn1.UnmarshalWithParams(udp.Payload, &req, packets.Krb5AsReqParam)
	if err != nil {
		return false
	}

	if s, err := req.String(); err == nil {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"krb5",
			srcIP.String(),
			dstIP.String(),
			nil,
			"%s %s -> %s : %s",
			tui.Wrap(tui.BACKRED+tui.FOREBLACK, "krb-as-req"),
			vIP(srcIP),
			vIP(dstIP),
			s,
		).Push()

		return true
	}

	return false
}
