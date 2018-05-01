package network

import (
	"net"
	"testing"
)

func TestIsZeroMac(t *testing.T) {
	exampleMAC, _ := net.ParseMAC("00:00:00:00:00:00")

	exp := true
	got := IsZeroMac(exampleMAC)
	if got != exp {
		t.Fatalf("expected '%t', got '%t'", exp, got)
	}
}

func TestIsBroadcastMac(t *testing.T) {
	exampleMAC, _ := net.ParseMAC("ff:ff:ff:ff:ff:ff")

	exp := true
	got := IsBroadcastMac(exampleMAC)
	if got != exp {
		t.Fatalf("expected '%t', got '%t'", exp, got)
	}
}

func TestNormalizeMac(t *testing.T) {
	exp := "ff:ff:ff:ff:ff:ff"
	got := NormalizeMac("fF-fF-fF-fF-fF-fF")
	if got != exp {
		t.Fatalf("expected '%s', got '%s'", exp, got)
	}
}

// TODO: refactor to parse targets with an actual alias map
func TestParseTargets(t *testing.T) {
	ips, macs, err := ParseTargets("192.168.1.2, 192.168.1.3", &Aliases{})
	if err != nil {
		t.Error("ips:", ips, "macs:", macs, "err:", err)
	}
	if len(ips) != 2 {
		t.Fatalf("expected '%d', got '%d'", 2, len(ips))
	}
	if len(macs) != 0 {
		t.Fatalf("expected '%d', got '%d'", 0, len(macs))
	}
}

func TestBuildEndpointFromInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
	}
	if len(ifaces) <= 0 {
		t.Error("Unable to find any network interfaces to run test with.")
	}
	_, err = buildEndpointFromInterface(ifaces[0])
	if err != nil {
		t.Error(err)
	}
}
