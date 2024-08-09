package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/evilsocket/islazy/data"
)

type CANDevNewCallback func(dev *CANDevice)
type CANDevLostCallback func(dev *CANDevice)

type CAN struct {
	sync.RWMutex
	devices map[string]*CANDevice
	newCb   CANDevNewCallback
	lostCb  CANDevLostCallback
}

type canJSON struct {
	Devices []*CANDevice `json:"devices"`
}

func NewCAN(aliases *data.UnsortedKV, newcb CANDevNewCallback, lostcb CANDevLostCallback) *CAN {
	return &CAN{
		devices: make(map[string]*CANDevice),
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (h *CAN) MarshalJSON() ([]byte, error) {
	doc := canJSON{
		Devices: make([]*CANDevice, 0),
	}

	for _, dev := range h.devices {
		doc.Devices = append(doc.Devices, dev)
	}

	return json.Marshal(doc)
}

func (b *CAN) Get(id string) (dev *CANDevice, found bool) {
	b.RLock()
	defer b.RUnlock()
	dev, found = b.devices[id]
	return
}

func (b *CAN) AddIfNew(name string, description string, payload []byte) (bool, *CANDevice) {
	b.Lock()
	defer b.Unlock()

	id := name

	if dev, found := b.devices[id]; found {
		dev.LastSeen = time.Now()
		dev.AddPayload(payload)
		return false, dev
	}

	newDev := NewCANDevice(name, description, payload)
	b.devices[id] = newDev

	if b.newCb != nil {
		b.newCb(newDev)
	}

	return true, newDev
}

func (b *CAN) Remove(id string) {
	b.Lock()
	defer b.Unlock()

	if dev, found := b.devices[id]; found {
		delete(b.devices, id)
		if b.lostCb != nil {
			b.lostCb(dev)
		}
	}
}

func (b *CAN) Devices() (devices []*CANDevice) {
	b.Lock()
	defer b.Unlock()

	devices = make([]*CANDevice, 0)
	for _, dev := range b.devices {
		devices = append(devices, dev)
	}
	return
}

func (b *CAN) EachDevice(cb func(mac string, d *CANDevice)) {
	b.Lock()
	defer b.Unlock()

	for m, dev := range b.devices {
		cb(m, dev)
	}
}

func (b *CAN) Clear() {
	b.Lock()
	defer b.Unlock()
	b.devices = make(map[string]*CANDevice)
}
