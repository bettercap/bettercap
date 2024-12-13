package network

import (
	"encoding/json"
	"sync"
	"time"
)

// CANDevice represents a CAN (Controller Area Network) device with activity tracking.
type CANDevice struct {
	sync.Mutex
	LastSeen    time.Time // Timestamp of the last activity.
	Name        string    // Name of the device.
	Description string    // Description of the device.
	Frames      uint64    // Number of frames sent/received.
	Read        uint64    // Total bytes read.
}

type canDeviceJSON struct {
	LastSeen    time.Time `json:"last_seen"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Frames      uint64    `json:"frames"`
	Read        uint64    `json:"read"`
}

// NewCANDevice initializes a new CANDevice with the provided name, description, and payload.
func NewCANDevice(name, description string, payload []byte) *CANDevice {
	return &CANDevice{
		LastSeen:    time.Now(),
		Name:        name,
		Description: description,
		Frames:      1,
		Read:        uint64(len(payload)),
	}
}

// MarshalJSON customizes the JSON serialization of CANDevice.
func (dev *CANDevice) MarshalJSON() ([]byte, error) {
	dev.Lock()
	defer dev.Unlock()

	doc := canDeviceJSON{
		LastSeen:    dev.LastSeen,
		Name:        dev.Name,
		Description: dev.Description,
		Frames:      dev.Frames,
		Read:        dev.Read,
	}

	return json.Marshal(doc)
}

// AddPayload updates the CANDevice's statistics with the new payload data.
func (dev *CANDevice) AddPayload(payload []byte) {
	dev.Lock()
	defer dev.Unlock()

	if len(payload) > 0 {
		dev.Read += uint64(len(payload))
	}
	dev.Frames++
}
