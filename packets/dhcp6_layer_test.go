package packets

import (
	"github.com/google/gopacket"
	"testing"
)

func TestDHCPv6Layer(t *testing.T) {
	layer := DHCPv6Layer{}

	exp := 0
	got := len(layer.Raw)

	if exp != got {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestDHCP6SerializeTo(t *testing.T) {
	layer := DHCPv6Layer{}
	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}

	got := layer.SerializeTo(buffer, opts)

	if got != nil {
		t.Fatalf("expected '%v', got '%v'", nil, got)
	}
}
