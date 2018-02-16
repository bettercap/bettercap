package modules

import (
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/session"
)

type WLan struct {
	sync.Mutex
	Session   *session.Session
	Interface *network.Endpoint
	Stations  map[string]*WirelessStation
}

func NewWLan(s *session.Session, iface *network.Endpoint) *WLan {
	return &WLan{
		Session:   s,
		Interface: iface,
		Stations:  make(map[string]*WirelessStation),
	}
}

func (w *WLan) List() (list []*WirelessStation) {
	w.Lock()
	defer w.Unlock()

	list = make([]*WirelessStation, 0)
	for _, t := range w.Stations {
		list = append(list, t)
	}
	return
}

func (w *WLan) Remove(mac string) {
	w.Lock()
	defer w.Unlock()

	if station, found := w.Stations[mac]; found {
		w.Session.Events.Add("wifi.station.lost", station)
		delete(w.Stations, mac)
	}
}

func (w *WLan) AddIfNew(ssid, mac string, isAp bool, channel int) *WirelessStation {
	w.Lock()
	defer w.Unlock()

	mac = network.NormalizeMac(mac)
	if station, found := w.Stations[mac]; found {
		w.Stations[mac].LastSeen = time.Now()
		return station
	}

	newStation := NewWirelessStation(ssid, mac, isAp, channel)
	w.Stations[mac] = newStation

	w.Session.Events.Add("wifi.station.new", newStation)

	return nil
}

func (w *WLan) Clear() error {
	w.Stations = make(map[string]*WirelessStation)
	return nil
}
