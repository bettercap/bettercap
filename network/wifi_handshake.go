package network

import (
	"github.com/google/gopacket"
	"sync"
)

type Handshake struct {
	sync.Mutex

	Challenges    []gopacket.Packet
	Responses     []gopacket.Packet
	Confirmations []gopacket.Packet
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

func (h *Handshake) Complete() bool {
	h.Lock()
	defer h.Unlock()

	nChal := len(h.Challenges)
	nResp := len(h.Responses)
	nConf := len(h.Confirmations)

	return nChal > 0 && nResp > 0 && nConf > 0
}

func (h *Handshake) NumUnsaved() int {
	h.Lock()
	defer h.Unlock()
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
