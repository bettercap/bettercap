package modules

import (
	"fmt"
	// "github.com/evilsocket/bettercap-ng/session"
	"github.com/google/gopacket"
)

type SnifferPacketParser func(pkt gopacket.Packet) bool

var PacketParsers = []SnifferPacketParser{}

func noParser(pkt gopacket.Packet) bool {
	fmt.Println(pkt.Dump())
	return true
}
