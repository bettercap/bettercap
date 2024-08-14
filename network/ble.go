//go:build !windows
// +build !windows

package network

import (
	"encoding/json"
	"sync"

	"github.com/evilsocket/islazy/data"
	"tinygo.org/x/bluetooth"
)

const BLEMacValidator = "([a-fA-F0-9:\\-]+)"

type BLEDevNewCallback func(dev *BLEDevice)
type BLEDevLostCallback func(dev *BLEDevice)

type BLE struct {
	sync.RWMutex
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
		devices: make(map[string]*BLEDevice),
		aliases: aliases,
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (b *BLE) MarshalJSON() ([]byte, error) {
	doc := bleJSON{
		Devices: b.Devices(),
	}
	return json.Marshal(doc)
}

func (b *BLE) Get(id string) (dev *BLEDevice, found bool) {
	b.RLock()
	defer b.RUnlock()

	dev, found = b.devices[id]
	return
}

func (b *BLE) AddIfNew(id string, scanResult bluetooth.ScanResult) *BLEDevice {
	b.Lock()
	defer b.Unlock()

	devAlias := b.aliases.GetOr(id, "")
	if dev, found := b.devices[id]; found {
		dev.Update(scanResult, devAlias)
		return dev
	}

	dev := NewBLEDevice(scanResult)
	dev.Update(scanResult, devAlias)

	b.devices[id] = dev

	if b.newCb != nil {
		b.newCb(dev)
	}

	return nil
}

func (b *BLE) Remove(id string) {
	b.Lock()
	defer b.Unlock()

	if dev, found := b.devices[id]; found {
		delete(b.devices, id)
		if b.lostCb != nil {
			b.lostCb(dev)
		}
	}
}

func (b *BLE) Devices() (devices []*BLEDevice) {
	b.Lock()
	defer b.Unlock()

	devices = make([]*BLEDevice, 0)
	for _, dev := range b.devices {
		devices = append(devices, dev)
	}
	return
}

func (b *BLE) EachDevice(cb func(mac string, d *BLEDevice)) {
	b.Lock()
	defer b.Unlock()

	for m, dev := range b.devices {
		cb(m, dev)
	}
}

func (b *BLE) Clear() {
	b.Lock()
	defer b.Unlock()
	b.devices = make(map[string]*BLEDevice)
}
