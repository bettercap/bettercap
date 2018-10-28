package packets

import (
	"errors"
	// TODO: refactor to use gopacket when gopacket folks
	// will fix this > https://github.com/google/gopacket/issues/334
	"github.com/mdlayher/dhcp6"
)

const DHCP6OptDNSServers = 23
const DHCP6OptDNSDomains = 24
const DHCP6OptClientFQDN = 39

// link-local
const IPv6Prefix = "fe80::"

var (
	ErrNoCID = errors.New("Unexpected DHCPv6 packet, could not find client id.")
)

func DHCP6EncodeList(elements []string) (encoded []byte) {
	encoded = make([]byte, 0)

	for _, elem := range elements {
		// this would be worth fuzzing btw
		encoded = append(encoded, byte(len(elem)&0xff))
		encoded = append(encoded, []byte(elem)...)
	}

	return
}

func DHCP6For(what dhcp6.MessageType, to dhcp6.Packet, duid []byte) (err error, p dhcp6.Packet) {
	p = dhcp6.Packet{
		MessageType:   what,
		TransactionID: to.TransactionID,
		Options:       make(dhcp6.Options),
	}

	var rawCID []byte
	if raw, found := to.Options[dhcp6.OptionClientID]; !found || len(raw) < 1 {
		return ErrNoCID, p
	} else {
		rawCID = raw[0]
	}

	p.Options.AddRaw(dhcp6.OptionClientID, rawCID)
	p.Options.AddRaw(dhcp6.OptionServerID, duid)

	return nil, p
}
