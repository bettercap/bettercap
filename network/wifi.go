package network

import (
	"sync"
	"time"
)

type StationNewCallback func(s *WiFiStation)
type StationLostCallback func(s *WiFiStation)

type WiFi struct {
	sync.Mutex
	Interface *Endpoint
	Stations  map[string]*WiFiStation

	newCb  StationNewCallback
	lostCb StationLostCallback
}

func NewWiFi(iface *Endpoint, newcb StationNewCallback, lostcb StationLostCallback) *WiFi {
	return &WiFi{
		Interface: iface,
		Stations:  make(map[string]*WiFiStation),
		newCb:     newcb,
		lostCb:    lostcb,
	}
}

func (w *WiFi) List() (list []*WiFiStation) {
	w.Lock()
	defer w.Unlock()

	list = make([]*WiFiStation, 0)
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

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, channel int, rssi int8) *WiFiStation {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	if station, found := w.Stations[mac]; found {
		station.LastSeen = time.Now()
		station.RSSI = rssi
		return station
	}

	newStation := NewWiFiStation(ssid, mac, isAp, channel, rssi)
	w.Stations[mac] = newStation

	if w.newCb != nil {
		w.newCb(newStation)
	}

	return nil
}

func (w *WiFi) Clear() error {
	w.Stations = make(map[string]*WiFiStation)
	return nil
}
