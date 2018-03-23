// Package dhcp6test provides utilities for testing DHCPv6 clients and servers.
package dhcp6test

import (
	"github.com/mdlayher/dhcp6"
)

// Recorder is a dhcp6.ResponseSender which captures a response's message type and
// options, for inspection during tests.
type Recorder struct {
	MessageType   dhcp6.MessageType
	TransactionID [3]byte
	OptionsMap    dhcp6.Options
	Packet        *dhcp6.Packet
}

// NewRecorder creates a new Recorder which uses the input transaction ID.
func NewRecorder(txID [3]byte) *Recorder {
	return &Recorder{
		TransactionID: txID,
		OptionsMap:    make(dhcp6.Options),
	}
}

// Options returns the Options map of a Recorder.
func (r *Recorder) Options() dhcp6.Options {
	return r.OptionsMap
}

// Send creates a new DHCPv6 packet using the input message type, and stores
// it for later inspection.
func (r *Recorder) Send(mt dhcp6.MessageType) (int, error) {
	r.MessageType = mt
	p := &dhcp6.Packet{
		MessageType:   mt,
		TransactionID: r.TransactionID,
		Options:       r.OptionsMap,
	}
	r.Packet = p

	b, err := p.MarshalBinary()
	return len(b), err
}
