package packets

import (
	"github.com/mdlayher/dhcp6"
	"testing"
)

func TestDHCP6OptDNSServers(t *testing.T) {
	exp := 23
	got := DHCP6OptDNSServers
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestDHCP6OptDNSDomains(t *testing.T) {
	exp := 24
	got := DHCP6OptDNSDomains
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestDHCP6OptClientFQDN(t *testing.T) {
	exp := 39
	got := DHCP6OptClientFQDN
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestIPv6Prefix(t *testing.T) {
	exp := "fe80::"
	got := IPv6Prefix
	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestDHCP6EncodeList(t *testing.T) {
	domains := []string{"microsoft.com", "google.com", "facebook.com"}

	encoded := DHCP6EncodeList(domains)

	if len(encoded) <= 0 {
		t.Error("unable to dhcp6 encode a given list")
	}
}

func TestDHCP6For(t *testing.T) {
	mesg := dhcp6.MessageTypeSolicit
	pakt := dhcp6.Packet{Options: dhcp6.Options{}}
	pakt.Options.AddRaw(dhcp6.OptionClientID, []byte{})
	duid := []byte{}

	err, _ := DHCP6For(mesg, pakt, duid)

	if err != nil {
		t.Error(err)
	}
}
