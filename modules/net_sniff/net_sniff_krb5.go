package net_sniff

import (
	"encoding/asn1"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

func krb5Parser(ip *layers.IPv4, pkt gopacket.Packet, udp *layers.UDP) bool {
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
			ip.SrcIP.String(),
			ip.DstIP.String(),
			nil,
			"%s %s -> %s : %s",
			tui.Wrap(tui.BACKRED+tui.FOREBLACK, "krb-as-req"),
			vIP(ip.SrcIP),
			vIP(ip.DstIP),
			s,
		).Push()

		return true
	}

	return false
}
