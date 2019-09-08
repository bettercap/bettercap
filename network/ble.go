// +build !windows

package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bettercap/gatt"

	"github.com/evilsocket/islazy/data"
)

const BLEMacValidator = "([a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2})"

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

func (b *BLE) AddIfNew(id string, p gatt.Peripheral, a *gatt.Advertisement, rssi int) *BLEDevice {
	b.Lock()
	defer b.Unlock()

	id = NormalizeMac(id)
	alias := b.aliases.GetOr(id, "")
	if dev, found := b.devices[id]; found {
		dev.LastSeen = time.Now()
		dev.RSSI = rssi
		dev.Advertisement = a
		if alias != "" {
			dev.Alias = alias
		}
		return dev
	}

	newDev := NewBLEDevice(p, a, rssi)
	newDev.Alias = alias
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
