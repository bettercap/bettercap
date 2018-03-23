package dhcp6server

import (
	"github.com/mdlayher/dhcp6"
)

// Handler provides an interface which allows structs to act as DHCPv6 server
// handlers.  ServeDHCP implementations receive a copy of the incoming DHCP
// request via the Request parameter, and allow outgoing communication via
// the ResponseSender.
//
// ServeDHCP implementations can choose to write a response packet using the
// ResponseSender interface, or choose to not write anything at all.  If no packet
// is sent back to the client, it may choose to back off and retry, or attempt
// to pursue communication with other DHCP servers.
type Handler interface {
	ServeDHCP(ResponseSender, *Request)
}

// HandlerFunc is an adapter type which allows the use of normal functions as
// DHCP handlers.  If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler struct that calls f.
type HandlerFunc func(ResponseSender, *Request)

// ServeDHCP calls f(w, r), allowing regular functions to implement Handler.
func (f HandlerFunc) ServeDHCP(w ResponseSender, r *Request) {
	f(w, r)
}

// ResponseSender provides an interface which allows a DHCP handler to construct
// and send a DHCP response packet.  In addition, the server automatically handles
// copying certain options from a client Request to a ResponseSender's Options,
// including:
//   - Client ID (OptionClientID)
//   - Server ID (OptionServerID)
//
// ResponseSender implementations should use the same transaction ID sent in a
// client Request.
type ResponseSender interface {
	// Options returns the Options map that will be sent to a client
	// after a call to Send.
	Options() dhcp6.Options

	// Send generates a DHCP response packet using the input message type
	// and any options set by Options.  Send returns the number of bytes
	// sent and any errors which occurred.
	Send(dhcp6.MessageType) (int, error)
}
