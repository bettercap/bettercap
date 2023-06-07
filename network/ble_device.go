// +build !windows

package network

import (
	"encoding/json"
	"time"

	"github.com/bettercap/gatt"
)

type BLECharacteristic struct {
	UUID       string      `json:"uuid"`
	Name       string      `json:"name"`
	Handle     uint16      `json:"handle"`
	Properties []string    `json:"properties"`
	Data       interface{} `json:"data"`
}

type BLEService struct {
	UUID            string              `json:"uuid"`
	Name            string              `json:"name"`
	Handle          uint16              `json:"handle"`
	EndHandle       uint16              `json:"end_handle"`
	Characteristics []BLECharacteristic `json:"characteristics"`
}

type BLEDevice struct {
	Alias         string
	LastSeen      time.Time
	DeviceName    string
	Vendor        string
	RSSI          int
	Device        gatt.Peripheral
	Advertisement *gatt.Advertisement
	Services      []BLEService
}

type bleDeviceJSON struct {
	LastSeen    time.Time    `json:"last_seen"`
	Name        string       `json:"name"`
	MAC         string       `json:"mac"`
	Alias       string       `json:"alias"`
	Vendor      string       `json:"vendor"`
	RSSI        int          `json:"rssi"`
	Connectable bool         `json:"connectable"`
	Flags       string       `json:"flags"`
	Services    []BLEService `json:"services"`
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
		Services:      make([]BLEService, 0),
	}
}

func (d *BLEDevice) Name() string {
	// get the name if it's being set during services enumeration via 'Device Name'
	name := d.DeviceName
	if name == "" {
		// get the name from the device
		name = d.Device.Name()
		if name == "" && d.Advertisement != nil {
			// get the name from the advertisement data
			name = d.Advertisement.LocalName
		}
	}
	return name
}

func (d *BLEDevice) MarshalJSON() ([]byte, error) {
	doc := bleDeviceJSON{
		LastSeen:    d.LastSeen,
		Name:        d.Name(),
		MAC:         d.Device.ID(),
		Alias:       d.Alias,
		Vendor:      d.Vendor,
		RSSI:        d.RSSI,
		Connectable: d.Advertisement.Connectable,
		Flags:       d.Advertisement.Flags.String(),
		Services:    d.Services,
	}
	return json.Marshal(doc)
}
