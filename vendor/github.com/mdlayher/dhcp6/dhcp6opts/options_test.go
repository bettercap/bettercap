package dhcp6opts

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"io"
	"net"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/mdlayher/dhcp6"
)

// TestOptionsAddBinaryMarshaler verifies that dhcp6.Options.Add correctly creates or
// appends OptionCode keys with BinaryMarshaler bytes values to an dhcp6.Options map.
//
// TODO: This should just be atest for each of the binary marshalers.
func TestOptionsAddBinaryMarshaler(t *testing.T) {
	var tests = []struct {
		desc    string
		code    dhcp6.OptionCode
		bin     encoding.BinaryMarshaler
		options dhcp6.Options
	}{
		{
			desc: "DUID-LLT",
			code: dhcp6.OptionClientID,
			bin: &DUIDLLT{
				Type:         DUIDTypeLLT,
				HardwareType: 1,
				Time:         duidLLTTime.Add(1 * time.Minute).Sub(duidLLTTime),
				HardwareAddr: net.HardwareAddr([]byte{0, 1, 0, 1, 0, 1}),
			},
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{
					0, 1,
					0, 1,
					0, 0, 0, 60,
					0, 1, 0, 1, 0, 1,
				}},
			},
		},
		{
			desc: "DUID-EN",
			code: dhcp6.OptionClientID,
			bin: &DUIDEN{
				Type:             DUIDTypeEN,
				EnterpriseNumber: 100,
				Identifier:       []byte{0, 1, 2, 3, 4},
			},
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{
					0, 2,
					0, 0, 0, 100,
					0, 1, 2, 3, 4,
				}},
			},
		},
		{
			desc: "DUID-LL",
			code: dhcp6.OptionClientID,
			bin: &DUIDLL{
				Type:         DUIDTypeLL,
				HardwareType: 1,
				HardwareAddr: net.HardwareAddr([]byte{0, 1, 0, 1, 0, 1}),
			},
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{
					0, 3,
					0, 1,
					0, 1, 0, 1, 0, 1,
				}},
			},
		},
		{
			desc: "DUID-UUID",
			code: dhcp6.OptionClientID,
			bin: &DUIDUUID{
				Type: DUIDTypeUUID,
				UUID: [16]byte{
					1, 1, 1, 1,
					2, 2, 2, 2,
					3, 3, 3, 3,
					4, 4, 4, 4,
				},
			},
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{
					0, 4,
					1, 1, 1, 1,
					2, 2, 2, 2,
					3, 3, 3, 3,
					4, 4, 4, 4,
				}},
			},
		},
		{
			desc: "IA_NA",
			code: dhcp6.OptionIANA,
			bin: &IANA{
				IAID: [4]byte{0, 1, 2, 3},
				T1:   30 * time.Second,
				T2:   60 * time.Second,
			},
			options: dhcp6.Options{
				dhcp6.OptionIANA: [][]byte{{
					0, 1, 2, 3,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
		},
		{
			desc: "IA_TA",
			code: dhcp6.OptionIATA,
			bin: &IATA{
				IAID: [4]byte{0, 1, 2, 3},
			},
			options: dhcp6.Options{
				dhcp6.OptionIATA: [][]byte{{
					0, 1, 2, 3,
				}},
			},
		},
		{
			desc: "IAAddr",
			code: dhcp6.OptionIAAddr,
			bin: &IAAddr{
				IP:                net.IPv6loopback,
				PreferredLifetime: 30 * time.Second,
				ValidLifetime:     60 * time.Second,
			},
			options: dhcp6.Options{
				dhcp6.OptionIAAddr: [][]byte{{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
		},
		{
			desc: "Preference",
			code: dhcp6.OptionPreference,
			bin:  Preference(255),
			options: dhcp6.Options{
				dhcp6.OptionPreference: [][]byte{{255}},
			},
		},
		{
			desc: "ElapsedTime",
			code: dhcp6.OptionElapsedTime,
			bin:  ElapsedTime(60 * time.Second),
			options: dhcp6.Options{
				dhcp6.OptionElapsedTime: [][]byte{{23, 112}},
			},
		},
		{
			desc: "Unicast IP",
			code: dhcp6.OptionUnicast,
			bin:  IP(net.IPv6loopback),
			options: dhcp6.Options{
				dhcp6.OptionUnicast: [][]byte{{
					0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 1,
				}},
			},
		},
		{
			desc: "StatusCode",
			code: dhcp6.OptionStatusCode,
			bin: &StatusCode{
				Code:    dhcp6.StatusSuccess,
				Message: "hello world",
			},
			options: dhcp6.Options{
				dhcp6.OptionStatusCode: [][]byte{{
					0, 0,
					'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd',
				}},
			},
		},
		{
			desc: "RapidCommit",
			code: dhcp6.OptionRapidCommit,
			bin:  nil,
			options: dhcp6.Options{
				dhcp6.OptionRapidCommit: [][]byte{nil},
			},
		},
		{
			desc: "Data (UserClass, VendorClass, BootFileParam)",
			code: dhcp6.OptionUserClass,
			bin: Data{
				[]byte{0},
				[]byte{0, 1},
				[]byte{0, 1, 2},
			},
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{
					0, 1, 0,
					0, 2, 0, 1,
					0, 3, 0, 1, 2,
				}},
			},
		},
		{
			desc: "IA_PD",
			code: dhcp6.OptionIAPD,
			bin: &IAPD{
				IAID: [4]byte{0, 1, 2, 3},
				T1:   30 * time.Second,
				T2:   60 * time.Second,
			},
			options: dhcp6.Options{
				dhcp6.OptionIAPD: [][]byte{{
					0, 1, 2, 3,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
		},
		{
			desc: "IAPrefix",
			code: dhcp6.OptionIAPrefix,
			bin: &IAPrefix{
				PreferredLifetime: 30 * time.Second,
				ValidLifetime:     60 * time.Second,
				PrefixLength:      64,
				Prefix: net.IP{
					1, 1, 1, 1, 1, 1, 1, 1,
					0, 0, 0, 0, 0, 0, 0, 0,
				},
			},
			options: dhcp6.Options{
				dhcp6.OptionIAPrefix: [][]byte{{
					0, 0, 0, 30,
					0, 0, 0, 60,
					64,
					1, 1, 1, 1, 1, 1, 1, 1,
					0, 0, 0, 0, 0, 0, 0, 0,
				}},
			},
		},
		{
			desc: "URL",
			code: dhcp6.OptionBootFileURL,
			bin: &URL{
				Scheme: "tftp",
				Host:   "192.168.1.1:69",
			},
			options: dhcp6.Options{
				dhcp6.OptionBootFileURL: [][]byte{[]byte("tftp://192.168.1.1:69")},
			},
		},
		{
			desc: "ArchTypes",
			code: dhcp6.OptionClientArchType,
			bin: ArchTypes{
				ArchTypeEFIx8664,
				ArchTypeIntelx86PC,
				ArchTypeIntelLeanClient,
			},
			options: dhcp6.Options{
				dhcp6.OptionClientArchType: [][]byte{{0, 9, 0, 0, 0, 5}},
			},
		},
		{
			desc: "NII",
			code: dhcp6.OptionNII,
			bin: &NII{
				Type:  1,
				Major: 2,
				Minor: 3,
			},
			options: dhcp6.Options{
				dhcp6.OptionNII: [][]byte{{1, 2, 3}},
			},
		},
	}

	for i, tt := range tests {
		o := make(dhcp6.Options)
		if err := o.Add(tt.code, tt.bin); err != nil {
			t.Fatal(err)
		}

		if want, got := tt.options, o; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected dhcp6.Options map:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetClientID verifies that dhcp6.Options.ClientID properly parses and returns
// a DUID value, if one is available with OptionClientID.
func TestGetClientID(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		duid    DUID
		err     error
	}{
		{
			desc: "OptionClientID not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionClientID present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{
					0, 3,
					0, 1,
					0, 1, 0, 1, 0, 1,
				}},
			},
			duid: &DUIDLL{
				Type:         DUIDTypeLL,
				HardwareType: 1,
				HardwareAddr: []byte{0, 1, 0, 1, 0, 1},
			},
		},
	}

	for i, tt := range tests {
		// DUID parsing is tested elsewhere, so errors should automatically fail
		// test here
		duid, err := GetClientID(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.ClientID(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.ClientID():\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetServerID verifies that dhcp6.Options.ServerID properly parses and returns
// a DUID value, if one is available with OptionServerID.
func TestGetServerID(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		duid    DUID
		err     error
	}{
		{
			desc: "OptionServerID not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionServerID present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionServerID: [][]byte{{
					0, 3,
					0, 1,
					0, 1, 0, 1, 0, 1,
				}},
			},
			duid: &DUIDLL{
				Type:         DUIDTypeLL,
				HardwareType: 1,
				HardwareAddr: []byte{0, 1, 0, 1, 0, 1},
			},
		},
	}

	for i, tt := range tests {
		// DUID parsing is tested elsewhere, so errors should automatically fail
		// test here
		duid, err := GetServerID(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.ServerID(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.duid, duid; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.ServerID():\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetIANA verifies that dhcp6.Options.IANA properly parses and
// returns multiple IANA values, if one or more are available with OptionIANA.
func TestGetIANA(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		iana    []*IANA
		err     error
	}{
		{
			desc: "OptionIANA not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionIANA present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionIANA: [][]byte{bytes.Repeat([]byte{0}, 11)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "one OptionIANA present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIANA: [][]byte{{
					1, 2, 3, 4,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
			iana: []*IANA{
				{
					IAID: [4]byte{1, 2, 3, 4},
					T1:   30 * time.Second,
					T2:   60 * time.Second,
				},
			},
		},
		{
			desc: "two OptionIANA present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIANA: [][]byte{
					append(bytes.Repeat([]byte{0}, 12), []byte{0, 1, 0, 1, 1}...),
					append(bytes.Repeat([]byte{0}, 12), []byte{0, 2, 0, 1, 2}...),
				},
			},
			iana: []*IANA{
				{
					Options: dhcp6.Options{
						dhcp6.OptionClientID: [][]byte{{1}},
					},
				},
				{
					Options: dhcp6.Options{
						dhcp6.OptionServerID: [][]byte{{2}},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		iana, err := GetIANA(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.IANA: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		for j := range tt.iana {
			want, err := tt.iana[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			got, err := iana[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.IANA():\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetIATA verifies that dhcp6.Options.IATA properly parses and
// returns multiple IATA values, if one or more are available with OptionIATA.
func TestGetIATA(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		iata    []*IATA
		err     error
	}{
		{
			desc: "OptionIATA not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionIATA present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionIATA: [][]byte{{0, 0, 0}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "one OptionIATA present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIATA: [][]byte{{
					1, 2, 3, 4,
				}},
			},
			iata: []*IATA{
				{
					IAID: [4]byte{1, 2, 3, 4},
				},
			},
		},
		{
			desc: "two OptionIATA present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIATA: [][]byte{
					{0, 1, 2, 3, 0, 1, 0, 1, 1},
					{4, 5, 6, 7, 0, 2, 0, 1, 2},
				},
			},
			iata: []*IATA{
				{
					IAID: [4]byte{0, 1, 2, 3},
					Options: dhcp6.Options{
						dhcp6.OptionClientID: [][]byte{{1}},
					},
				},
				{
					IAID: [4]byte{4, 5, 6, 7},
					Options: dhcp6.Options{
						dhcp6.OptionServerID: [][]byte{{2}},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		iata, err := GetIATA(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.IATA: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		for j := range tt.iata {
			want, err := tt.iata[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			got, err := iata[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.IATA():\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetIAAddr verifies that dhcp6.Options.IAAddr properly parses and
// returns multiple IAAddr values, if one or more are available with
// OptionIAAddr.
func TestGetIAAddr(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		iaaddr  []*IAAddr
		err     error
	}{
		{
			desc: "OptionIAAddr not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionIAAddr present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionIAAddr: [][]byte{bytes.Repeat([]byte{0}, 23)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "one OptionIAAddr present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAAddr: [][]byte{{
					0, 0, 0, 0,
					1, 1, 1, 1,
					2, 2, 2, 2,
					3, 3, 3, 3,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
			iaaddr: []*IAAddr{
				{
					IP: net.IP{
						0, 0, 0, 0,
						1, 1, 1, 1,
						2, 2, 2, 2,
						3, 3, 3, 3,
					},
					PreferredLifetime: 30 * time.Second,
					ValidLifetime:     60 * time.Second,
				},
			},
		},
		{
			desc: "two OptionIAAddr present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAAddr: [][]byte{
					bytes.Repeat([]byte{0}, 24),
					bytes.Repeat([]byte{0}, 24),
				},
			},
			iaaddr: []*IAAddr{
				{
					IP: net.IPv6zero,
				},
				{
					IP: net.IPv6zero,
				},
			},
		},
	}

	for i, tt := range tests {
		iaaddr, err := GetIAAddr(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.IAAddr: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		for j := range tt.iaaddr {
			want, err := tt.iaaddr[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			got, err := iaaddr[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.IAAddr():\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetOptionRequest verifies that dhcp6.Options.OptionRequest properly parses
// and returns a slice of OptionCode values, if they are available with
// OptionORO.
func TestGetOptionRequest(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		codes   OptionRequestOption
		err     error
	}{
		{
			desc: "OptionORO not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionORO present in dhcp6.Options map, but not even length",
			options: dhcp6.Options{
				dhcp6.OptionORO: [][]byte{{0}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionORO present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionORO: [][]byte{{0, 1}},
			},
			codes: []dhcp6.OptionCode{1},
		},
		{
			desc: "OptionORO present in dhcp6.Options map, with multiple values",
			options: dhcp6.Options{
				dhcp6.OptionORO: [][]byte{{0, 1, 0, 2, 0, 3}},
			},
			codes: []dhcp6.OptionCode{1, 2, 3},
		},
	}

	for i, tt := range tests {
		codes, err := GetOptionRequest(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.OptionRequest(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.codes, codes; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.OptionRequest():\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetPreference verifies that dhcp6.Options.Preference properly parses
// and returns an integer value, if it is available with OptionPreference.
func TestGetPreference(t *testing.T) {
	var tests = []struct {
		desc       string
		options    dhcp6.Options
		preference Preference
		err        error
	}{
		{
			desc: "OptionPreference not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionPreference present in dhcp6.Options map, but too short length",
			options: dhcp6.Options{
				dhcp6.OptionPreference: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionPreference present in dhcp6.Options map, but too long length",
			options: dhcp6.Options{
				dhcp6.OptionPreference: [][]byte{{0, 1}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionPreference present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionPreference: [][]byte{{255}},
			},
			preference: 255,
		},
	}

	for i, tt := range tests {
		preference, err := GetPreference(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.Preference(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.preference, preference; want != got {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.Preference(): %v != %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetUnicast verifies that dhcp6.Options.Unicast properly parses
// and returns an IPv6 address or an error, if available with OptionUnicast.
func TestGetUnicast(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		ip      IP
		err     error
	}{
		{
			desc: "OptionUnicast not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionUnicast present in dhcp6.Options map, but too short length",
			options: dhcp6.Options{
				dhcp6.OptionUnicast: [][]byte{bytes.Repeat([]byte{0}, 15)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionUnicast present in dhcp6.Options map, but too long length",
			options: dhcp6.Options{
				dhcp6.OptionUnicast: [][]byte{bytes.Repeat([]byte{0}, 17)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionUnicast present in dhcp6.Options map with IPv4 address",
			options: dhcp6.Options{
				dhcp6.OptionUnicast: [][]byte{net.IPv4(192, 168, 1, 1)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionUnicast present in dhcp6.Options map with IPv6 address",
			options: dhcp6.Options{
				dhcp6.OptionUnicast: [][]byte{net.IPv6loopback},
			},
			ip: IP(net.IPv6loopback),
		},
	}

	for i, tt := range tests {
		ip, err := GetUnicast(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.Unicast(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.ip, ip; !bytes.Equal(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.Unicast():\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetStatusCode verifies that dhcp6.Options.StatusCode properly parses
// and returns a StatusCode value, if it is available with OptionStatusCode.
func TestGetStatusCode(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		sc      *StatusCode
		err     error
	}{
		{
			desc: "OptionStatusCode not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionStatusCode present in dhcp6.Options map, but too short length",
			options: dhcp6.Options{
				dhcp6.OptionStatusCode: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionStatusCode present in dhcp6.Options map, no message",
			options: dhcp6.Options{
				dhcp6.OptionStatusCode: [][]byte{{0, 0}},
			},
			sc: &StatusCode{
				Code: dhcp6.StatusSuccess,
			},
		},
		{
			desc: "OptionStatusCode present in dhcp6.Options map, with message",
			options: dhcp6.Options{
				dhcp6.OptionStatusCode: [][]byte{append([]byte{0, 0}, []byte("deadbeef")...)},
			},
			sc: &StatusCode{
				Code:    dhcp6.StatusSuccess,
				Message: "deadbeef",
			},
		},
	}

	for i, tt := range tests {
		sc, err := GetStatusCode(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.StatusCode(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.sc, sc; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.StatusCode():\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetElapsedTime verifies that dhcp6.Options.ElapsedTime properly parses and
// returns a time.Duration value, if one is available with OptionElapsedTime.
func TestGetElapsedTime(t *testing.T) {
	var tests = []struct {
		desc     string
		options  dhcp6.Options
		duration ElapsedTime
		err      error
	}{
		{
			desc: "OptionElapsedTime not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionElapsedTime present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionElapsedTime: [][]byte{{1}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionElapsedTime present in dhcp6.Options map, but too long",
			options: dhcp6.Options{
				dhcp6.OptionElapsedTime: [][]byte{{1, 2, 3}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionElapsedTime present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionElapsedTime: [][]byte{{1, 1}},
			},
			duration: ElapsedTime(2570 * time.Millisecond),
		},
	}

	for i, tt := range tests {
		duration, err := GetElapsedTime(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.ElapsedTime: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.duration, duration; want != got {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.ElapsedTime(): %v != %v",
				i, tt.desc, want, got)
		}
	}
}

// TestElapsedTimeMarshalBinary verifies that dhcp6.Options.ElapsedTime properly
// marsharls into bytes array.
func TestElapsedTimeMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc        string
		elapsedTime ElapsedTime
		buf         []byte
		err         error
	}{
		{
			desc: "OptionElapsedTime elapsed-time = 0",
			buf:  []byte{0, 0},
		},
		{
			desc:        "OptionElapsedTime elapsed-time = 65534 hundredths of a second",
			elapsedTime: ElapsedTime(655340 * time.Millisecond),
			buf:         []byte{0xff, 0xfe},
		},
		{
			desc:        "OptionElapsedTime elapsed-time = 65535 hundredths of a second",
			elapsedTime: ElapsedTime(655350 * time.Millisecond),
			buf:         []byte{0xff, 0xff},
		},
		{
			desc:        "OptionElapsedTime elapsed-time = 65537 hundredths of a second",
			elapsedTime: ElapsedTime(655370 * time.Millisecond),
			buf:         []byte{0xff, 0xff},
		},
	}

	for i, tt := range tests {
		buf, err := tt.elapsedTime.MarshalBinary()
		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.ElapsedTime\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.buf, buf; !bytes.Equal(want, got) {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.ElapsedTime\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetRelayMessage verifies that dhcp6.Options.RelayMessageOption properly parses and
// returns an relay message option value, if one is available with RelayMessageOption.
func TestGetRelayMessage(t *testing.T) {
	var tests = []struct {
		desc           string
		options        dhcp6.Options
		authentication RelayMessageOption
		err            error
	}{
		{
			desc: "RelayMessageOption not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "RelayMessageOption present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionRelayMsg: [][]byte{{1, 1, 2, 3}},
			},
			authentication: []byte{1, 1, 2, 3},
		},
	}

	for i, tt := range tests {
		relayMsg, err := GetRelayMessageOption(tt.options)
		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.RelayMessageOption\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.authentication, relayMsg; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.RelayMessageOption()\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.RelayMessageOption(): %v != %v", i, tt.desc, want, got)
		}
	}
}

// TestAuthentication verifies that dhcp6.Options.Authentication properly parses and
// returns an authentication value, if one is available with Authentication.
func TestAuthentication(t *testing.T) {
	var tests = []struct {
		desc           string
		options        dhcp6.Options
		authentication *Authentication
		err            error
	}{
		{
			desc: "Authentication not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "Authentication present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionAuth: [][]byte{bytes.Repeat([]byte{0}, 10)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "Authentication present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionAuth: [][]byte{{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}},
			},
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
		authentication, err := GetAuthentication(tt.options)
		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.Authentication\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.authentication, authentication; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.Authentication()\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.Authentication(): %v != %v", i, tt.desc, want, got)
		}
	}
}

// TestGetRapidCommit verifies that dhcp6.Options.RapidCommit properly indicates
// if OptionRapidCommit was present in dhcp6.Options.
func TestGetRapidCommit(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		err     error
	}{
		{
			desc: "OptionRapidCommit not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionRapidCommit present in dhcp6.Options map, but non-empty",
			options: dhcp6.Options{
				dhcp6.OptionRapidCommit: [][]byte{{1}},
			},
			err: dhcp6.ErrInvalidPacket,
		},
		{
			desc: "OptionRapidCommit present in dhcp6.Options map, empty",
			options: dhcp6.Options{
				dhcp6.OptionRapidCommit: [][]byte{},
			},
		},
	}

	for i, tt := range tests {
		err := GetRapidCommit(tt.options)
		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.RapidCommit: %v != %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetUserClass verifies that dhcp6.Options.UserClass properly parses
// and returns raw user class data, if it is available with OptionUserClass.
func TestGetUserClass(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		classes [][]byte
		err     error
	}{
		{
			desc: "OptionUserClass not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionUserClass present in dhcp6.Options map, but empty",
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionUserClass present in dhcp6.Options map, one item, zero length",
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{
					0, 0,
				}},
			},
			classes: [][]byte{{}},
		},
		{
			desc: "OptionUserClass present in dhcp6.Options map, one item, extra byte",
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{
					0, 1, 1, 255,
				}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionUserClass present in dhcp6.Options map, one item",
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{
					0, 1, 1,
				}},
			},
			classes: [][]byte{{1}},
		},
		{
			desc: "OptionUserClass present in dhcp6.Options map, three items",
			options: dhcp6.Options{
				dhcp6.OptionUserClass: [][]byte{{
					0, 1, 1,
					0, 2, 2, 2,
					0, 3, 3, 3, 3,
				}},
			},
			classes: [][]byte{{1}, {2, 2}, {3, 3, 3}},
		},
	}

	for i, tt := range tests {
		classes, err := GetUserClass(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.UserClass: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := len(tt.classes), len(classes); want != got {
			t.Errorf("[%02d] test %q, unexpected classes slice length: %v != %v",
				i, tt.desc, want, got)

		}

		for j := range classes {
			if want, got := tt.classes[j], classes[j]; !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.UserClass()\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetVendorClass verifies that dhcp6.Options.VendorClass properly parses
// and returns raw vendor class data, if it is available with OptionVendorClass.
func TestGetVendorClass(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		classes [][]byte
		err     error
	}{
		{
			desc: "OptionVendorClass not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, but empty",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, zero item",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{
					0, 0, 5, 0x58,
				}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, one item, zero length",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{
					0, 0, 5, 0x58,
					0, 0,
				}},
			},
			classes: [][]byte{{}},
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, one item, extra byte",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{
					0, 0, 5, 0x58,
					0, 1, 1, 255,
				}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, one item",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{
					0, 0, 5, 0x58,
					0, 1, 1,
				}},
			},
			classes: [][]byte{{1}},
		},
		{
			desc: "OptionVendorClass present in dhcp6.Options map, three items",
			options: dhcp6.Options{
				dhcp6.OptionVendorClass: [][]byte{{
					0, 0, 5, 0x58,
					0, 1, 1,
					0, 2, 2, 2,
					0, 3, 3, 3, 3,
				}},
			},
			classes: [][]byte{{1}, {2, 2}, {3, 3, 3}},
		},
	}

	for i, tt := range tests {
		classes, err := GetVendorClass(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.VendorClass: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := len(tt.classes), len(classes.VendorClassData); want != got {
			t.Errorf("[%02d] test %q, unexpected classes slice length: %v != %v",
				i, tt.desc, want, got)

		}

		for j := range classes.VendorClassData {
			if want, got := tt.classes[j], classes.VendorClassData[j]; !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.VendorClass()\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestInterfaceID verifies that dhcp6.Options.InterfaceID properly parses
// and returns raw interface-id data, if it is available with InterfaceID.
func TestInterfaceID(t *testing.T) {
	var tests = []struct {
		desc        string
		options     dhcp6.Options
		interfaceID InterfaceID
		err         error
	}{
		{
			desc: "InterfaceID not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "InterfaceID present in dhcp6.Options map, one item",
			options: dhcp6.Options{
				dhcp6.OptionInterfaceID: [][]byte{{
					0, 1, 1,
				}},
			},
			interfaceID: []byte{0, 1, 1},
		},
		{
			desc: "InterfaceID present in dhcp6.Options map with no interface-id data",
			options: dhcp6.Options{
				dhcp6.OptionInterfaceID: [][]byte{{}},
			},
			interfaceID: []byte{},
		},
	}

	for i, tt := range tests {
		interfaceID, err := GetInterfaceID(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.InterfaceID: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.interfaceID, interfaceID; !bytes.Equal(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.InterfaceID()\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetIAPD verifies that dhcp6.Options.IAPD properly parses and
// returns multiple IAPD values, if one or more are available with OptionIAPD.
func TestGetIAPD(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		iapd    []*IAPD
		err     error
	}{
		{
			desc: "OptionIAPD not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionIAPD present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionIAPD: [][]byte{bytes.Repeat([]byte{0}, 11)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "one OptionIAPD present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAPD: [][]byte{{
					1, 2, 3, 4,
					0, 0, 0, 30,
					0, 0, 0, 60,
				}},
			},
			iapd: []*IAPD{
				{
					IAID: [4]byte{1, 2, 3, 4},
					T1:   30 * time.Second,
					T2:   60 * time.Second,
				},
			},
		},
		{
			desc: "two OptionIAPD present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAPD: [][]byte{
					append(bytes.Repeat([]byte{0}, 12), []byte{0, 1, 0, 1, 1}...),
					append(bytes.Repeat([]byte{0}, 12), []byte{0, 2, 0, 1, 2}...),
				},
			},
			iapd: []*IAPD{
				{
					Options: dhcp6.Options{
						dhcp6.OptionClientID: [][]byte{{1}},
					},
				},
				{
					Options: dhcp6.Options{
						dhcp6.OptionServerID: [][]byte{{2}},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		iapd, err := GetIAPD(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.IAPD: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		for j := range tt.iapd {
			want, err := tt.iapd[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			got, err := iapd[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.IAPD():\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetIAPrefix verifies that dhcp6.Options.IAPrefix properly parses and
// returns multiple IAPrefix values, if one or more are available with
// OptionIAPrefix.
func TestGetIAPrefix(t *testing.T) {
	var tests = []struct {
		desc     string
		options  dhcp6.Options
		iaprefix []*IAPrefix
		err      error
	}{
		{
			desc: "OptionIAPrefix not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionIAPrefix present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionIAPrefix: [][]byte{bytes.Repeat([]byte{0}, 24)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "one OptionIAPrefix present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAPrefix: [][]byte{{
					0, 0, 0, 30,
					0, 0, 0, 60,
					32,
					32, 1, 13, 184, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0,
				}},
			},
			iaprefix: []*IAPrefix{
				{
					PreferredLifetime: 30 * time.Second,
					ValidLifetime:     60 * time.Second,
					PrefixLength:      32,
					Prefix: net.IP{
						32, 1, 13, 184, 0, 0, 0, 0,
						0, 0, 0, 0, 0, 0, 0, 0,
					},
				},
			},
		},
		{
			desc: "two OptionIAPrefix present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionIAPrefix: [][]byte{
					bytes.Repeat([]byte{0}, 25),
					bytes.Repeat([]byte{0}, 25),
				},
			},
			iaprefix: []*IAPrefix{
				{
					Prefix: net.IPv6zero,
				},
				{
					Prefix: net.IPv6zero,
				},
			},
		},
	}

	for i, tt := range tests {
		iaprefix, err := GetIAPrefix(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.IAPrefix: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		for j := range tt.iaprefix {
			want, err := tt.iaprefix[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			got, err := iaprefix[j].MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.IAPrefix():\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestRemoteIdentifier verifies that dhcp6.Options.RemoteIdentifier properly parses
// and returns a RemoteIdentifier, if it is available with dhcp6.OptionsRemoteIdentifier.
func TestGetRemoteIdentifier(t *testing.T) {
	var tests = []struct {
		desc             string
		options          dhcp6.Options
		remoteIdentifier *RemoteIdentifier
		err              error
	}{
		{
			desc: "OptionsRemoteIdentifier not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionsRemoteIdentifier present in dhcp6.Options map, but too short",
			options: dhcp6.Options{
				dhcp6.OptionRemoteIdentifier: [][]byte{{
					0, 0, 5, 0x58,
				}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionsRemoteIdentifier present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionRemoteIdentifier: [][]byte{{
					0, 0, 5, 0x58,
					0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf,
				}},
			},
			remoteIdentifier: &RemoteIdentifier{
				EnterpriseNumber: 1368,
				RemoteID:         []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xe, 0xf},
			},
		},
	}
	for i, tt := range tests {
		remoteIdentifier, err := GetRemoteIdentifier(tt.options)
		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.RemoteIdentifier\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.remoteIdentifier, remoteIdentifier; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.RemoteIdentifier()\n- want: %v\n-  got: %v", i, tt.desc, want, got)
		}

		if want, got := tt.err, err; want != got {
			t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.RemoteIdentifier(): %v != %v", i, tt.desc, want, got)
		}
	}
}

// TestGetBootFileURL verifies that dhcp6.Options.BootFileURL properly parses
// and returns a URL, if it is available with OptionBootFileURL.
func TestGetBootFileURL(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		u       *url.URL
		err     error
	}{
		{
			desc: "OptionBootFileURL not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionBootFileURL present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionBootFileURL: [][]byte{[]byte("tftp://192.168.1.1:69")},
			},
			u: &url.URL{
				Scheme: "tftp",
				Host:   "192.168.1.1:69",
			},
		},
	}

	for i, tt := range tests {
		u, err := GetBootFileURL(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected err for dhcp6.Options.BootFileURL(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		ttuu := url.URL(*tt.u)
		uu := url.URL(*u)
		if want, got := ttuu.String(), uu.String(); want != got {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.BootFileURL(): %v != %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetBootFileParam verifies that dhcp6.Options.BootFileParam properly parses
// and returns boot file parameter data, if it is available with
// OptionBootFileParam.
func TestGetBootFileParam(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		param   BootFileParam
		err     error
	}{
		{
			desc: "OptionBootFileParam not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionBootFileParam present in dhcp6.Options map, but empty",
			options: dhcp6.Options{
				dhcp6.OptionBootFileParam: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionBootFileParam present in dhcp6.Options map, one item, zero length",
			options: dhcp6.Options{
				dhcp6.OptionBootFileParam: [][]byte{{
					0, 0,
				}},
			},
			param: []string{""},
		},
		{
			desc: "OptionBootFileParam present in dhcp6.Options map, one item, extra byte",
			options: dhcp6.Options{
				dhcp6.OptionBootFileParam: [][]byte{{
					0, 1, 1, 255,
				}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionBootFileParam present in dhcp6.Options map, one item",
			options: dhcp6.Options{
				dhcp6.OptionBootFileParam: [][]byte{{
					0, 3, 'f', 'o', 'o',
				}},
			},
			param: []string{"foo"},
		},
		{
			desc: "OptionBootFileParam present in dhcp6.Options map, three items",
			options: dhcp6.Options{
				dhcp6.OptionBootFileParam: [][]byte{{
					0, 1, 'a',
					0, 2, 'a', 'b',
					0, 3, 'a', 'b', 'c',
				}},
			},
			param: []string{"a", "ab", "abc"},
		},
	}

	for i, tt := range tests {
		param, err := GetBootFileParam(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.BootFileParam: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := len(tt.param), len(param); want != got {
			t.Errorf("[%02d] test %q, unexpected param slice length: %v != %v",
				i, tt.desc, want, got)

		}

		for j := range param {
			if want, got := tt.param[j], param[j]; want != got {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.BootFileParam()\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetClientArchType verifies that dhcp6.Options.ClientArchType properly parses
// and returns client architecture type data, if it is available with
// OptionClientArchType.
func TestGetClientArchType(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		arch    ArchTypes
		err     error
	}{
		{
			desc: "OptionClientArchType not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionClientArchType present in dhcp6.Options map, but empty",
			options: dhcp6.Options{
				dhcp6.OptionClientArchType: [][]byte{{}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionClientArchType present in dhcp6.Options map, but not divisible by 2",
			options: dhcp6.Options{
				dhcp6.OptionClientArchType: [][]byte{{0, 0, 0}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionClientArchType present in dhcp6.Options map, one architecture",
			options: dhcp6.Options{
				dhcp6.OptionClientArchType: [][]byte{{0, 9}},
			},
			arch: ArchTypes{ArchTypeEFIx8664},
		},
		{
			desc: "OptionClientArchType present in dhcp6.Options map, three architectures",
			options: dhcp6.Options{
				dhcp6.OptionClientArchType: [][]byte{{0, 5, 0, 9, 0, 0}},
			},
			arch: ArchTypes{
				ArchTypeIntelLeanClient,
				ArchTypeEFIx8664,
				ArchTypeIntelx86PC,
			},
		},
	}

	for i, tt := range tests {
		arch, err := GetClientArchType(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.ClientArchType: %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := len(tt.arch), len(arch); want != got {
			t.Errorf("[%02d] test %q, unexpected arch slice length: %v != %v",
				i, tt.desc, want, got)
		}

		for j := range arch {
			if want, got := tt.arch[j], arch[j]; !reflect.DeepEqual(want, got) {
				t.Errorf("[%02d:%02d] test %q, unexpected value for dhcp6.Options.ClientArchType()\n- want: %v\n-  got: %v",
					i, j, tt.desc, want, got)
			}
		}
	}
}

// TestGetNII verifies that dhcp6.Options.NII properly parses and returns a
// Network Interface Identifier value, if it is available with OptionNII.
func TestGetNII(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		nii     *NII
		err     error
	}{
		{
			desc: "OptionNII not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionNII present in dhcp6.Options map, but too short length",
			options: dhcp6.Options{
				dhcp6.OptionNII: [][]byte{{1, 2}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionNII present in dhcp6.Options map, but too long length",
			options: dhcp6.Options{
				dhcp6.OptionNII: [][]byte{{1, 2, 3, 4}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionNII present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionNII: [][]byte{{1, 2, 3}},
			},
			nii: &NII{
				Type:  1,
				Major: 2,
				Minor: 3,
			},
		},
	}

	for i, tt := range tests {
		nii, err := GetNII(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for dhcp6.Options.NII(): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.nii, nii; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for dhcp6.Options.NII(): %v != %v",
				i, tt.desc, want, got)
		}
	}
}

// TestGetDNSServers verifies that dhcp6opts.GetDNSServers properly parses and
// returns a list of net.IPs, if it is available with OptionDNSServers.
func TestGetDNSServers(t *testing.T) {
	var tests = []struct {
		desc    string
		options dhcp6.Options
		dns     IPs
		err     error
	}{
		{
			desc: "OptionDNSServers not present in dhcp6.Options map",
			err:  dhcp6.ErrOptionNotPresent,
		},
		{
			desc: "OptionDNSServers present in dhcp6.Options map, but too short length",
			options: dhcp6.Options{
				dhcp6.OptionDNSServers: [][]byte{{255, 255, 255}},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "OptionDNSServers present in dhcp6.Options map, but too long length",
			options: dhcp6.Options{
				dhcp6.OptionDNSServers: [][]byte{bytes.Repeat([]byte{0xff}, 17)},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			desc: "One OptionDNSServers present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionDNSServers: [][]byte{bytes.Repeat([]byte{0xfe}, 16)},
			},
			dns: IPs{
				net.IP(bytes.Repeat([]byte{0xfe}, 16)),
			},
		},
		{
			desc: "Two OptionDNSServers present in dhcp6.Options map",
			options: dhcp6.Options{
				dhcp6.OptionDNSServers: [][]byte{append(bytes.Repeat([]byte{0xfd}, 16), bytes.Repeat([]byte{0xfc}, 16)...)},
			},
			dns: IPs{
				net.IP(bytes.Repeat([]byte{0xfd}, 16)),
				net.IP(bytes.Repeat([]byte{0xfc}, 16)),
			},
		},
	}

	for i, tt := range tests {
		dns, err := GetDNSServers(tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for GetDNSServers(dhcp6.Options): %v != %v",
					i, tt.desc, want, got)
			}
			continue
		}

		if want, got := tt.dns, dns; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for GetDNSServers(dhcp6.Options): %v != %v",
				i, tt.desc, want, got)
		}
	}
}
