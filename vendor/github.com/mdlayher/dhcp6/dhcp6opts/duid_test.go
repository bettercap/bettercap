package dhcp6opts

import (
	"bytes"
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

// TestNewDUIDLLT verifies that NewDUIDLLT generates a proper DUIDLLT or error
// from an input hardware type, time value, and hardware address.
func TestNewDUIDLLT(t *testing.T) {
	var tests = []struct {
		desc         string
		hardwareType uint16
		time         time.Time
		hardwareAddr net.HardwareAddr
		duid         *DUIDLLT
		err          error
	}{
		{
			desc: "date too early",
			time: duidLLTTime.Add(-1 * time.Minute),
			err:  ErrInvalidDUIDLLTTime,
		},
		{
			desc:         "OK",
			hardwareType: 1,
			time:         duidLLTTime.Add(1 * time.Minute),
			hardwareAddr: net.HardwareAddr([]byte{0, 1, 0, 1, 0, 1}),
			duid: &DUIDLLT{
				Type:         DUIDTypeLLT,
				HardwareType: 1,
				Time:         duidLLTTime.Add(1 * time.Minute).Sub(duidLLTTime),
				HardwareAddr: net.HardwareAddr([]byte{0, 1, 0, 1, 0, 1}),
			},
		},
	}

	for i, tt := range tests {
		duid, err := NewDUIDLLT(tt.hardwareType, tt.time, tt.hardwareAddr)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected DUIDLLT:\n- want %v\n-  got %v",
				i, tt.desc, want, got)
		}
	}
}

// TestDUIDLLTUnmarshalBinary verifies that DUIDLLT.UnmarshalBinary creates
// appropriate DUIDLLTs and errors for various input byte slices.
func TestDUIDLLTUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		buf  []byte
		duid *DUIDLLT
		err  error
	}{
		{
			desc: "nil buffer, invalid DUID-LLT",
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "empty buffer, invalid DUID-LLT",
			buf:  []byte{},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 7 buffer, invalid DUID-LLT",
			buf:  bytes.Repeat([]byte{0}, 7),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "wrong DUID type",
			buf: []byte{
				0, 2,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
			err: errInvalidDUIDLLT,
		},
		{
			desc: "OK DUIDLLT",
			buf: []byte{
				0, 1,
				0, 1,
				0, 0, 0, 60,
				0, 1, 0, 1, 0, 1,
			},
			duid: &DUIDLLT{
				Type:         DUIDTypeLLT,
				HardwareType: 1,
				Time:         1 * time.Minute,
				HardwareAddr: []byte{0, 1, 0, 1, 0, 1},
			},
		},
	}

	for i, tt := range tests {
		duid := new(DUIDLLT)
		if err := duid.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected DUID-LLT:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestNewDUIDEN verifies that NewDUIDEN generates a proper DUIDEN from
// an input enterprise number and identifier.
func TestNewDUIDEN(t *testing.T) {
	var tests = []struct {
		enterpriseNumber uint32
		identifier       []byte
		duid             *DUIDEN
	}{
		{
			enterpriseNumber: 100,
			identifier:       []byte{0, 1, 2, 3, 4},
			duid: &DUIDEN{
				Type:             DUIDTypeEN,
				EnterpriseNumber: 100,
				Identifier:       []byte{0, 1, 2, 3, 4},
			},
		},
	}

	for i, tt := range tests {
		if want, got := tt.duid, NewDUIDEN(tt.enterpriseNumber, tt.identifier); !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected DUIDEN:\n- want %v\n-  got %v", i, want, got)
		}
	}
}

// TestDUIDENUnmarshalBinary verifies that DUIDEN.UnmarshalBinary creates
// appropriate DUIDENs and errors for various input byte slices.
func TestDUIDENUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		buf  []byte
		duid *DUIDEN
		err  error
	}{
		{
			desc: "nil buffer, invalid DUID-EN",
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "empty buffer, invalid DUID-EN",
			buf:  []byte{},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 5 buffer, invalid DUID-EN",
			buf:  bytes.Repeat([]byte{0}, 5),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "wrong DUID type",
			buf: []byte{
				0, 3,
				0, 0, 0, 0,
			},
			err: errInvalidDUIDEN,
		},
		{
			desc: "OK DUIDEN",
			buf: []byte{
				0, 2,
				0, 0, 0, 100,
				0, 1, 2, 3, 4, 5,
			},
			duid: &DUIDEN{
				Type:             DUIDTypeEN,
				EnterpriseNumber: 100,
				Identifier:       []byte{0, 1, 2, 3, 4, 5},
			},
		},
	}

	for i, tt := range tests {
		duid := new(DUIDEN)
		if err := duid.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected DUID-EN:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestNewDUIDLL verifies that NewDUIDLL generates a proper DUIDLL from
// an input hardware type and hardware address.
func TestNewDUIDLL(t *testing.T) {
	var tests = []struct {
		hardwareType uint16
		hardwareAddr net.HardwareAddr
		duid         *DUIDLL
	}{
		{
			hardwareType: 1,
			hardwareAddr: net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0}),
			duid: &DUIDLL{
				Type:         DUIDTypeLL,
				HardwareType: 1,
				HardwareAddr: net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0}),
			},
		},
	}

	for i, tt := range tests {
		if want, got := tt.duid, NewDUIDLL(tt.hardwareType, tt.hardwareAddr); !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected DUIDLL:\n- want %v\n-  got %v", i, want, got)
		}
	}
}

// TestDUIDLLUnmarshalBinary verifies that DUIDLL.UnmarshalBinary creates
// appropriate DUIDLLs and errors for various input byte slices.
func TestDUIDLLUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		buf  []byte
		duid *DUIDLL
		err  error
	}{
		{
			desc: "nil buffer, invalid DUID-LL",
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "empty buffer, invalid DUID-LL",
			buf:  []byte{},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 7 buffer, invalid DUID-LL",
			buf:  bytes.Repeat([]byte{0}, 7),
			err:  errInvalidDUIDLL,
		},
		{
			desc: "wrong DUID type",
			buf: []byte{
				0, 1,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
			err: errInvalidDUIDLL,
		},
		{
			desc: "OK DUIDLL",
			buf: []byte{
				0, 3,
				0, 1,
				0, 1, 0, 1, 0, 1,
			},
			duid: &DUIDLL{
				Type:         DUIDTypeLL,
				HardwareType: 1,
				HardwareAddr: []byte{0, 1, 0, 1, 0, 1},
			},
		},
	}

	for i, tt := range tests {
		duid := new(DUIDLL)
		if err := duid.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected DUID-LL:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestNewDUIDUUID verifies that NewDUIDUUID generates a proper DUIDUUID from
// an input UUID.
func TestNewDUIDUUID(t *testing.T) {
	var tests = []struct {
		uuid [16]byte
		duid *DUIDUUID
	}{
		{
			uuid: [16]byte{
				1, 1, 1, 1,
				2, 2, 2, 2,
				3, 3, 3, 3,
				4, 4, 4, 4,
			},
			duid: &DUIDUUID{
				Type: DUIDTypeUUID,
				UUID: [16]byte{
					1, 1, 1, 1,
					2, 2, 2, 2,
					3, 3, 3, 3,
					4, 4, 4, 4,
				},
			},
		},
	}

	for i, tt := range tests {
		if want, got := tt.duid, NewDUIDUUID(tt.uuid); !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected DUIDUUID:\n- want %v\n-  got %v", i, want, got)
		}
	}
}

// TestDUIDUUIDUnmarshalBinary verifies that DUIDUUID.UnmarshalBinary returns
// appropriate DUIDUUIDs and errors for various input byte slices.
func TestDUIDUUIDUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		buf  []byte
		duid *DUIDUUID
		err  error
	}{
		{
			desc: "nil buffer, invalid DUID-UUID",
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "empty buffer, invalid DUID-UUID",
			buf:  []byte{},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 17 buffer, invalid DUID-UUID",
			buf:  bytes.Repeat([]byte{0}, 17),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "length 19 buffer, invalid DUID-UUID",
			buf:  bytes.Repeat([]byte{0}, 19),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "wrong DUID type",
			buf: []byte{
				0, 2,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			err: errInvalidDUIDUUID,
		},
		{
			desc: "OK DUIDUUID",
			buf: []byte{
				0, 4,
				1, 1, 1, 1,
				2, 2, 2, 2,
				3, 3, 3, 3,
				4, 4, 4, 4,
			},
			duid: &DUIDUUID{
				Type: DUIDTypeUUID,
				UUID: [16]byte{
					1, 1, 1, 1,
					2, 2, 2, 2,
					3, 3, 3, 3,
					4, 4, 4, 4,
				},
			},
		},
	}

	for i, tt := range tests {
		duid := new(DUIDUUID)
		if err := duid.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected DUID-UUID:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// Test_parseDUID verifies that parseDUID detects the correct DUID type for a
// variety of input data.
func Test_parseDUID(t *testing.T) {
	var tests = []struct {
		buf    []byte
		result reflect.Type
		err    error
	}{
		{
			buf: []byte{0},
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{0, 0},
			err: errUnknownDUID,
		},
		// Known types padded out to be just long enough to not error
		{
			buf:    []byte{0, 1, 0, 0, 0, 0, 0, 0},
			result: reflect.TypeOf(&DUIDLLT{}),
		},
		{
			buf:    []byte{0, 2, 0, 0, 0, 0},
			result: reflect.TypeOf(&DUIDEN{}),
		},
		{
			buf:    []byte{0, 3, 0, 0},
			result: reflect.TypeOf(&DUIDLL{}),
		},
		{
			buf:    append([]byte{0, 4}, bytes.Repeat([]byte{0}, 16)...),
			result: reflect.TypeOf(&DUIDUUID{}),
		},
		{
			buf: []byte{0, 5},
			err: errUnknownDUID,
		},
	}

	for i, tt := range tests {
		d, err := parseDUID(tt.buf)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] unexpected error for parseDUID(%v): %v != %v",
					i, tt.buf, want, got)
			}

			continue
		}

		if want, got := tt.result, reflect.TypeOf(d); want != got {
			t.Fatalf("[%02d] unexpected type for parseDUID(%v): %v != %v",
				i, tt.buf, want, got)
		}
	}
}
