package net_sniff

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func onDOT11(radiotap *layers.RadioTap, dot11 *layers.Dot11, pkt gopacket.Packet, verbose bool) {
	NewSnifferEvent(
		pkt.Metadata().Timestamp,
		"802.11",
		"-",
		"-",
		len(pkt.Data()),
		"%s %s proto=%d a1=%s a2=%s a3=%s a4=%s seqn=%d frag=%d",
		dot11.Type,
		dot11.Flags,
		dot11.Proto,
		dot11.Address1,
		dot11.Address2,
		dot11.Address3,
		dot11.Address4,
		dot11.SequenceNumber,
		dot11.FragmentNumber,
	).Push()
}
