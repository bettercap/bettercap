package network

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	pathNameCleaner = regexp.MustCompile("[^a-zA-Z0-9]+")

	// Station.WPS is written from the packet-processing goroutine
	// (WiFiModule.updateInfo, on every incoming frame with WPS info
	// elements) with no synchronization at all, while being read
	// concurrently both by the interactive "wifi.show"/"wifi.show wps"
	// commands and, critically, by AccessPoint.MarshalJSON() every time
	// the REST API streams an event over its websocket -- Go's json
	// package walks the WPS map directly via reflection for that,
	// completely bypassing AccessPoint's own RWMutex (which only
	// protects the AccessPoint struct itself, not fields on the Station
	// objects it embeds/references). A write racing with that read
	// panics with a runtime map-corruption error (confirmed via a real
	// device crash: "index out of range" inside
	// encoding/json.mapEncoder.encode, deep inside
	// AccessPoint.MarshalJSON -> json.Marshal(doc) on this exact field).
	// A single package-level lock is deliberately coarse-grained rather
	// than one lock per Station: WPS reads/writes are rare relative to
	// packet-processing throughput, so contention is a non-issue, and a
	// shared lock avoids embedding a sync.Mutex inside Station itself --
	// Station.BSSID() already takes a value receiver, so copying an
	// embedded live mutex by value would be its own bug.
	wpsMu sync.RWMutex
)

type Station struct {
	*Endpoint
	Frequency      int               `json:"frequency"`
	Channel        int               `json:"channel"`
	RSSI           int8              `json:"rssi"`
	Sent           uint64            `json:"sent"`
	Received       uint64            `json:"received"`
	Encryption     string            `json:"encryption"`
	Cipher         string            `json:"cipher"`
	Authentication string            `json:"authentication"`
	WPS            map[string]string `json:"wps"`
	Handshake      *Handshake        `json:"-"`
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
	return &Station{
		Endpoint:  NewEndpointNoResolve(MonitorModeAddress, bssid, cleanESSID(essid), 0),
		Frequency: frequency,
		Channel:   Dot11Freq2Chan(frequency),
		RSSI:      rssi,
		WPS:       make(map[string]string),
		Handshake: NewHandshake(),
	}
}

func (s Station) BSSID() string {
	return s.HwAddress
}

func (s *Station) ESSID() string {
	return s.Hostname
}

func (s *Station) HasWPS() bool {
	wpsMu.RLock()
	defer wpsMu.RUnlock()
	return len(s.WPS) > 0
}

// SetWPS safely records a single WPS info element, replacing any direct
// write to s.WPS -- see the wpsMu comment for why this can't just be a
// plain map assignment.
func (s *Station) SetWPS(name, value string) {
	wpsMu.Lock()
	defer wpsMu.Unlock()
	s.WPS[name] = value
}

// WPSInfo returns a snapshot copy of the WPS map, safe to range over
// without holding any lock for the duration of the caller's loop.
func (s *Station) WPSInfo() map[string]string {
	wpsMu.RLock()
	defer wpsMu.RUnlock()
	cp := make(map[string]string, len(s.WPS))
	for k, v := range s.WPS {
		cp[k] = v
	}
	return cp
}

// MarshalJSON overrides the default reflection-based encoding so WPS is
// read under wpsMu -- without this, json.Marshal walks s.WPS directly,
// which is exactly the access pattern that raced with SetWPS() and
// crashed bettercap. stationAlias has the identical field layout so the
// output format is unchanged; it exists purely to avoid infinitely
// recursing back into this same MarshalJSON method.
func (s *Station) MarshalJSON() ([]byte, error) {
	wpsMu.RLock()
	defer wpsMu.RUnlock()
	type stationAlias Station
	return json.Marshal((*stationAlias)(s))
}

func (s *Station) IsOpen() bool {
	return s.Encryption == "" || s.Encryption == "OPEN"
}

func (s *Station) PathFriendlyName() string {
	name := ""
	bssid := strings.Replace(s.HwAddress, ":", "", -1)
	if essid := pathNameCleaner.ReplaceAllString(s.Hostname, ""); essid != "" {
		name = fmt.Sprintf("%s_%s", essid, bssid)
	} else {
		name = bssid
	}
	return name
}
