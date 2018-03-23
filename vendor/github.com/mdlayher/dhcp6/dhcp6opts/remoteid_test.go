package dhcp6opts

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestRemoteIdentifierMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc             string
		buf              []byte
		remoteIdentifier *RemoteIdentifier
	}{
		{
			desc: "all zero values",
			buf:  bytes.Repeat([]byte{0}, 5),
			remoteIdentifier: &RemoteIdentifier{
				RemoteID: []byte{0},
			},
		},
		{
			desc: "[0, 0, 5, 0x58] EnterpriseNumber, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf] RemoteID",
			buf:  []byte{0, 0, 5, 0x58, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf},
			remoteIdentifier: &RemoteIdentifier{
				EnterpriseNumber: 1368,
				RemoteID:         []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf},
			},
		},
	}

	for i, tt := range tests {
		want := tt.buf
		got, err := tt.remoteIdentifier.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected RemoteIdentifier bytes for MarshalBinary(%v)\n- want: %v\n-  got: %v",
				i, tt.desc, tt.buf, want, got)
		}
	}
}

func TestRemoteIdentifierUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf              []byte
		remoteIdentifier *RemoteIdentifier
		err              error
	}{
		{
			buf: []byte{0, 0, 5, 0x58, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf},
			remoteIdentifier: &RemoteIdentifier{
				EnterpriseNumber: 1368,
				RemoteID:         []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf},
			},
		},
		{
			buf: bytes.Repeat([]byte{0}, 4),
			err: io.ErrUnexpectedEOF,
		},
	}

	for i, tt := range tests {
		remoteIdentifier := new(RemoteIdentifier)
		if want, got := tt.err, remoteIdentifier.UnmarshalBinary(tt.buf); want != got {
			t.Fatalf("[%02d] unexpected error for parseRemoteIdentifier(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
		}

		if tt.err == nil {
			if want, got := tt.remoteIdentifier, remoteIdentifier; !reflect.DeepEqual(want, got) {
				t.Fatalf("[%02d] unexpected RemoteIdentifier for parseRemoteIdentifier(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
			}
		}
	}
}
