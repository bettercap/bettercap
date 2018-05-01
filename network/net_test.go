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
