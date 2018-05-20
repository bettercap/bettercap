package packets

import (
	"reflect"
	"testing"
)

func TestDot11CipherTypes(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{uint8(Dot11CipherWep), uint8(1)},
		{uint8(Dot11CipherTkip), uint8(2)},
		{uint8(Dot11CipherWrap), uint8(3)},
		{uint8(Dot11CipherCcmp), uint8(4)},
		{uint8(Dot11CipherWep104), uint8(5)},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11CipherTypesAsString(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{Dot11CipherWep.String(), "WEP"},
		{Dot11CipherTkip.String(), "TKIP"},
		{Dot11CipherWrap.String(), "WRAP"},
		{Dot11CipherCcmp.String(), "CCMP"},
		{Dot11CipherWep104.String(), "WEP104"},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11AuthTypes(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{uint8(Dot11AuthMgt), uint8(1)},
		{uint8(Dot11AuthPsk), uint8(2)},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11AuthTypesAsString(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{Dot11AuthMgt.String(), "MGT"},
		{Dot11AuthPsk.String(), "PSK"},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11CipherSuite(t *testing.T) {
	// must be three, but not currently
	// implemented to really enforce [3]byte
	bytes := []byte{1, 2, 3}
	cs := CipherSuite{
		OUI: bytes,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{cs.OUI, bytes},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11AuthSuite(t *testing.T) {
	// must be three, but not currently
	// implemented to really enforce [3]byte
	bytes := []byte{1, 2, 3}
	cs := AuthSuite{
		OUI: bytes,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{cs.OUI, bytes},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11CipherSuiteSelector(t *testing.T) {
	count := uint16(1)
	cs := CipherSuiteSelector{
		Count: count,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{cs.Count, count},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11AuthSuiteSelector(t *testing.T) {
	count := uint16(1)
	cs := AuthSuiteSelector{
		Count: count,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{cs.Count, count},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11RSNInfo(t *testing.T) {
	version := uint16(1)
	rsn := RSNInfo{
		Version: version,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{rsn.Version, version},
		{rsn.Group, CipherSuite{}},
		{rsn.Pairwise, CipherSuiteSelector{}},
		{rsn.AuthKey, AuthSuiteSelector{}},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11VendorInfo(t *testing.T) {
	version := uint16(1)
	vendor := VendorInfo{
		WPAVersion: version,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{vendor.WPAVersion, version},
		{vendor.Multicast, CipherSuite{}},
		{vendor.Unicast, CipherSuiteSelector{}},
		{vendor.AuthKey, AuthSuiteSelector{}},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11canParse(t *testing.T) {
	err := canParse("example", []byte{}, 0)
	if err != nil {
		t.Error("unable to check if able to parse")
	}
}

func TestDot11parsePairwiseSuite(t *testing.T) {
	buf := []byte{0, 0, 1, 1}
	suite, err := parsePairwiseSuite(buf)
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{suite.OUI, buf[0:3]},
		{suite.Type, Dot11CipherType(buf[3])},
		{err, nil},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11parseAuthkeySuite(t *testing.T) {
	buf := []byte{0, 0, 1, 1}
	suite, err := parseAuthkeySuite(buf)
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{suite.OUI, buf[0:3]},
		{suite.Type, Dot11AuthType(buf[3])},
		{err, nil},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

// TODO: add test for Dot11InformationElementVendorInfoDecode
// TODO: add test for Dot11InformationElementIDDSSetDecode
