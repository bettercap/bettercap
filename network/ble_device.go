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
	vendor := ManufLookup(NormalizeMac(p.ID()))
	if vendor == "" && a != nil {
		vendor = a.Company
	}
	return &BLEDevice{
		LastSeen:      time.Now(),
		Device:        p,
		Vendor:        vendor,
		Advertisement: a,
		RSSI:          rssi,
	}
}

func (d *BLEDevice) Name() string {
	name := d.Device.Name()
	if name == "" && d.Advertisement != nil {
		name = d.Advertisement.LocalName
	}
	return name
}

func (d *BLEDevice) MarshalJSON() ([]byte, error) {
	doc := bleDeviceJSON{
		LastSeen: d.LastSeen,
		Name:     d.Name(),
		MAC:      d.Device.ID(),
		Vendor:   d.Vendor,
		RSSI:     d.RSSI,
	}
	return json.Marshal(doc)
}
