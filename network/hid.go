package network

import (
	"sync"
	"time"
)

type HIDDevNewCallback func(dev *HIDDevice)
type HIDDevLostCallback func(dev *HIDDevice)

type HID struct {
	sync.RWMutex
	devices map[string]*HIDDevice
	newCb   HIDDevNewCallback
	lostCb  HIDDevLostCallback
}

func NewHID(newcb HIDDevNewCallback, lostcb HIDDevLostCallback) *HID {
	return &HID{
		devices: make(map[string]*HIDDevice),
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (b *HID) Get(id string) (dev *HIDDevice, found bool) {
	b.RLock()
	defer b.RUnlock()
	dev, found = b.devices[id]
	return
}

func (b *HID) AddIfNew(address []byte, channel int, payload []byte) (bool, *HIDDevice) {
	b.Lock()
	defer b.Unlock()

	id := HIDAddress(address)
	if dev, found := b.devices[id]; found {
		dev.LastSeen = time.Now()
		dev.AddChannel(channel)
		dev.AddPayload(payload)
		return false, dev
	}

	newDev := NewHIDDevice(address, channel, payload)
	b.devices[id] = newDev

	if b.newCb != nil {
		b.newCb(newDev)
	}

	return true, newDev
}

func (b *HID) Remove(id string) {
	b.Lock()
	defer b.Unlock()

	if dev, found := b.devices[id]; found {
		delete(b.devices, id)
		if b.lostCb != nil {
			b.lostCb(dev)
		}
	}
}

func (b *HID) Devices() (devices []*HIDDevice) {
	b.Lock()
	defer b.Unlock()

	devices = make([]*HIDDevice, 0)
	for _, dev := range b.devices {
		devices = append(devices, dev)
	}
	return
}

func (b *HID) EachDevice(cb func(mac string, d *HIDDevice)) {
	b.Lock()
	defer b.Unlock()

	for m, dev := range b.devices {
		cb(m, dev)
	}
}

func (b *HID) Clear() {
	b.Lock()
	defer b.Unlock()
	b.devices = make(map[string]*HIDDevice)
}
