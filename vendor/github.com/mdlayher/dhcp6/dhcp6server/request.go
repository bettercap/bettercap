package dhcp6server

import (
	"net"

	"github.com/mdlayher/dhcp6"
)

// Request represents a processed DHCP request received by a server.
// Its struct members contain information regarding the request's message
// type, transaction ID, client ID, options, etc.
type Request struct {
	// DHCP message type, such as Solicit, Request, or Renew.
	MessageType dhcp6.MessageType

	// Unique transaction ID, which should be preserved across
	// multiple requests to the same DHCP server.  ServeDHCP
	// implementations must manually verify that the same
	// transaction ID is used.
	TransactionID [3]byte

	// Map of options sent by client, carrying additional
	// information or requesting additional information from
	// the server.  Its methods can be used to check for and parse
	// additional information relating to a request.
	Options dhcp6.Options

	// Length of the DHCP request, in bytes.
	Length int64

	// Network address which was used to contact the DHCP server.
	RemoteAddr string
}

// ParseRequest creates a new Request from an input byte slice and UDP address.
// It populates the basic struct members which can be used in a DHCP handler.
//
// If the input byte slice is not a valid DHCP packet, ErrInvalidPacket is
// returned.
func ParseRequest(b []byte, remoteAddr *net.UDPAddr) (*Request, error) {
	p := new(dhcp6.Packet)
	if err := p.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return &Request{
		MessageType:   p.MessageType,
		TransactionID: p.TransactionID,
		Options:       p.Options,
		Length:        int64(len(b)),
		RemoteAddr:    remoteAddr.String(),
	}, nil
}
