package packets

import (
	"net"
	"reflect"
	"testing"
)

func TestNewARPTo(t *testing.T) {
	from := net.IP{0, 0, 0, 0}
	from_hw, _ := net.ParseMAC("01:23:45:67:89:ab")
	to := net.IP{0, 0, 0, 0}
	to_hw, _ := net.ParseMAC("01:23:45:67:89:ab")
	req := uint16(0)

	eth, arp := NewARPTo(from, from_hw, to, to_hw, req)

	if !reflect.DeepEqual(eth.SrcMAC, from_hw) {
		t.Fatalf("expected '%s', got '%s'", eth.SrcMAC, from_hw)
	}

	if !reflect.DeepEqual(eth.DstMAC, to_hw) {
		t.Fatalf("expected '%s', got '%s'", eth.DstMAC, to_hw)
	}

	if !reflect.DeepEqual(arp.Operation, req) {
		t.Fatalf("expected '%d', got '%d'", arp.Operation, req)
	}

	if !reflect.DeepEqual(arp.SourceHwAddress, []byte(from_hw)) {
		t.Fatalf("expected '%v', got '%v'", arp.SourceHwAddress, []byte(from_hw))
	}

	if !reflect.DeepEqual(arp.DstHwAddress, []byte(to_hw)) {
		t.Fatalf("expected '%v', got '%v'", arp.DstHwAddress, []byte(to_hw))
	}
}

func TestNewARP(t *testing.T) {
	from := net.IP{0, 0, 0, 0}
	from_hw, _ := net.ParseMAC("01:23:45:67:89:ab")
	to_hw, _ := net.ParseMAC("00:00:00:00:00:00")
	to := net.IP{0, 0, 0, 0}
	req := uint16(0)

	eth, arp := NewARP(from, from_hw, to, req)

	if !reflect.DeepEqual(eth.SrcMAC, from_hw) {
		t.Fatalf("expected '%s', got '%s'", eth.SrcMAC, from_hw)
	}

	if !reflect.DeepEqual(eth.DstMAC, to_hw) {
		t.Fatalf("expected '%s', got '%s'", eth.DstMAC, to_hw)
	}

	if !reflect.DeepEqual(arp.Operation, req) {
		t.Fatalf("expected '%d', got '%d'", arp.Operation, req)
	}

	if !reflect.DeepEqual(arp.SourceHwAddress, []byte(from_hw)) {
		t.Fatalf("expected '%v', got '%v'", arp.SourceHwAddress, []byte(from_hw))
	}

	if !reflect.DeepEqual(arp.DstHwAddress, []byte{0, 0, 0, 0, 0, 0}) {
		t.Fatalf("expected '%v', got '%v'", arp.DstHwAddress, []byte{0, 0, 0, 0, 0, 0})
	}
}

func TestNewARPRequest(t *testing.T) {
	from := net.IP{0, 0, 0, 0}
	from_hw, _ := net.ParseMAC("01:23:45:67:89:ab")
	to := net.IP{0, 0, 0, 0}

	err, bytes := NewARPRequest(from, from_hw, to)
	if err != nil {
		t.Error(err)
	}

	if len(bytes) <= 0 {
		t.Error("unable to serialize new arp request packet")
	}
}

func TestNewARPReply(t *testing.T) {
	from := net.IP{0, 0, 0, 0}
	from_hw, _ := net.ParseMAC("01:23:45:67:89:ab")
	to := net.IP{0, 0, 0, 0}
	to_hw, _ := net.ParseMAC("01:23:45:67:89:ab")

	err, bytes := NewARPReply(from, from_hw, to, to_hw)
	if err != nil {
		t.Error(err)
	}

	if len(bytes) <= 0 {
		t.Error("unable to serialize new arp request packet")
	}
}
