// Package dhcp6 implements a DHCPv6 server, as described in RFC 3315.
//
// Unless otherwise stated, any reference to "DHCP" in this package refers to
// DHCPv6 only.
package dhcp6

import (
	"errors"
)

//go:generate stringer -output=string.go -type=MessageType,Status,OptionCode

var (
	// ErrInvalidOptions is returned when invalid options data is encountered
	// during parsing.  The data could report an incorrect length or have
	// trailing bytes which are not part of the option.
	ErrInvalidOptions = errors.New("invalid options data")

	// ErrInvalidPacket is returned when a byte slice does not contain enough
	// data to create a valid Packet.  A Packet must have at least a message type
	// and transaction ID.
	ErrInvalidPacket = errors.New("not enough bytes for valid packet")

	// ErrOptionNotPresent is returned when a requested opcode is not in
	// the packet.
	ErrOptionNotPresent = errors.New("option code not present in packet")
)
