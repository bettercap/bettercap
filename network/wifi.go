package network

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"
)

func Dot11Freq2Chan(freq int) int {
	if freq <= 2472 {
		return ((freq - 2412) / 5) + 1
	} else if freq == 2484 {
		return 14
	} else if freq >= 5035 && freq <= 5865 {
		return ((freq - 5035) / 5) + 7
	}
	return 0
}

func Dot11Chan2Freq(channel int) int {
	if channel <= 13 {
		return ((channel - 1) * 5) + 2412
	} else if channel == 14 {
		return 2484
	} else if channel <= 173 {
		return ((channel - 7) * 5) + 5035
	}

	return 0
}

type APNewCallback func(ap *AccessPoint)
type APLostCallback func(ap *AccessPoint)

type WiFi struct {
	sync.Mutex

	aps    map[string]*AccessPoint
	iface  *Endpoint
	newCb  APNewCallback
	lostCb APLostCallback
}

type wifiJSON struct {
	AccessPoints []*AccessPoint `json:"aps"`
}

func NewWiFi(iface *Endpoint, newcb APNewCallback, lostcb APLostCallback) *WiFi {
	return &WiFi{
		aps:    make(map[string]*AccessPoint),
		iface:  iface,
		newCb:  newcb,
		lostCb: lostcb,
	}
}

func (w *WiFi) MarshalJSON() ([]byte, error) {
	doc := wifiJSON{
		AccessPoints: make([]*AccessPoint, 0),
	}

	for _, ap := range w.aps {
		doc.AccessPoints = append(doc.AccessPoints, ap)
	}

	return json.Marshal(doc)
}

func (w *WiFi) EachAccessPoint(cb func(mac string, ap *AccessPoint)) {
	w.Lock()
	defer w.Unlock()

	for m, ap := range w.aps {
		cb(m, ap)
	}
}

func (w *WiFi) Stations() (list []*Station) {
	w.Lock()
	defer w.Unlock()

	list = make([]*Station, 0)
	for _, ap := range w.aps {
		list = append(list, ap.Station)
	}
	return
}

func (w *WiFi) List() (list []*AccessPoint) {
	w.Lock()
	defer w.Unlock()

	list = make([]*AccessPoint, 0)
	for _, ap := range w.aps {
		list = append(list, ap)
	}
	return
}

func (w *WiFi) Remove(mac string) {
	w.Lock()
	defer w.Unlock()

	if ap, found := w.aps[mac]; found {
		delete(w.aps, mac)
		if w.lostCb != nil {
			w.lostCb(ap)
		}
	}
}

// when iface is in monitor mode, error
// correction on macOS is crap and we
// get non printable characters .... (ref #61)
func isBogusMacESSID(essid string) bool {
	for _, c := range essid {
		if strconv.IsPrint(c) == false {
			return true
		}
	}
	return false
}

func (w *WiFi) AddIfNew(ssid, mac string, frequency int, rssi int8) *AccessPoint {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	if ap, found := w.aps[mac]; found {
		ap.LastSeen = time.Now()
		ap.RSSI = rssi
		// always get the cleanest one
		if isBogusMacESSID(ssid) == false {
			ap.Hostname = ssid
		}
		return ap
	}

	newAp := NewAccessPoint(ssid, mac, frequency, rssi)
	w.aps[mac] = newAp

	if w.newCb != nil {
		w.newCb(newAp)
	}

	return nil
}

func (w *WiFi) Get(mac string) (*AccessPoint, bool) {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	ap, found := w.aps[mac]
	return ap, found
}

func (w *WiFi) GetClient(mac string) (*Station, bool) {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	for _, ap := range w.aps {
		if client, found := ap.Get(mac); found == true {
			return client, true
		}
	}

	return nil, false
}

func (w *WiFi) Clear() error {
	w.aps = make(map[string]*AccessPoint)
	return nil
}
