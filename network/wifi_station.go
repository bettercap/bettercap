package network

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/evilsocket/islazy/tui"
)

var pathNameCleaner = regexp.MustCompile("[^a-zA-Z0-9]+")

// StationSnapshot is a point-in-time view of a Station. Directly mutable
// reference values such as HW and WPS are copied; Meta remains shared because
// it owns its synchronization and is also shared by Endpoint snapshots.
type StationSnapshot struct {
	HW             net.HardwareAddr  `json:"-"`
	IpAddress      string            `json:"ipv4"`
	Ip6Address     string            `json:"ipv6"`
	HwAddress      string            `json:"mac"`
	Hostname       string            `json:"hostname"`
	Alias          string            `json:"alias"`
	Vendor         string            `json:"vendor"`
	FirstSeen      time.Time         `json:"first_seen"`
	LastSeen       time.Time         `json:"last_seen"`
	Meta           *Meta             `json:"meta"`
	Frequency      int               `json:"frequency"`
	Channel        int               `json:"channel"`
	RSSI           int8              `json:"rssi"`
	Sent           uint64            `json:"sent"`
	Received       uint64            `json:"received"`
	Encryption     string            `json:"encryption"`
	Cipher         string            `json:"cipher"`
	Authentication string            `json:"authentication"`
	WPS            map[string]string `json:"wps"`
}

type Station struct {
	mu sync.RWMutex

	hasEndpoint    bool
	endpoint       Endpoint
	frequency      int
	channel        int
	rssi           int8
	sent           uint64
	received       uint64
	encryption     string
	cipher         string
	authentication string
	wps            map[string]string
	handshake      *Handshake
}

// stationFieldsJSON is the legacy encoding of a Station whose embedded
// Endpoint was nil. Real stations always have an endpoint, but preserving the
// zero-value shape keeps the wire format compatible with older releases.
type stationFieldsJSON struct {
	Frequency      int               `json:"frequency"`
	Channel        int               `json:"channel"`
	RSSI           int8              `json:"rssi"`
	Sent           uint64            `json:"sent"`
	Received       uint64            `json:"received"`
	Encryption     string            `json:"encryption"`
	Cipher         string            `json:"cipher"`
	Authentication string            `json:"authentication"`
	WPS            map[string]string `json:"wps"`
}

func stationFieldsFromSnapshot(snapshot StationSnapshot) stationFieldsJSON {
	return stationFieldsJSON{
		Frequency:      snapshot.Frequency,
		Channel:        snapshot.Channel,
		RSSI:           snapshot.RSSI,
		Sent:           snapshot.Sent,
		Received:       snapshot.Received,
		Encryption:     snapshot.Encryption,
		Cipher:         snapshot.Cipher,
		Authentication: snapshot.Authentication,
		WPS:            snapshot.WPS,
	}
}

func cleanESSID(essid string) string {
	res := ""
	for _, c := range essid {
		if strconv.IsPrint(c) {
			res += string(c)
		} else {
			break
		}
	}
	return res
}

func NewStation(essid, bssid string, frequency int, rssi int8) *Station {
	endpoint := NewEndpointNoResolve(MonitorModeAddress, bssid, cleanESSID(essid), 0)
	return &Station{
		hasEndpoint: true,
		endpoint:    *endpoint,
		frequency:   frequency,
		channel:     Dot11Freq2Chan(frequency),
		rssi:        rssi,
		wps:         make(map[string]string),
		handshake:   NewHandshake(),
	}
}

func stationFromSnapshot(snapshot StationSnapshot) *Station {
	hw := append(net.HardwareAddr(nil), snapshot.HW...)
	if len(hw) == 0 && snapshot.HwAddress != "" {
		hw, _ = net.ParseMAC(snapshot.HwAddress)
	}

	endpoint := Endpoint{
		HW:         hw,
		HwAddress:  snapshot.HwAddress,
		IpAddress:  snapshot.IpAddress,
		Ip6Address: snapshot.Ip6Address,
		Hostname:   snapshot.Hostname,
		Alias:      snapshot.Alias,
		Vendor:     snapshot.Vendor,
		FirstSeen:  snapshot.FirstSeen,
		LastSeen:   snapshot.LastSeen,
		Meta:       snapshot.Meta,
	}
	if snapshot.IpAddress != "" {
		endpoint.IP = net.ParseIP(snapshot.IpAddress)
	}
	if snapshot.Ip6Address != "" {
		endpoint.IPv6 = net.ParseIP(snapshot.Ip6Address)
	}

	return &Station{
		hasEndpoint:    true,
		endpoint:       endpoint,
		frequency:      snapshot.Frequency,
		channel:        snapshot.Channel,
		rssi:           snapshot.RSSI,
		sent:           snapshot.Sent,
		received:       snapshot.Received,
		encryption:     snapshot.Encryption,
		cipher:         snapshot.Cipher,
		authentication: snapshot.Authentication,
		wps:            cloneWPS(snapshot.WPS),
		handshake:      NewHandshake(),
	}
}

func cloneWPS(wps map[string]string) map[string]string {
	if wps == nil {
		return nil
	}
	cp := make(map[string]string, len(wps))
	for k, v := range wps {
		cp[k] = v
	}
	return cp
}

func (s *Station) Snapshot() StationSnapshot {
	snapshot, _ := s.snapshotState()
	return snapshot
}

func (s *Station) snapshotState() (StationSnapshot, bool) {
	if s == nil {
		return StationSnapshot{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return StationSnapshot{
		HW:             append(net.HardwareAddr(nil), s.endpoint.HW...),
		IpAddress:      s.endpoint.IpAddress,
		Ip6Address:     s.endpoint.Ip6Address,
		HwAddress:      s.endpoint.HwAddress,
		Hostname:       s.endpoint.Hostname,
		Alias:          s.endpoint.Alias,
		Vendor:         s.endpoint.Vendor,
		FirstSeen:      s.endpoint.FirstSeen,
		LastSeen:       s.endpoint.LastSeen,
		Meta:           s.endpoint.Meta,
		Frequency:      s.frequency,
		Channel:        s.channel,
		RSSI:           s.rssi,
		Sent:           s.sent,
		Received:       s.received,
		Encryption:     s.encryption,
		Cipher:         s.cipher,
		Authentication: s.authentication,
		WPS:            cloneWPS(s.wps),
	}, s.hasEndpoint
}

func (s *Station) BSSID() string {
	return s.Snapshot().HwAddress
}

func (s *Station) HardwareAddr() net.HardwareAddr {
	return s.Snapshot().HW
}

func (s *Station) ESSID() string {
	return s.Snapshot().Hostname
}

func (s *Station) SetAlias(alias string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.endpoint.Alias = alias
	s.mu.Unlock()
}

func (s *Station) SetVendor(vendor string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.endpoint.Vendor = vendor
	s.mu.Unlock()
}

func (s *Station) AddTraffic(sent, received uint64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.sent += sent
	s.received += received
	s.mu.Unlock()
}

func (s *Station) updateAccessPoint(essid string, rssi int8, alias string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.endpoint.LastSeen = time.Now()
	if rssi != 0 {
		s.rssi = rssi
	}
	if !isBogusMacESSID(essid) {
		s.endpoint.Hostname = essid
	}
	if alias != "" {
		s.endpoint.Alias = alias
	}
}

func (s *Station) updateClient(frequency int, rssi int8, alias string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.frequency = frequency
	s.rssi = rssi
	s.endpoint.LastSeen = time.Now()
	if alias != "" {
		s.endpoint.Alias = alias
	}
}

func (s *Station) HasWPS() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.wps) > 0
}

func (s *Station) SetWPS(name, value string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.wps == nil {
		s.wps = make(map[string]string)
	}
	s.wps[name] = value
	s.mu.Unlock()
}

func (s *Station) WPSInfo() map[string]string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneWPS(s.wps)
}

func (s *Station) SetEncryption(enc, cipher, auth string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.encryption = enc
	s.cipher = cipher
	s.authentication = auth
	s.mu.Unlock()
}

func (s *Station) SetEncryptionIfOpen(enc, cipher, auth string) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.encryption != "" && s.encryption != "OPEN" {
		return false
	}
	s.encryption = enc
	s.cipher = cipher
	s.authentication = auth
	return true
}

func (s *Station) Handshake() *Handshake {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	handshake := s.handshake
	s.mu.RUnlock()
	if handshake != nil {
		return handshake
	}

	s.mu.Lock()
	if s.handshake == nil {
		s.handshake = NewHandshake()
	}
	handshake = s.handshake
	s.mu.Unlock()
	return handshake
}

func (s *Station) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	snapshot, hasEndpoint := s.snapshotState()
	if !hasEndpoint {
		return json.Marshal(stationFieldsFromSnapshot(snapshot))
	}
	return json.Marshal(snapshot)
}

func (s *Station) UnmarshalJSON(raw []byte) error {
	var snapshot StationSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return err
	}
	replacement := stationFromSnapshot(snapshot)
	replacement.hasEndpoint = stationJSONHasEndpoint(raw)

	s.mu.Lock()
	s.hasEndpoint = replacement.hasEndpoint
	s.endpoint = replacement.endpoint
	s.frequency = replacement.frequency
	s.channel = replacement.channel
	s.rssi = replacement.rssi
	s.sent = replacement.sent
	s.received = replacement.received
	s.encryption = replacement.encryption
	s.cipher = replacement.cipher
	s.authentication = replacement.authentication
	s.wps = replacement.wps
	s.handshake = replacement.handshake
	s.mu.Unlock()
	return nil
}

func stationJSONHasEndpoint(raw []byte) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return false
	}
	for _, name := range []string{
		"ipv4", "ipv6", "mac", "hostname", "alias", "vendor", "first_seen", "last_seen", "meta",
	} {
		if _, found := fields[name]; found {
			return true
		}
	}
	return false
}

func (s *Station) IsOpen() bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.encryption == "" || s.encryption == "OPEN"
}

func (s *Station) PathFriendlyName() string {
	snapshot := s.Snapshot()
	bssid := strings.Replace(snapshot.HwAddress, ":", "", -1)
	if essid := pathNameCleaner.ReplaceAllString(snapshot.Hostname, ""); essid != "" {
		return fmt.Sprintf("%s_%s", essid, bssid)
	}
	return bssid
}

func (s *Station) String() string {
	snapshot := s.Snapshot()
	ipPart := fmt.Sprintf("%s : ", snapshot.IpAddress)
	if snapshot.IpAddress == MonitorModeAddress {
		ipPart = ""
	}

	if snapshot.HwAddress == "" {
		return snapshot.IpAddress
	} else if snapshot.Vendor == "" {
		return fmt.Sprintf("%s%s", ipPart, snapshot.HwAddress)
	} else if snapshot.Hostname == "" && snapshot.Alias == "" {
		return fmt.Sprintf("%s%s (%s)", ipPart, snapshot.HwAddress, snapshot.Vendor)
	}

	name := snapshot.Hostname
	if snapshot.Alias != "" {
		name = snapshot.Alias
	}
	return fmt.Sprintf("%s%s (%s) - %s", ipPart, snapshot.HwAddress, snapshot.Vendor, tui.Bold(name))
}

func (s *Station) ShortString() string {
	snapshot := s.Snapshot()
	parts := []string{snapshot.IpAddress}
	if snapshot.Vendor != "" {
		parts = append(parts, tui.Dim(fmt.Sprintf("(%s)", snapshot.Vendor)))
	}
	name := snapshot.Hostname
	if snapshot.Alias != "" {
		name = snapshot.Alias
	}
	if name != "" {
		parts = append(parts, tui.Bold(name))
	}
	return strings.Join(parts, " ")
}
