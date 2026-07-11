package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/evilsocket/islazy/data"
)

type AccessPoint struct {
	*Station
	sync.RWMutex

	aliases         *data.UnsortedKV
	clients         map[string]*Station
	withKeyMaterial bool
}

type apJSON struct {
	*Station
	Clients   []*Station `json:"clients"`
	Handshake bool       `json:"handshake"`
}

func NewAccessPoint(essid, bssid string, frequency int, rssi int8, aliases *data.UnsortedKV) *AccessPoint {
	return &AccessPoint{
		Station: NewStation(essid, bssid, frequency, rssi),
		aliases: aliases,
		clients: make(map[string]*Station),
	}
}

func (ap *AccessPoint) MarshalJSON() ([]byte, error) {
	ap.RLock()
	defer ap.RUnlock()

	// Station.MarshalJSON() (see wifi_station.go) exists to lock wpsMu
	// while reading the WPS field -- but apJSON embeds *Station
	// anonymously, and Go promotes an embedded field's MarshalJSON onto
	// the outer struct too. That means a naive json.Marshal(apJSON{...})
	// here would call ONLY Station.MarshalJSON() and silently drop
	// apJSON's own Clients/Handshake fields from the output entirely --
	// confirmed on-device: this broke pwnagotchi's own agent.py, which
	// expects every AP object to always have a "clients" key
	// (KeyError: 'clients'). Marshaling the station and the extra fields
	// separately, then splicing the two JSON objects together, keeps
	// both: the wpsMu-locked Station encoding via its own MarshalJSON,
	// and apJSON's additional fields, without either being silently
	// discarded by that promotion behavior.
	stationBytes, err := json.Marshal(ap.Station)
	if err != nil {
		return nil, err
	}

	extra := struct {
		Clients   []*Station `json:"clients"`
		Handshake bool       `json:"handshake"`
	}{
		Clients:   make([]*Station, 0, len(ap.clients)),
		Handshake: ap.withKeyMaterial,
	}
	for _, c := range ap.clients {
		extra.Clients = append(extra.Clients, c)
	}
	extraBytes, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}

	// both are guaranteed-well-formed JSON objects ("{...}") -- splice
	// them into one object by joining their inner contents with a comma
	if len(stationBytes) < 2 || len(extraBytes) < 2 {
		return stationBytes, nil
	}
	merged := make([]byte, 0, len(stationBytes)+len(extraBytes))
	merged = append(merged, stationBytes[:len(stationBytes)-1]...)
	merged = append(merged, ',')
	merged = append(merged, extraBytes[1:]...)
	return merged, nil
}

func (ap *AccessPoint) UnmarshalJSON(raw []byte) (err error) {
	ap.RLock()
	defer ap.RUnlock()

	var apData apJSON
	if err = json.Unmarshal(raw, &apData); err != nil {
		return
	}

	clients := make(map[string]*Station)
	for _, c := range apData.Clients {
		clients[c.HwAddress] = c
	}

	ap.Station = apData.Station
	ap.clients = clients
	ap.aliases, err = data.NewMemUnsortedKV()

	return
}

func (ap *AccessPoint) Get(bssid string) (*Station, bool) {
	ap.RLock()
	defer ap.RUnlock()

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
	delete(ap.clients, bssid)
}

func (ap *AccessPoint) AddClientIfNew(bssid string, frequency int, rssi int8) (*Station, bool) {
	ap.Lock()
	defer ap.Unlock()

	bssid = NormalizeMac(bssid)
	alias := ap.aliases.GetOr(bssid, "")

	if s, found := ap.clients[bssid]; found {
		// update
		s.Frequency = frequency
		s.RSSI = rssi
		s.LastSeen = time.Now()

		if alias != "" {
			s.Alias = alias
		}

		return s, false
	}

	s := NewStation("", bssid, frequency, rssi)
	s.Alias = alias
	ap.clients[bssid] = s

	return s, true
}

func (ap *AccessPoint) NumClients() int {
	ap.RLock()
	defer ap.RUnlock()
	return len(ap.clients)
}

func (ap *AccessPoint) Clients() (list []*Station) {
	ap.RLock()
	defer ap.RUnlock()

	list = make([]*Station, 0, len(ap.clients))
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
	ap.RLock()
	defer ap.RUnlock()

	return ap.withKeyMaterial
}

func (ap *AccessPoint) NumHandshakes() int {
	ap.RLock()
	defer ap.RUnlock()

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
	ap.RLock()
	defer ap.RUnlock()

	for _, c := range ap.clients {
		if c.Handshake.HasPMKID() {
			return true
		}
	}

	return false
}
