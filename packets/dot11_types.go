package packets

import (
	"encoding/binary"
	"fmt"
)

type Dot11CipherType uint8

const (
	Dot11CipherWep    Dot11CipherType = 1
	Dot11CipherTkip   Dot11CipherType = 2
	Dot11CipherWrap   Dot11CipherType = 3
	Dot11CipherCcmp   Dot11CipherType = 4
	Dot11CipherWep104 Dot11CipherType = 5
)

func (a Dot11CipherType) String() string {
	switch a {
	case Dot11CipherWep:
		return "WEP"
	case Dot11CipherTkip:
		return "TKIP"
	case Dot11CipherWrap:
		return "WRAP"
	case Dot11CipherCcmp:
		return "CCMP"
	case Dot11CipherWep104:
		return "WEP104"
	default:
		return "UNK"
	}
}

type Dot11AuthType uint8

const (
	Dot11AuthMgt Dot11AuthType = 1
	Dot11AuthPsk Dot11AuthType = 2
)

func (a Dot11AuthType) String() string {
	switch a {
	case Dot11AuthMgt:
		return "MGT"
	case Dot11AuthPsk:
		return "PSK"
	default:
		return "UNK"
	}
}

type CipherSuite struct {
	OUI  []byte // 3 bytes
	Type Dot11CipherType
}

type AuthSuite struct {
	OUI  []byte // 3 bytes
	Type Dot11AuthType
}

type CipherSuiteSelector struct {
	Count  uint16
	Suites []CipherSuite
}

type AuthSuiteSelector struct {
	Count  uint16
	Suites []AuthSuite
}

type RSNInfo struct {
	Version  uint16
	Group    CipherSuite
	Pairwise CipherSuiteSelector
	AuthKey  AuthSuiteSelector
}

type VendorInfo struct {
	WPAVersion uint16
	Multicast  CipherSuite
	Unicast    CipherSuiteSelector
	AuthKey    AuthSuiteSelector
}

func canParse(what string, buf []byte, need int) error {
	available := len(buf)
	if need > available {
		return fmt.Errorf("Malformed 802.11 packet, could not parse %s: needed %d bytes but only %d are available.", what, need, available)
	}
	return nil
}

func parsePairwiseSuite(buf []byte) (suite CipherSuite, err error) {
	if err = canParse("RSN.Pairwise.Suite", buf, 4); err == nil {
		suite.OUI = buf[0:3]
		suite.Type = Dot11CipherType(buf[3])
	}
	return
}

func parseAuthkeySuite(buf []byte) (suite AuthSuite, err error) {
	if err = canParse("RSN.AuthKey.Suite", buf, 4); err == nil {
		suite.OUI = buf[0:3]
		suite.Type = Dot11AuthType(buf[3])
	}
	return
}

func Dot11InformationElementVendorInfoDecode(buf []byte) (v VendorInfo, err error) {
	if err = canParse("Vendor", buf, 8); err == nil {
		v.WPAVersion = binary.LittleEndian.Uint16(buf[0:2])
		v.Multicast.OUI = buf[2:5]
		v.Multicast.Type = Dot11CipherType(buf[5])
		v.Unicast.Count = binary.LittleEndian.Uint16(buf[6:8])
		buf = buf[8:]
	} else {
		v.Unicast.Count = 0
		return
	}

	// check what we're left with
	if err = canParse("Vendor.Unicast.Suites", buf, int(v.Unicast.Count)*4); err == nil {
		for i := uint16(0); i < v.Unicast.Count; i++ {
			if suite, err := parsePairwiseSuite(buf); err == nil {
				v.Unicast.Suites = append(v.Unicast.Suites, suite)
				buf = buf[4:]
			} else {
				return v, err
			}
		}
	} else {
		v.Unicast.Count = 0
		return
	}

	if err = canParse("Vendor.AuthKey.Count", buf, 2); err == nil {
		v.AuthKey.Count = binary.LittleEndian.Uint16(buf[0:2])
		buf = buf[2:]
	} else {
		v.AuthKey.Count = 0
		return
	}

	// just like before, check if we have enough data
	if err = canParse("Vendor.AuthKey.Suites", buf, int(v.AuthKey.Count)*4); err == nil {
		for i := uint16(0); i < v.AuthKey.Count; i++ {
			if suite, err := parseAuthkeySuite(buf); err == nil {
				v.AuthKey.Suites = append(v.AuthKey.Suites, suite)
				buf = buf[4:]
			} else {
				return v, err
			}
		}
	} else {
		v.AuthKey.Count = 0
	}

	return
}

func Dot11InformationElementRSNInfoDecode(buf []byte) (rsn RSNInfo, err error) {
	if err = canParse("RSN", buf, 8); err == nil {
		rsn.Version = binary.LittleEndian.Uint16(buf[0:2])
		rsn.Group.OUI = buf[2:5]
		rsn.Group.Type = Dot11CipherType(buf[5])
		rsn.Pairwise.Count = binary.LittleEndian.Uint16(buf[6:8])
		buf = buf[8:]
	} else {
		rsn.Pairwise.Count = 0
		return
	}

	// check what we're left with
	if err = canParse("RSN.Pairwise.Suites", buf, int(rsn.Pairwise.Count)*4); err == nil {
		for i := uint16(0); i < rsn.Pairwise.Count; i++ {
			if suite, err := parsePairwiseSuite(buf); err == nil {
				rsn.Pairwise.Suites = append(rsn.Pairwise.Suites, suite)
				buf = buf[4:]
			} else {
				return rsn, err
			}
		}
	} else {
		rsn.Pairwise.Count = 0
		return
	}

	if err = canParse("RSN.AuthKey.Count", buf, 2); err == nil {
		rsn.AuthKey.Count = binary.LittleEndian.Uint16(buf[0:2])
		buf = buf[2:]
	} else {
		rsn.AuthKey.Count = 0
		return
	}

	// just like before, check if we have enough data
	if err = canParse("RSN.AuthKey.Suites", buf, int(rsn.AuthKey.Count)*4); err == nil {
		for i := uint16(0); i < rsn.AuthKey.Count; i++ {
			if suite, err := parseAuthkeySuite(buf); err == nil {
				rsn.AuthKey.Suites = append(rsn.AuthKey.Suites, suite)
				buf = buf[4:]
			} else {
				return rsn, err
			}
		}
	} else {
		rsn.AuthKey.Count = 0
	}

	return
}

func Dot11InformationElementIDDSSetDecode(buf []byte) (channel int, err error) {
	if err = canParse("DSSet.channel", buf, 1); err == nil {
		channel = int(buf[0])
	}

	return
}
