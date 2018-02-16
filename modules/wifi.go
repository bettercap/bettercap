package modules

import (
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/session"
)

type WiFi struct {
	sync.Mutex
	Session   *session.Session
	Interface *network.Endpoint
	Stations  map[string]*WiFiStation
}

func NewWiFi(s *session.Session, iface *network.Endpoint) *WiFi {
	return &WiFi{
		Session:   s,
		Interface: iface,
		Stations:  make(map[string]*WiFiStation),
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

	if station, found := w.Stations[mac]; found {
		w.Session.Events.Add("wifi.station.lost", station)
		delete(w.Stations, mac)
	}
}

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, channel int) *WiFiStation {
	w.Lock()
	defer w.Unlock()

	mac = network.NormalizeMac(mac)
	if station, found := w.Stations[mac]; found {
		w.Stations[mac].LastSeen = time.Now()
		return station
	}

	newStation := NewWiFiStation(ssid, mac, isAp, channel)
	w.Stations[mac] = newStation

	w.Session.Events.Add("wifi.station.new", newStation)

	return nil
}

func (w *WiFi) Clear() error {
	w.Stations = make(map[string]*WiFiStation)
	return nil
}
