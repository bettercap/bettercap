package network

import (
	"encoding/json"
	"sync"
	"time"
)

type CANDevice struct {
	sync.Mutex
	LastSeen    time.Time
	Name        string
	Description string
	Read        uint64
}

type canDeviceJSON struct {
	LastSeen    time.Time `json:"last_seen"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Read        uint64    `json:"description"`
}

func NewCANDevice(name string, description string, payload []byte) *CANDevice {
	dev := &CANDevice{
		LastSeen:    time.Now(),
		Name:        name,
		Description: description,
		Read:        uint64(len(payload)),
	}

	return dev
}

func (dev *CANDevice) MarshalJSON() ([]byte, error) {
	dev.Lock()
	defer dev.Unlock()

	doc := canDeviceJSON{
		LastSeen:    dev.LastSeen,
		Name:        dev.Name,
		Description: dev.Description,
		Read:        dev.Read,
	}

	return json.Marshal(doc)
}

func (dev *CANDevice) AddPayload(payload []byte) {
	dev.Lock()
	defer dev.Unlock()

	sz := len(payload)
	if payload != nil && sz > 0 {
		dev.Read += uint64(sz)
	}
}
