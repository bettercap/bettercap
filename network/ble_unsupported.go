// +build windows

package network

import (
	"encoding/json"
	"time"

	"github.com/evilsocket/islazy/data"
)

type BLEDevice struct {
	LastSeen time.Time
	Alias    string
}

func NewBLEDevice() *BLEDevice {
	return &BLEDevice{
		LastSeen: time.Now(),
	}
}

type BLEDevNewCallback func(dev *BLEDevice)
type BLEDevLostCallback func(dev *BLEDevice)

type BLE struct {
	aliases *data.UnsortedKV
	devices map[string]*BLEDevice
	newCb   BLEDevNewCallback
	lostCb  BLEDevLostCallback
}

type bleJSON struct {
	Devices []*BLEDevice `json:"devices"`
}

func NewBLE(aliases *data.UnsortedKV, newcb BLEDevNewCallback, lostcb BLEDevLostCallback) *BLE {
	return &BLE{
		aliases: aliases,
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
