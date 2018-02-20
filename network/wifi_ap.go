package network

import (
	"encoding/json"
	"sync"
	"time"
)

type AccessPoint struct {
	*Station
	sync.Mutex

	clients map[string]*Station
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
	if s, found := ap.clients[bssid]; found == true {
		return s, true
	}
	return nil, false
}

func (ap *AccessPoint) AddClient(bssid string, frequency int, rssi int8) *Station {
	ap.Lock()
	defer ap.Unlock()

	bssid = NormalizeMac(bssid)

	if s, found := ap.clients[bssid]; found == true {
		// update
		s.Frequency = frequency
		s.RSSI = rssi
		s.LastSeen = time.Now()

		return s
	}

	s := NewStation("", bssid, frequency, rssi)
	ap.clients[bssid] = s

	return s
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
