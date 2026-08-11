package network

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/evilsocket/islazy/data"
)

type AccessPoint struct {
	mu sync.RWMutex

	station         *Station
	aliases         *data.UnsortedKV
	clients         map[string]*Station
	withKeyMaterial bool
}

type accessPointJSON struct {
	*StationSnapshot
	Clients   []*Station `json:"clients"`
	Handshake bool       `json:"handshake"`
}

type accessPointFieldsJSON struct {
	stationFieldsJSON
	Clients   []*Station `json:"clients"`
	Handshake bool       `json:"handshake"`
}

func NewAccessPoint(essid, bssid string, frequency int, rssi int8, aliases *data.UnsortedKV) *AccessPoint {
	return &AccessPoint{
		station: NewStation(essid, bssid, frequency, rssi),
		aliases: aliases,
		clients: make(map[string]*Station),
	}
}

func (ap *AccessPoint) Station() *Station {
	if ap == nil {
		return nil
	}
	ap.mu.RLock()
	station := ap.station
	ap.mu.RUnlock()
	return station
}

func (ap *AccessPoint) Snapshot() StationSnapshot { return ap.Station().Snapshot() }
func (ap *AccessPoint) BSSID() string             { return ap.Station().BSSID() }
func (ap *AccessPoint) HardwareAddr() net.HardwareAddr {
	return ap.Station().HardwareAddr()
}
func (ap *AccessPoint) ESSID() string              { return ap.Station().ESSID() }
func (ap *AccessPoint) IsOpen() bool               { return ap.Station().IsOpen() }
func (ap *AccessPoint) String() string             { return ap.Station().String() }
func (ap *AccessPoint) ShortString() string        { return ap.Station().ShortString() }
func (ap *AccessPoint) PathFriendlyName() string   { return ap.Station().PathFriendlyName() }
func (ap *AccessPoint) HasWPS() bool               { return ap.Station().HasWPS() }
func (ap *AccessPoint) WPSInfo() map[string]string { return ap.Station().WPSInfo() }
func (ap *AccessPoint) SetWPS(name, value string)  { ap.Station().SetWPS(name, value) }
func (ap *AccessPoint) SetAlias(alias string)      { ap.Station().SetAlias(alias) }
func (ap *AccessPoint) AddTraffic(sent, received uint64) {
	ap.Station().AddTraffic(sent, received)
}
func (ap *AccessPoint) SetEncryption(enc, cipher, auth string) {
	ap.Station().SetEncryption(enc, cipher, auth)
}
func (ap *AccessPoint) SetEncryptionIfOpen(enc, cipher, auth string) bool {
	return ap.Station().SetEncryptionIfOpen(enc, cipher, auth)
}
func (ap *AccessPoint) Handshake() *Handshake { return ap.Station().Handshake() }

func (ap *AccessPoint) MarshalJSON() ([]byte, error) {
	if ap == nil {
		return []byte("null"), nil
	}

	ap.mu.RLock()
	station := ap.station
	withKeyMaterial := ap.withKeyMaterial
	clientPointers := make([]*Station, 0, len(ap.clients))
	for _, client := range ap.clients {
		clientPointers = append(clientPointers, client)
	}
	ap.mu.RUnlock()

	var stationSnapshot *StationSnapshot
	hasEndpoint := false
	if station != nil {
		snapshot, hasStationEndpoint := station.snapshotState()
		stationSnapshot = &snapshot
		hasEndpoint = hasStationEndpoint
	}
	if stationSnapshot != nil && !hasEndpoint {
		return json.Marshal(accessPointFieldsJSON{
			stationFieldsJSON: stationFieldsFromSnapshot(*stationSnapshot),
			Clients:           clientPointers,
			Handshake:         withKeyMaterial,
		})
	}

	return json.Marshal(accessPointJSON{
		StationSnapshot: stationSnapshot,
		Clients:         clientPointers,
		Handshake:       withKeyMaterial,
	})
}

func (ap *AccessPoint) UnmarshalJSON(raw []byte) error {
	var doc accessPointJSON
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}

	var station *Station
	if doc.StationSnapshot != nil {
		station = stationFromSnapshot(*doc.StationSnapshot)
		station.hasEndpoint = stationJSONHasEndpoint(raw)
	}
	clients := make(map[string]*Station, len(doc.Clients))
	for _, client := range doc.Clients {
		if client != nil {
			clients[client.BSSID()] = client
		}
	}
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		return err
	}

	ap.mu.Lock()
	ap.station = station
	ap.clients = clients
	ap.withKeyMaterial = doc.Handshake
	ap.aliases = aliases
	ap.mu.Unlock()
	return nil
}

func (ap *AccessPoint) Get(bssid string) (*Station, bool) {
	bssid = NormalizeMac(bssid)
	ap.mu.RLock()
	station, found := ap.clients[bssid]
	ap.mu.RUnlock()
	return station, found
}

func (ap *AccessPoint) RemoveClient(mac string) {
	bssid := NormalizeMac(mac)
	ap.mu.Lock()
	delete(ap.clients, bssid)
	ap.mu.Unlock()
}

func (ap *AccessPoint) AddClientIfNew(bssid string, frequency int, rssi int8) (*Station, bool) {
	bssid = NormalizeMac(bssid)
	ap.mu.RLock()
	station, found := ap.clients[bssid]
	aliases := ap.aliases
	ap.mu.RUnlock()

	alias := ""
	if aliases != nil {
		alias = aliases.GetOr(bssid, "")
	}
	if found {
		station.updateClient(frequency, rssi, alias)
		return station, false
	}

	newStation := NewStation("", bssid, frequency, rssi)
	newStation.SetAlias(alias)

	ap.mu.Lock()
	if station, found = ap.clients[bssid]; found {
		ap.mu.Unlock()
		station.updateClient(frequency, rssi, alias)
		return station, false
	}
	ap.clients[bssid] = newStation
	ap.mu.Unlock()
	return newStation, true
}

func (ap *AccessPoint) NumClients() int {
	ap.mu.RLock()
	n := len(ap.clients)
	ap.mu.RUnlock()
	return n
}

func (ap *AccessPoint) Clients() []*Station {
	ap.mu.RLock()
	list := make([]*Station, 0, len(ap.clients))
	for _, client := range ap.clients {
		list = append(list, client)
	}
	ap.mu.RUnlock()
	return list
}

func (ap *AccessPoint) EachClient(cb func(mac string, station *Station)) {
	type entry struct {
		mac     string
		station *Station
	}
	ap.mu.RLock()
	entries := make([]entry, 0, len(ap.clients))
	for mac, station := range ap.clients {
		entries = append(entries, entry{mac: mac, station: station})
	}
	ap.mu.RUnlock()
	for _, item := range entries {
		cb(item.mac, item.station)
	}
}

func (ap *AccessPoint) WithKeyMaterial(state bool) {
	ap.mu.Lock()
	ap.withKeyMaterial = state
	ap.mu.Unlock()
}

func (ap *AccessPoint) HasKeyMaterial() bool {
	ap.mu.RLock()
	state := ap.withKeyMaterial
	ap.mu.RUnlock()
	return state
}

func (ap *AccessPoint) NumHandshakes() int {
	sum := 0
	for _, client := range ap.Clients() {
		if client.Handshake().Complete() {
			sum++
		}
	}
	return sum
}

func (ap *AccessPoint) HasHandshakes() bool { return ap.NumHandshakes() > 0 }

func (ap *AccessPoint) HasPMKID() bool {
	for _, client := range ap.Clients() {
		if client.Handshake().HasPMKID() {
			return true
		}
	}
	return false
}
