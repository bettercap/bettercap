// +build !windows
// +build !darwin

package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bettercap/gatt"
)

type BLEDevNewCallback func(dev *BLEDevice)
type BLEDevLostCallback func(dev *BLEDevice)

type BLE struct {
	sync.RWMutex
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

func (b *BLE) Get(id string) (dev *BLEDevice, found bool) {
	b.RLock()
	defer b.RUnlock()

	dev, found = b.devices[id]
	return
}

func (b *BLE) AddIfNew(id string, p gatt.Peripheral, a *gatt.Advertisement, rssi int) *BLEDevice {
	b.Lock()
	defer b.Unlock()

	id = NormalizeMac(id)
	if dev, found := b.devices[id]; found == true {
		dev.LastSeen = time.Now()
		dev.RSSI = rssi
		dev.Advertisement = a
		return dev
	}

	newDev := NewBLEDevice(p, a, rssi)
	b.devices[id] = newDev

	if b.newCb != nil {
		b.newCb(newDev)
	}

	return nil
}

func (b *BLE) Remove(id string) {
	b.Lock()
	defer b.Unlock()

	id = NormalizeMac(id)
	if dev, found := b.devices[id]; found == true {
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
