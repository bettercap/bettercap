package network

import (
	"sync"
	"time"
)

type StationNewCallback func(s *Station)
type StationLostCallback func(s *Station)

type WiFi struct {
	sync.Mutex
	iface    *Endpoint
	stations map[string]*Station
	newCb    StationNewCallback
	lostCb   StationLostCallback
}

func NewWiFi(iface *Endpoint, newcb StationNewCallback, lostcb StationLostCallback) *WiFi {
	return &WiFi{
		iface:    iface,
		stations: make(map[string]*Station),
		newCb:    newcb,
		lostCb:   lostcb,
	}
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

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, channel int, rssi int8) *Station {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	if station, found := w.stations[mac]; found {
		station.LastSeen = time.Now()
		station.RSSI = rssi
		return station
	}

	newStation := NewStation(ssid, mac, isAp, channel, rssi)
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
