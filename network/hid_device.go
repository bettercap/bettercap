package network

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/evilsocket/islazy/str"
)

type HIDType int

const (
	HIDTypeUnknown   HIDType = 0
	HIDTypeLogitech  HIDType = 1
	HIDTypeAmazon    HIDType = 2
	HIDTypeMicrosoft HIDType = 3
	HIDTypeDell      HIDType = 4
)

func (t HIDType) String() string {
	switch t {
	case HIDTypeLogitech:
		return "Logitech"
	case HIDTypeAmazon:
		return "Amazon"
	case HIDTypeMicrosoft:
		return "Microsoft"
	case HIDTypeDell:
		return "Dell"
	}
	return ""
}

type HIDPayload []byte

type HIDDevice struct {
	sync.Mutex
	LastSeen   time.Time
	Type       HIDType
	Alias      string
	Address    string
	RawAddress []byte
	channels   map[int]bool
	payloads   []HIDPayload
	payloadsSz uint64
}

type hidDeviceJSON struct {
	LastSeen     time.Time `json:"last_seen"`
	Type         string    `json:"type"`
	Address      string    `json:"address"`
	Alias        string    `json:"alias"`
	Channels     []string  `json:"channels"`
	Payloads     []string  `json:"payloads"`
	PayloadsSize uint64    `json:"payloads_size"`
}

func NormalizeHIDAddress(address string) string {
	parts := strings.Split(address, ":")
	for i, p := range parts {
		if len(p) < 2 {
			parts[i] = "0" + p
		}
	}
	return strings.ToLower(strings.Join(parts, ":"))

}

func HIDAddress(address []byte) string {
	octects := []string{}
	for _, b := range address {
		octects = append(octects, fmt.Sprintf("%02x", b))
	}
	return strings.ToLower(strings.Join(octects, ":"))
}

func NewHIDDevice(address []byte, channel int, payload []byte) *HIDDevice {
	dev := &HIDDevice{
		LastSeen:   time.Now(),
		Type:       HIDTypeUnknown,
		RawAddress: address,
		Address:    HIDAddress(address),
		channels:   make(map[int]bool),
		payloads:   make([]HIDPayload, 0),
		payloadsSz: 0,
	}

	dev.AddChannel(channel)
	dev.AddPayload(payload)

	return dev
}

func (dev *HIDDevice) MarshalJSON() ([]byte, error) {
	dev.Lock()
	defer dev.Unlock()

	doc := hidDeviceJSON{
		LastSeen:     dev.LastSeen,
		Type:         dev.Type.String(),
		Address:      dev.Address,
		Alias:        dev.Alias,
		Channels:     dev.channelsListUnlocked(),
		Payloads:     make([]string, 0),
		PayloadsSize: dev.payloadsSz,
	}

	// get the latest 50 payloads
	for i := len(dev.payloads) - 1; i >= 0; i-- {
		data := str.Trim(hex.Dump(dev.payloads[i]))
		doc.Payloads = append([]string{data}, doc.Payloads...)
		if len(doc.Payloads) == 50 {
			break
		}
	}

	return json.Marshal(doc)
}

func (dev *HIDDevice) AddChannel(ch int) {
	dev.Lock()
	defer dev.Unlock()

	dev.channels[ch] = true
}

func (dev *HIDDevice) channelsListUnlocked() []string {
	chans := []string{}
	for ch := range dev.channels {
		chans = append(chans, fmt.Sprintf("%d", ch))
	}

	sort.Strings(chans)

	return chans
}
func (dev *HIDDevice) ChannelsList() []string {
	dev.Lock()
	defer dev.Unlock()
	return dev.channelsListUnlocked()
}

func (dev *HIDDevice) Channels() string {
	return strings.Join(dev.ChannelsList(), ",")
}

// credits to https://github.com/insecurityofthings/jackit/tree/master/jackit/plugins
func (dev *HIDDevice) onEventFrame(p []byte, sz int) {
	// return if type has been already determined
	if dev.Type != HIDTypeUnknown {
		return
	}

	if sz == 6 {
		dev.Type = HIDTypeAmazon
		return
	}

	if sz == 10 && p[0] == 0x00 && p[1] == 0xc2 {
		dev.Type = HIDTypeLogitech // mouse movement
		return
	} else if sz == 22 && p[0] == 0x00 && p[1] == 0xd3 {
		dev.Type = HIDTypeLogitech // keystroke
		return
	} else if sz == 5 && p[0] == 0x00 && p[1] == 0x40 {
		dev.Type = HIDTypeLogitech // keepalive
		return
		// TODO: review this condition
	} else if sz == 10 && p[0] == 0x00 { //&& p[0] == 0x4f {
		dev.Type = HIDTypeLogitech // sleep timer
		return
	}

	if sz == 19 && (p[0] == 0x08 || p[0] == 0x0c) && p[6] == 0x40 {
		dev.Type = HIDTypeMicrosoft
		return
	}

	// TODO: Dell
}

func (dev *HIDDevice) AddPayload(payload []byte) {
	dev.Lock()
	defer dev.Unlock()

	sz := len(payload)
	if payload != nil && sz > 0 {
		dev.payloads = append(dev.payloads, payload)
		dev.payloadsSz += uint64(sz)
		dev.onEventFrame(payload, sz)
	}
}

func (dev *HIDDevice) EachPayload(cb func([]byte) bool) {
	dev.Lock()
	defer dev.Unlock()

	for _, payload := range dev.payloads {
		if done := cb(payload); done {
			break
		}
	}
}

func (dev *HIDDevice) NumPayloads() int {
	dev.Lock()
	defer dev.Unlock()
	return len(dev.payloads)
}

func (dev *HIDDevice) PayloadsSize() uint64 {
	dev.Lock()
	defer dev.Unlock()
	return dev.payloadsSz
}
