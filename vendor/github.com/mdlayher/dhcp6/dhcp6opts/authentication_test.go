package dhcp6opts

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"testing"
)

func TestAuthenticationMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc           string
		buf            []byte
		authentication *Authentication
	}{
		{
			desc:           "all zero values",
			buf:            bytes.Repeat([]byte{0}, 11),
			authentication: &Authentication{},
		},
		{
			desc: "Protocol: 0, Alforithm: 1, RDM: 2, [3, 4, 5, 6, 7, 8, 9, 0xa] ReplayDetection, [0xb, 0xc, 0xd, 0xe, 0xf] AuthenticationInformation",
			buf:  []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
			authentication: &Authentication{
				Protocol:                  0,
				Algorithm:                 1,
				RDM:                       2,
				ReplayDetection:           binary.BigEndian.Uint64([]byte{3, 4, 5, 6, 7, 8, 9, 0xa}),
				AuthenticationInformation: []byte{0xb, 0xc, 0xd, 0xe, 0xf},
			},
		},
	}

	for i, tt := range tests {
		want := tt.buf
		got, err := tt.authentication.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected Authentication bytes for MarshalBinary(%v)\n- want: %v\n-  got: %v",
				i, tt.desc, tt.buf, want, got)
		}
	}
}

func TestAuthenticationUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf            []byte
		authentication *Authentication
		err            error
	}{
		{
			buf: bytes.Repeat([]byte{0}, 10),
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
			authentication: &Authentication{
				Protocol:                  0,
				Algorithm:                 1,
				RDM:                       2,
				ReplayDetection:           binary.BigEndian.Uint64([]byte{3, 4, 5, 6, 7, 8, 9, 0xa}),
				AuthenticationInformation: []byte{0xb, 0xc, 0xd, 0xe, 0xf},
			},
		},
	}

	for i, tt := range tests {
		authentication := new(Authentication)
		if want, got := tt.err, authentication.UnmarshalBinary(tt.buf); want != got {
			t.Fatalf("[%02d] unexpected error for parseAuthentication(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.authentication, authentication; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected Authentication for parseAuthentication(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
		}
	}
}
