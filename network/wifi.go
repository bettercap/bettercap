package network

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"

	"github.com/evilsocket/islazy/data"
)

func Dot11Freq2Chan(freq int) int {
	if freq <= 2472 {
		return ((freq - 2412) / 5) + 1
	} else if freq == 2484 {
		return 14
	} else if freq >= 5035 && freq <= 5865 {
		return ((freq - 5035) / 5) + 7
	} else if freq >= 5875 && freq <= 5895 {
		return 177
	} else if freq >= 5955 && freq <= 7115 { // 6GHz
		return ((freq - 5955) / 5) + 1
	}
	return 0
}
func Dot11Chan2Freq(channel int) int {
	if channel <= 13 {
		return ((channel - 1) * 5) + 2412
	} else if channel == 14 {
		return 2484
	} else if channel == 36 || channel == 40 || channel == 44 || channel == 48 ||
		channel == 52 || channel == 56 || channel == 60 || channel == 64 ||
		channel == 68 || channel == 72 || channel == 76 || channel == 80 ||
		channel == 100 || channel == 104 || channel == 108 || channel == 112 ||
		channel == 116 || channel == 120 || channel == 124 || channel == 128 ||
		channel == 132 || channel == 136 || channel == 140 || channel == 144 ||
		channel == 149 || channel == 153 || channel == 157 || channel == 161 ||
		channel == 165 || channel == 169 || channel == 173 || channel == 177 {
		return ((channel - 7) * 5) + 5035
		// 6GHz - Skipped 1-13 to avoid 2Ghz channels conflict
	} else if channel >= 17 && channel <= 253 {
		return ((channel - 1) * 5) + 5955
	}
	return 0
}

type APNewCallback func(ap *AccessPoint)
type APLostCallback func(ap *AccessPoint)

type WiFi struct {
	sync.RWMutex

	aliases *data.UnsortedKV
	aps     map[string]*AccessPoint
	iface   *Endpoint
	newCb   APNewCallback
	lostCb  APLostCallback
}

type wifiJSON struct {
	AccessPoints []*AccessPoint `json:"aps"`
}

func NewWiFi(iface *Endpoint, aliases *data.UnsortedKV, newcb APNewCallback, lostcb APLostCallback) *WiFi {
	return &WiFi{
		aps:     make(map[string]*AccessPoint),
		aliases: aliases,
		iface:   iface,
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (w *WiFi) MarshalJSON() ([]byte, error) {

	doc := wifiJSON{
		// we know the length so preallocate to reduce memory allocations
		AccessPoints: make([]*AccessPoint, 0, len(w.aps)),
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
	w.RLock()
	defer w.RUnlock()

	list = make([]*Station, 0, len(w.aps))

	for _, ap := range w.aps {
		list = append(list, ap.Station)
	}
	return
}

func (w *WiFi) List() (list []*AccessPoint) {
	w.RLock()
	defer w.RUnlock()

	list = make([]*AccessPoint, 0, len(w.aps))

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
		if !strconv.IsPrint(c) {
			return true
		}
	}
	return false
}

func (w *WiFi) AddIfNew(ssid, mac string, frequency int, rssi int8) (*AccessPoint, bool) {
	w.Lock()
	defer w.Unlock()

	mac = NormalizeMac(mac)
	alias := w.aliases.GetOr(mac, "")
	if ap, found := w.aps[mac]; found {
		ap.LastSeen = time.Now()
		if rssi != 0 {
			ap.RSSI = rssi
		}
		// always get the cleanest one
		if !isBogusMacESSID(ssid) {
			ap.Hostname = ssid
		}

		if alias != "" {
			ap.Alias = alias
		}
		return ap, false
	}

	newAp := NewAccessPoint(ssid, mac, frequency, rssi, w.aliases)
	newAp.Alias = alias
	w.aps[mac] = newAp

	if w.newCb != nil {
		w.newCb(newAp)
	}

	return newAp, true
}

func (w *WiFi) Get(mac string) (*AccessPoint, bool) {
	w.RLock()
	defer w.RUnlock()

	mac = NormalizeMac(mac)
	ap, found := w.aps[mac]
	return ap, found
}

func (w *WiFi) GetClient(mac string) (*Station, bool) {
	w.RLock()
	defer w.RUnlock()

	mac = NormalizeMac(mac)
	for _, ap := range w.aps {
		if client, found := ap.Get(mac); found {
			return client, true
		}
	}

	return nil, false
}

func (w *WiFi) Clear() {
	w.Lock()
	defer w.Unlock()
	w.aps = make(map[string]*AccessPoint)
}

func (w *WiFi) NumAPs() int {
	w.RLock()
	defer w.RUnlock()

	return len(w.aps)
}

func (w *WiFi) NumHandshakes() int {
	w.RLock()
	defer w.RUnlock()

	sum := 0
	for _, ap := range w.aps {
		for _, station := range ap.Clients() {
			if station.Handshake.Complete() {
				sum++
			}
		}
	}

	return sum
}

func (w *WiFi) SaveHandshakesTo(fileName string, linkType layers.LinkType) error {
	// check if folder exists first
	dirName := filepath.Dir(fileName)
	if _, err := os.Stat(dirName); err != nil {
		if err = os.MkdirAll(dirName, os.ModePerm); err != nil {
			return err
		}
	}

	fp, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()

	writer, err := pcapgo.NewNgWriter(fp, linkType)
	if err != nil {
		return err
	}

	defer writer.Flush()

	w.RLock()
	defer w.RUnlock()

	for _, ap := range w.aps {
		for _, station := range ap.Clients() {
			// if half (which includes also complete) or has pmkid
			if station.Handshake.Any() {
				err = nil
				station.Handshake.EachUnsavedPacket(func(pkt gopacket.Packet) {
					if err == nil {
						ci := pkt.Metadata().CaptureInfo
						ci.InterfaceIndex = 0
						err = writer.WritePacket(ci, pkt.Data())
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
