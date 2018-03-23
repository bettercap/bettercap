package dhcp6server

import (
	"net"
	"reflect"
	"testing"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/dhcp6opts"
)

// TestParseRequest verifies that ParseRequest returns a consistent
// Request struct for use in Handler types.
func TestParseRequest(t *testing.T) {
	p := &dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeSolicit,
		TransactionID: [3]byte{1, 2, 3},
		Options:       make(dhcp6.Options),
	}
	var uuid [16]byte
	p.Options.Add(dhcp6.OptionClientID, dhcp6opts.NewDUIDUUID(uuid))

	addr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: 546,
	}

	buf, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	r := &Request{
		MessageType:   p.MessageType,
		TransactionID: p.TransactionID,
		Options:       make(dhcp6.Options),
		Length:        int64(len(buf)),
		RemoteAddr:    "[::1]:546",
	}
	r.Options.Add(dhcp6.OptionClientID, dhcp6opts.NewDUIDUUID(uuid))

	gotR, err := ParseRequest(buf, addr)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := r, gotR; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected Request for ParseRequest(%v, %v)\n- want: %v\n-  got: %v",
			p, addr, want, got)
	}
}
