package network

import (
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Handshake struct {
	sync.RWMutex

	Beacon        gopacket.Packet
	Challenges    []gopacket.Packet
	Responses     []gopacket.Packet
	Confirmations []gopacket.Packet
	hasPMKID      bool
	unsaved       []gopacket.Packet
}

func NewHandshake() *Handshake {
	return &Handshake{
		Challenges:    make([]gopacket.Packet, 0),
		Responses:     make([]gopacket.Packet, 0),
		Confirmations: make([]gopacket.Packet, 0),
		unsaved:       make([]gopacket.Packet, 0),
	}
}

func (h *Handshake) SetBeacon(pkt gopacket.Packet) {
	h.Lock()
	defer h.Unlock()

	if h.Beacon == nil {
		h.Beacon = pkt
		h.unsaved = append(h.unsaved, pkt)
	}
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
				h.Lock()
				defer h.Unlock()
				h.hasPMKID = true
				return info.Info
			}
		}

		prevWasKey = false
	}

	return nil
}

func (h *Handshake) AddFrame(n int, pkt gopacket.Packet) {
	h.Lock()
	defer h.Unlock()

	switch n {
	case 0:
		h.Challenges = append(h.Challenges, pkt)
	case 1:
		h.Responses = append(h.Responses, pkt)
	case 2:
		h.Confirmations = append(h.Confirmations, pkt)
	}

	h.unsaved = append(h.unsaved, pkt)
}

func (h *Handshake) AddExtra(pkt gopacket.Packet) {
	h.Lock()
	defer h.Unlock()
	h.unsaved = append(h.unsaved, pkt)
}

func (h *Handshake) Complete() bool {
	h.RLock()
	defer h.RUnlock()

	nChal := len(h.Challenges)
	nResp := len(h.Responses)
	nConf := len(h.Confirmations)

	return nChal > 0 && nResp > 0 && nConf > 0
}

func (h *Handshake) Half() bool {
	h.RLock()
	defer h.RUnlock()

	/*
	 * You can use every combination of the handshake to crack the net:
	 * M1/M2
	 * M2/M3
	 * M3/M4
	 * M1/M4 (if M4 snonce is not zero)
	 * We only have M1 (the challenge), M2 (the response) and M3 (the confirmation)
	 */
	nChal := len(h.Challenges)
	nResp := len(h.Responses)
	nConf := len(h.Confirmations)

	return (nChal > 0 && nResp > 0) || (nResp > 0 && nConf > 0)
}

func (h *Handshake) HasPMKID() bool {
	h.RLock()
	defer h.RUnlock()
	return h.hasPMKID
}

func (h *Handshake) Any() bool {
	return h.HasPMKID() || h.Half() || h.Complete()
}

func (h *Handshake) NumUnsaved() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.unsaved)
}

func (h *Handshake) EachUnsavedPacket(cb func(gopacket.Packet)) {
	h.Lock()
	defer h.Unlock()

	for _, pkt := range h.unsaved {
		cb(pkt)
	}
	h.unsaved = make([]gopacket.Packet, 0)
}
