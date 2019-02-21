// +build windows darwin

package network

import (
	"encoding/json"
	"time"
)

type BLEDevice struct {
	LastSeen time.Time
}

func NewBLEDevice() *BLEDevice {
	return &BLEDevice{
		LastSeen: time.Now(),
	}
}

type BLEDevNewCallback func(dev *BLEDevice)
type BLEDevLostCallback func(dev *BLEDevice)

type BLE struct {
	devices map[string]*BLEDevice
	newCb   BLEDevNewCallback
	lostCb  BLEDevLostCallback
}

type bleJSON struct {
	Devices []*BLEDevice `json:"devices"`
}

func NewBLE(newcb BLEDevNewCallback, lostcb BLEDevLostCallback) *BLE {
	return &BLE{
		devices: make(map[string]*BLEDevice),
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (b *BLE) Get(id string) (dev *BLEDevice, found bool) {
	return
}

func (b *BLE) MarshalJSON() ([]byte, error) {
	doc := bleJSON{
		Devices: make([]*BLEDevice, 0),
	}
	return json.Marshal(doc)
}

func (b *BLE) EachDevice(cb func(mac string, d *BLEDevice)) {

}
