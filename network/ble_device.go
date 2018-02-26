// +build !windows

package network

import (
	"time"

	"github.com/currantlabs/gatt"
)

type BLEDevice struct {
	LastSeen      time.Time
	Device        gatt.Peripheral
	Vendor        string
	Advertisement *gatt.Advertisement
	RSSI          int
}

func NewBLEDevice(p gatt.Peripheral, a *gatt.Advertisement, rssi int) *BLEDevice {
	return &BLEDevice{
		LastSeen:      time.Now(),
		Device:        p,
		Vendor:        OuiLookup(NormalizeMac(p.ID())),
		Advertisement: a,
		RSSI:          rssi,
	}
}
