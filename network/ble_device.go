// +build !windows
// +build !darwin

package network

import (
	"encoding/json"
	"time"

	"github.com/bettercap/gatt"
)

type BLEDevice struct {
	LastSeen      time.Time
	Vendor        string
	RSSI          int
	Device        gatt.Peripheral
	Advertisement *gatt.Advertisement
}

type bleDeviceJSON struct {
	LastSeen time.Time `json:"last_seen"`
	Name     string    `json:"name"`
	MAC      string    `json:"mac"`
	Vendor   string    `json:"vendor"`
	RSSI     int       `json:"rssi"`
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

func (d *BLEDevice) MarshalJSON() ([]byte, error) {
	doc := bleDeviceJSON{
		LastSeen: d.LastSeen,
		Name:     d.Device.Name(),
		MAC:      d.Device.ID(),
		Vendor:   d.Vendor,
		RSSI:     d.RSSI,
	}

	return json.Marshal(doc)
}
