package modules

import (
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/session"
)

const StationsDefaultTTL = 30

type WirelessStation struct {
	Endpoint *network.Endpoint
	Essid    string
	IsAP     bool
	Channel  int
}

type WiFi struct {
	sync.Mutex

	Session   *session.Session `json:"-"`
	Interface *network.Endpoint
	Stations  map[string]*WirelessStation
}

func NewWirelessStation(essid, mac string, isAp bool, channel int) *WirelessStation {
	return &WirelessStation{
		Endpoint: network.NewEndpointNoResolve("0.0.0.0", mac, "", 0),
		Essid:    essid,
		IsAP:     isAp,
		Channel:  channel,
	}
}

func NewWiFi(s *session.Session, iface *network.Endpoint) *WiFi {
	return &WiFi{
		Session:   s,
		Interface: iface,
		Stations:  make(map[string]*WirelessStation),
	}
}

func (w *WiFi) List() (list []*WirelessStation) {
	w.Lock()
	defer w.Unlock()

	list = make([]*WirelessStation, 0)
	for _, t := range w.Stations {
		list = append(list, t)
	}
	return
}

func (w *WiFi) Remove(mac string) {
	w.Lock()
	defer w.Unlock()

	if e, found := w.Stations[mac]; found {
		w.Session.Events.Add("wifi.station.lost", e.Endpoint)
		delete(w.Stations, mac)
	}
}

func (w *WiFi) AddIfNew(ssid, mac string, isAp bool, channel int) *WirelessStation {
	w.Lock()
	defer w.Unlock()

	mac = network.NormalizeMac(mac)
	if t, found := w.Stations[mac]; found {
		w.Stations[mac].Endpoint.LastSeen = time.Now()
		return t
	}

	e := NewWirelessStation(ssid, mac, isAp, channel)

	w.Stations[mac] = e

	w.Session.Events.Add("wifi.station.new", e.Endpoint)

	return nil
}

func (w *WiFi) Clear() error {
	w.Stations = make(map[string]*WirelessStation)
	return nil
}
