package network

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"

	"github.com/evilsocket/islazy/data"
	"github.com/evilsocket/islazy/fs"
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
	} else if channel == 177 {
		return 5885
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

	doHead := !fs.Exists(fileName)
	fp, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()

	writer := pcapgo.NewWriter(fp)

	if doHead {
		if err = writer.WriteFileHeader(65536, linkType); err != nil {
			return err
		}
	}

	w.RLock()
	defer w.RUnlock()

	for _, ap := range w.aps {
		for _, station := range ap.Clients() {
			// if half (which includes also complete) or has pmkid
			if station.Handshake.Any() {
				err = nil
				station.Handshake.EachUnsavedPacket(func(pkt gopacket.Packet) {
					if err == nil {
						err = writer.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data())
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
