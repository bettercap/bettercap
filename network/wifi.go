package network

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"

	"github.com/evilsocket/islazy/data"
)

func Dot11Freq2Chan(freq int) int {
	switch {
	case freq <= 2472:
		return ((freq - 2412) / 5) + 1

	case freq == 2484:
		return 14

	case freq >= 5035 && freq <= 5865:
		return ((freq - 5035) / 5) + 7

	case freq >= 5875 && freq <= 5895:
		return 177

	case freq >= 5955 && freq <= 7115: // 6GHz
		return ((freq - 5955) / 5) + 1
	}

	return 0
}

var dot11Channel5GHz = map[int]struct{}{
	36: {}, 40: {}, 44: {}, 48: {},
	52: {}, 56: {}, 60: {}, 64: {},

	68: {}, 72: {}, 76: {}, 80: {},
	100: {}, 104: {}, 108: {}, 112: {},

	116: {}, 120: {}, 124: {}, 128: {},
	132: {}, 136: {}, 140: {}, 144: {},

	149: {}, 153: {}, 157: {}, 161: {},
	165: {}, 169: {}, 173: {}, 177: {},
}

func Dot11Chan2Freq(channel int) int {
	if channel <= 13 {
		return ((channel - 1) * 5) + 2412
	}

	if channel == 14 {
		return 2484
	}

	if _, ok := dot11Channel5GHz[channel]; ok {
		return ((channel - 7) * 5) + 5035
	}

	// 6GHz - Skipped 1-13 to avoid 2Ghz channels conflict
	if channel >= 17 && channel <= 253 {
		return ((channel - 1) * 5) + 5955
	}

	return 0
}

type APNewCallback func(ap *AccessPoint)
type APLostCallback func(ap *AccessPoint)

type WiFi struct {
	mu sync.RWMutex

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
		AccessPoints: w.List(),
	}
	return json.Marshal(doc)
}

func (w *WiFi) EachAccessPoint(cb func(mac string, ap *AccessPoint)) {
	type entry struct {
		mac string
		ap  *AccessPoint
	}
	w.mu.RLock()
	entries := make([]entry, 0, len(w.aps))
	for m, ap := range w.aps {
		entries = append(entries, entry{mac: m, ap: ap})
	}
	w.mu.RUnlock()
	for _, item := range entries {
		cb(item.mac, item.ap)
	}
}

func (w *WiFi) Stations() (list []*Station) {
	w.mu.RLock()
	list = make([]*Station, 0, len(w.aps))
	for _, ap := range w.aps {
		list = append(list, ap.Station())
	}
	w.mu.RUnlock()
	return
}

func (w *WiFi) List() (list []*AccessPoint) {
	w.mu.RLock()
	list = make([]*AccessPoint, 0, len(w.aps))
	for _, ap := range w.aps {
		list = append(list, ap)
	}
	w.mu.RUnlock()
	return
}

func (w *WiFi) Remove(mac string) {
	w.mu.Lock()
	ap, found := w.aps[mac]
	if found {
		delete(w.aps, mac)
	}
	w.mu.Unlock()
	if found && w.lostCb != nil {
		w.lostCb(ap)
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
	mac = NormalizeMac(mac)
	alias := w.aliases.GetOr(mac, "")
	w.mu.RLock()
	ap, found := w.aps[mac]
	w.mu.RUnlock()
	if found {
		ap.Station().updateAccessPoint(ssid, rssi, alias)
		return ap, false
	}

	candidate := NewAccessPoint(ssid, mac, frequency, rssi, w.aliases)
	candidate.SetAlias(alias)

	w.mu.Lock()
	if ap, found = w.aps[mac]; found {
		w.mu.Unlock()
		ap.Station().updateAccessPoint(ssid, rssi, alias)
		return ap, false
	}
	w.aps[mac] = candidate
	w.mu.Unlock()
	if w.newCb != nil {
		w.newCb(candidate)
	}
	return candidate, true
}

func (w *WiFi) Get(mac string) (*AccessPoint, bool) {
	mac = NormalizeMac(mac)
	w.mu.RLock()
	ap, found := w.aps[mac]
	w.mu.RUnlock()
	return ap, found
}

func (w *WiFi) GetClient(mac string) (*Station, bool) {
	mac = NormalizeMac(mac)
	for _, ap := range w.List() {
		if client, found := ap.Get(mac); found {
			return client, true
		}
	}

	return nil, false
}

func (w *WiFi) Clear() {
	w.mu.Lock()
	w.aps = make(map[string]*AccessPoint)
	w.mu.Unlock()
}

func (w *WiFi) NumAPs() int {
	w.mu.RLock()
	n := len(w.aps)
	w.mu.RUnlock()
	return n
}

func (w *WiFi) NumHandshakes() int {
	sum := 0
	for _, ap := range w.List() {
		for _, station := range ap.Clients() {
			if station.Handshake().Complete() {
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

	for _, ap := range w.List() {
		for _, station := range ap.Clients() {
			// if half (which includes also complete) or has pmkid
			handshake := station.Handshake()
			if handshake.Any() {
				err = nil
				handshake.EachUnsavedPacket(func(pkt gopacket.Packet) {
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
