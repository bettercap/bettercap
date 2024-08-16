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
	Frames      uint64
	Read        uint64
}

type canDeviceJSON struct {
	LastSeen    time.Time `json:"last_seen"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Frames      uint64    `json:"frames"`
	Read        uint64    `json:"read"`
}

func NewCANDevice(name string, description string, payload []byte) *CANDevice {
	dev := &CANDevice{
		LastSeen:    time.Now(),
		Name:        name,
		Description: description,
		Read:        uint64(len(payload)),
		Frames:      1,
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
		Frames:      dev.Frames,
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

	dev.Frames += 1
}
