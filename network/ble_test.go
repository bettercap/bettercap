//go:build !windows && !freebsd && !openbsd && !netbsd
// +build !windows,!freebsd,!openbsd,!netbsd

package network

import (
	"testing"

	"github.com/bettercap/gatt"
	"github.com/evilsocket/islazy/data"
)

type bleTestPeripheral struct {
	gatt.Peripheral
	id     string
	device gatt.Device
}

func (p *bleTestPeripheral) Device() gatt.Device {
	return p.device
}

func (p *bleTestPeripheral) ID() string {
	return p.id
}

func (p *bleTestPeripheral) Name() string {
	return "fake"
}

func TestBLEAddIfNewRefreshesPeripheral(t *testing.T) {
	aliases, err := data.NewUnsortedKV("", 0)
	if err != nil {
		t.Fatalf("Failed to create aliases store: %v", err)
	}

	ble := NewBLE(aliases, nil, nil)
	first := &bleTestPeripheral{id: "AA:BB:CC:DD:EE:FF"}
	second := &bleTestPeripheral{id: "AA:BB:CC:DD:EE:FF"}

	if dev := ble.AddIfNew(first.id, first, nil, -50); dev != nil {
		t.Fatalf("Expected nil for a new device, got %v", dev)
	}

	dev, found := ble.Get(first.id)
	if !found {
		t.Fatalf("Device %s was not stored", first.id)
	}
	if dev.Device != first {
		t.Errorf("Expected first peripheral %p, got %p", first, dev.Device)
	}

	if updated := ble.AddIfNew(second.id, second, nil, -40); updated != dev {
		t.Errorf("Expected the existing device %p, got %p", dev, updated)
	}
	if dev.Device != second {
		t.Errorf("Expected refreshed peripheral %p, got %p", second, dev.Device)
	}
	if dev.RSSI != -40 {
		t.Errorf("Expected RSSI -40, got %d", dev.RSSI)
	}
}
