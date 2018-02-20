package network

import (
	"sync"
	"time"
)

type StationNewCallback func(s *Station)
type StationLostCallback func(s *Station)

var Channels5Ghz = [...]int{36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124, 126, 128, 132, 134, 136, 138, 140, 142, 144, 149, 151, 153, 155, 157, 159, 161, 165, 169, 173}

type WiFi struct {
	sync.Mutex

	Stations map[string]*Station

	iface  *Endpoint
	newCb  StationNewCallback
	lostCb StationLostCallback
}

func NewWiFi(iface *Endpoint, newcb StationNewCallback, lostcb StationLostCallback) *WiFi {
	return &WiFi{
		Stations: make(map[string]*Station),
		iface:    iface,
		newCb:    newcb,
		lostCb:   lostcb,
	}
}

func (w *WiFi) List() (list []*Station) {
	w.Lock()
	defer w.Unlock()

	list = make([]*Station, 0)
	for _, t := range w.Stations {
		list = append(list, t)
	}
	return
}

func (w *WiFi) Remove(mac string) {
	w.Lock()
	defer w.Unlock()

	if s, found := w.Stations[mac]; found {
		delete(w.Stations, mac)
		if w.lostCb != nil {
			w.lostCb(s)
		}
	}
}

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, frequency int, rssi int8) *Station {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	if station, found := w.Stations[mac]; found {
		station.LastSeen = time.Now()
		station.RSSI = rssi
		return station
	}

	newStation := NewStation(ssid, mac, isAp, frequency, rssi)
	w.Stations[mac] = newStation

	if w.newCb != nil {
		w.newCb(newStation)
	}

	return nil
}

func (w *WiFi) Get(mac string) (*Station, bool) {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	station, found := w.Stations[mac]
	return station, found
}

func (w *WiFi) Clear() error {
	w.Stations = make(map[string]*Station)
	return nil
}
