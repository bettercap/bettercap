package network

import (
	"encoding/json"
	"sync"
	"time"
)

type StationNewCallback func(s *Station)
type StationLostCallback func(s *Station)

type WiFi struct {
	sync.Mutex

	stations map[string]*Station
	iface    *Endpoint
	newCb    StationNewCallback
	lostCb   StationLostCallback
}

type wifiJSON struct {
	Stations []*Station `json:"stations"`
}

func NewWiFi(iface *Endpoint, newcb StationNewCallback, lostcb StationLostCallback) *WiFi {
	return &WiFi{
		stations: make(map[string]*Station),
		iface:    iface,
		newCb:    newcb,
		lostCb:   lostcb,
	}
}

func (w *WiFi) MarshalJSON() ([]byte, error) {
	doc := wifiJSON{
		Stations: make([]*Station, 0),
	}

	for _, s := range w.stations {
		doc.Stations = append(doc.Stations, s)
	}

	return json.Marshal(doc)
}

func (w *WiFi) List() (list []*Station) {
	w.Lock()
	defer w.Unlock()

	list = make([]*Station, 0)
	for _, t := range w.stations {
		list = append(list, t)
	}
	return
}

func (w *WiFi) Remove(mac string) {
	w.Lock()
	defer w.Unlock()

	if s, found := w.stations[mac]; found {
		delete(w.stations, mac)
		if w.lostCb != nil {
			w.lostCb(s)
		}
	}
}

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, frequency int, rssi int8) *Station {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	if station, found := w.stations[mac]; found {
		station.LastSeen = time.Now()
		station.RSSI = rssi
		return station
	}

	newStation := NewStation(ssid, mac, isAp, frequency, rssi)
	w.stations[mac] = newStation

	if w.newCb != nil {
		w.newCb(newStation)
	}

	return nil
}

func (w *WiFi) Get(mac string) (*Station, bool) {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	station, found := w.stations[mac]
	return station, found
}

func (w *WiFi) Clear() error {
	w.stations = make(map[string]*Station)
	return nil
}
