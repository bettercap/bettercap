package network

import (
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

type Handshake struct {
	mu sync.RWMutex

	beacon        gopacket.Packet
	challenges    []gopacket.Packet
	responses     []gopacket.Packet
	confirmations []gopacket.Packet
	hasPMKID      bool
	unsaved       []gopacket.Packet
}

func NewHandshake() *Handshake {
	return &Handshake{
		challenges:    make([]gopacket.Packet, 0),
		responses:     make([]gopacket.Packet, 0),
		confirmations: make([]gopacket.Packet, 0),
		unsaved:       make([]gopacket.Packet, 0),
	}
}

func (h *Handshake) SetBeacon(pkt gopacket.Packet) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.beacon == nil {
		h.beacon = pkt
		h.unsaved = append(h.unsaved, pkt)
	}
}

// UpdateBeacon replaces the AP beacon without adding it to the station's
// unsaved handshake frames.
func (h *Handshake) UpdateBeacon(pkt gopacket.Packet) {
	h.mu.Lock()
	h.beacon = pkt
	h.mu.Unlock()
}

func (h *Handshake) Beacon() gopacket.Packet {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.beacon
}

func (h *Handshake) AddAndGetPMKID(pkt gopacket.Packet) []byte {
	h.AddFrame(0, pkt)

	prevWasKey := false
	for _, layer := range pkt.Layers() {
		if layer.LayerType() == layers.LayerTypeEAPOLKey {
			prevWasKey = true
			continue
		}

		if prevWasKey && layer.LayerType() == layers.LayerTypeDot11InformationElement {
			info := layer.(*layers.Dot11InformationElement)
			if info.ID == layers.Dot11InformationElementIDVendor && info.Length == 20 {
				h.mu.Lock()
				h.hasPMKID = true
				h.mu.Unlock()
				return info.Info
			}
		}

		prevWasKey = false
	}
	return nil
}

func (h *Handshake) AddFrame(n int, pkt gopacket.Packet) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch n {
	case 0:
		h.challenges = append(h.challenges, pkt)
	case 1:
		h.responses = append(h.responses, pkt)
	case 2:
		h.confirmations = append(h.confirmations, pkt)
	}
	h.unsaved = append(h.unsaved, pkt)
}

func (h *Handshake) AddExtra(pkt gopacket.Packet) {
	h.mu.Lock()
	h.unsaved = append(h.unsaved, pkt)
	h.mu.Unlock()
}

func (h *Handshake) Complete() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.challenges) > 0 && len(h.responses) > 0 && len(h.confirmations) > 0
}

func (h *Handshake) Half() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	nChal := len(h.challenges)
	nResp := len(h.responses)
	nConf := len(h.confirmations)
	return (nChal > 0 && nResp > 0) || (nResp > 0 && nConf > 0)
}

func (h *Handshake) HasPMKID() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.hasPMKID
}

func (h *Handshake) Any() bool {
	return h.HasPMKID() || h.Half() || h.Complete()
}

func (h *Handshake) NumUnsaved() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.unsaved)
}

func (h *Handshake) EachUnsavedPacket(cb func(gopacket.Packet)) {
	h.mu.Lock()
	packets := h.unsaved
	h.unsaved = make([]gopacket.Packet, 0)
	h.mu.Unlock()

	for _, pkt := range packets {
		cb(pkt)
	}
}
