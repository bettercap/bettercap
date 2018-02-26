package network

import (
	"encoding/json"
)

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

func (b *BLE) MarshalJSON() ([]byte, error) {
	doc := bleJSON{
		Devices: make([]*BLEDevice, 0),
	}

	for _, dev := range b.Devices() {
		doc.Devices = append(doc.Devices, dev)
	}

	return json.Marshal(doc)
}
