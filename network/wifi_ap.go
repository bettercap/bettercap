package network

import (
	"encoding/json"
	"sync"
	"time"
)

type AccessPoint struct {
	*Station
	sync.Mutex

	clients         map[string]*Station
	withKeyMaterial bool
}

type apJSON struct {
	*Station
	Clients []*Station `json:"clients"`
}

func NewAccessPoint(essid, bssid string, frequency int, rssi int8) *AccessPoint {
	return &AccessPoint{
		Station: NewStation(essid, bssid, frequency, rssi),
		clients: make(map[string]*Station),
	}
}

func (ap *AccessPoint) MarshalJSON() ([]byte, error) {
	ap.Lock()
	defer ap.Unlock()

	doc := apJSON{
		Station: ap.Station,
		Clients: make([]*Station, 0),
	}

	for _, c := range ap.clients {
		doc.Clients = append(doc.Clients, c)
	}

	return json.Marshal(doc)
}

func (ap *AccessPoint) Get(bssid string) (*Station, bool) {
	ap.Lock()
	defer ap.Unlock()

	bssid = NormalizeMac(bssid)
	if s, found := ap.clients[bssid]; found {
		return s, true
	}
	return nil, false
}

func (ap *AccessPoint) RemoveClient(mac string) {
	ap.Lock()
	defer ap.Unlock()

	bssid := NormalizeMac(mac)
	if _, found := ap.clients[bssid]; found {
		delete(ap.clients, bssid)
	}
}

func (ap *AccessPoint) AddClientIfNew(bssid string, frequency int, rssi int8) (*Station, bool) {
	ap.Lock()
	defer ap.Unlock()

	bssid = NormalizeMac(bssid)

	if s, found := ap.clients[bssid]; found {
		// update
		s.Frequency = frequency
		s.RSSI = rssi
		s.LastSeen = time.Now()

		return s, false
	}

	s := NewStation("", bssid, frequency, rssi)
	ap.clients[bssid] = s

	return s, true
}

func (ap *AccessPoint) NumClients() int {
	ap.Lock()
	defer ap.Unlock()
	return len(ap.clients)
}

func (ap *AccessPoint) Clients() (list []*Station) {
	ap.Lock()
	defer ap.Unlock()

	list = make([]*Station, 0)
	for _, c := range ap.clients {
		list = append(list, c)
	}
	return
}

func (ap *AccessPoint) EachClient(cb func(mac string, station *Station)) {
	ap.Lock()
	defer ap.Unlock()

	for m, station := range ap.clients {
		cb(m, station)
	}
}

func (ap *AccessPoint) WithKeyMaterial(state bool) {
	ap.Lock()
	defer ap.Unlock()

	ap.withKeyMaterial = state
}

func (ap *AccessPoint) HasKeyMaterial() bool {
	ap.Lock()
	defer ap.Unlock()

	return ap.withKeyMaterial
}

func (ap *AccessPoint) NumHandshakes() int {
	ap.Lock()
	defer ap.Unlock()

	sum := 0

	for _, c := range ap.clients {
		if c.Handshake.Complete() {
			sum++
		}
	}

	return sum
}

func (ap *AccessPoint) HasHandshakes() bool {
	return ap.NumHandshakes() > 0
}

func (ap *AccessPoint) HasPMKID() bool {
	ap.Lock()
	defer ap.Unlock()

	for _, c := range ap.clients {
		if c.Handshake.HasPMKID() {
			return true
		}
	}

	return false
}
