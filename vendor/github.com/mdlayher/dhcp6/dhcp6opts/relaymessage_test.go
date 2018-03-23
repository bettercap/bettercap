package dhcp6opts

import (
	"bytes"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/mdlayher/dhcp6"
)

// TestRelayMessageMarshalBinary verifies that RelayMessage.MarshalBinary allocates and returns a correct
// byte slice for a variety of input data.
func TestRelayMessageMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc     string
		relayMsg *RelayMessage
		buf      []byte
	}{
		{
			desc:     "empty packet",
			relayMsg: &RelayMessage{},
			buf:      make([]byte, 34),
		},
		{
			desc: "RelayForw only",
			relayMsg: &RelayMessage{
				MessageType: dhcp6.MessageTypeRelayForw,
			},
			buf: []byte{12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			desc: "RelayReply only",
			relayMsg: &RelayMessage{
				MessageType: dhcp6.MessageTypeRelayRepl,
			},
			buf: []byte{13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			desc: "RelayForw, 15 Hopcount, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16] LinkAddress, [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32] PeerAddress",
			relayMsg: &RelayMessage{
				MessageType: dhcp6.MessageTypeRelayForw,
				HopCount:    15,
				LinkAddress: net.IP([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				PeerAddress: net.IP([]byte{17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}),
			},
			buf: []byte{12, 15, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		},
	}

	for i, tt := range tests {
		buf, err := tt.relayMsg.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := tt.buf, buf; !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected packet bytes:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestRelayMessageUnmarshalBinary verifies that RelayMessage.UnmarshalBinary returns
// appropriate RelayMessages and errors for various input byte slices.
func TestRelayMessageUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc     string
		buf      []byte
		relayMsg *RelayMessage
		err      error
	}{
		{
			desc: "nil buffer, malformed packet",
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "empty buffer, malformed packet",
			buf:  []byte{},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 1 buffer, malformed packet",
			buf:  []byte{0},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 33 buffer, malformed packet",
			buf:  make([]byte, 33),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "invalid options in packet",
			buf:  []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1},
			err:  dhcp6.ErrInvalidPacket,
		},
		{
			desc: "length 34 buffer, OK",
			buf:  []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			relayMsg: &RelayMessage{
				LinkAddress: net.IP(make([]byte, net.IPv6len)),
				PeerAddress: net.IP(make([]byte, net.IPv6len)),
				Options:     make(dhcp6.Options),
			},
		},
		{
			desc: "RelayForw, 15 Hopcount, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16] LinkAddress, [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32] PeerAddress",
			buf:  []byte{12, 15, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			relayMsg: &RelayMessage{
				MessageType: dhcp6.MessageTypeRelayForw,
				HopCount:    15,
				LinkAddress: net.IP([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				PeerAddress: net.IP([]byte{17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}),
				Options:     make(dhcp6.Options),
			},
		},
	}

	for i, tt := range tests {
		p := new(RelayMessage)
		if err := p.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.relayMsg, p; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected packet:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}
