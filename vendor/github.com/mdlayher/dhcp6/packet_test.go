package dhcp6

import (
	"bytes"
	"reflect"
	"testing"
)

// TestPacketMarshalBinary verifies that Packet.MarshalBinary allocates and returns a correct
// byte slice for a variety of input data.
func TestPacketMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc   string
		packet *Packet
		buf    []byte
	}{
		{
			desc:   "empty packet",
			packet: &Packet{},
			buf:    []byte{0, 0, 0, 0},
		},
		{
			desc: "Solicit only",
			packet: &Packet{
				MessageType: MessageTypeSolicit,
			},
			buf: []byte{1, 0, 0, 0},
		},
		{
			desc: "Solicit, [1 2 3] transaction ID",
			packet: &Packet{
				MessageType:   MessageTypeSolicit,
				TransactionID: [3]byte{1, 2, 3},
			},
			buf: []byte{1, 1, 2, 3},
		},
		{
			desc: "Solicit, [2, 3, 4] transaction ID, option client ID [0 1]",
			packet: &Packet{
				MessageType:   MessageTypeSolicit,
				TransactionID: [3]byte{1, 2, 3},
				Options: Options{
					OptionClientID: [][]byte{{0, 1}},
				},
			},
			buf: []byte{1, 1, 2, 3, 0, 1, 0, 2, 0, 1},
		},
	}

	for i, tt := range tests {
		buf, err := tt.packet.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := tt.buf, buf; !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected packet bytes:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestPacketUnmarshalBinary verifies that Packet.UnmarshalBinary returns
// appropriate Packets and errors for various input byte slices.
func TestPacketUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc   string
		buf    []byte
		packet *Packet
		err    error
	}{
		{
			desc: "nil buffer, malformed packet",
			err:  ErrInvalidPacket,
		},
		{
			desc: "empty buffer, malformed packet",
			buf:  []byte{},
			err:  ErrInvalidPacket,
		},
		{
			desc: "length 1 buffer, malformed packet",
			buf:  []byte{0},
			err:  ErrInvalidPacket,
		},
		{
			desc: "length 3 buffer, malformed packet",
			buf:  []byte{0, 0, 0},
			err:  ErrInvalidPacket,
		},
		{
			desc: "invalid options in packet",
			buf:  []byte{0, 0, 0, 0, 0, 1, 0, 1},
			err:  ErrInvalidPacket,
		},
		{
			desc: "length 4 buffer, OK",
			buf:  []byte{0, 0, 0, 0},
			packet: &Packet{
				MessageType:   0,
				TransactionID: [3]byte{0, 0, 0},
				Options:       make(Options),
			},
		},
		{
			desc: "Solicit, [1 2 3] transaction ID, OK",
			buf:  []byte{1, 1, 2, 3},
			packet: &Packet{
				MessageType:   MessageTypeSolicit,
				TransactionID: [3]byte{1, 2, 3},
				Options:       make(Options),
			},
		},
		{
			desc: "Solicit, [2 3 4] transaction ID, option client ID [0 1], OK",
			buf:  []byte{1, 2, 3, 4, 0, 1, 0, 2, 0, 1},
			packet: &Packet{
				MessageType:   MessageTypeSolicit,
				TransactionID: [3]byte{2, 3, 4},
				Options: Options{
					OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		p := new(Packet)
		if err := p.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.packet, p; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected packet:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}
