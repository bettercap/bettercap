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

	// Several Station fields (WPS, and separately Encryption/Cipher/
	// Authentication) are written from the packet-processing goroutine
	// (WiFiModule.updateInfo, on every incoming frame) with no
	// synchronization at all, while being read concurrently both by
	// interactive console commands and, critically, by
	// AccessPoint.MarshalJSON() every time the REST API streams an event
	// over its websocket -- Go's json package walks these fields
	// directly via reflection for that, completely bypassing
	// AccessPoint's own RWMutex (which only protects the AccessPoint
	// struct itself, not fields on the Station objects it embeds/
	// references). A write racing with that read corrupts memory and
	// panics -- confirmed via two separate real device crashes:
	//   1) "index out of range" inside encoding/json.mapEncoder.encode,
	//      from the WPS map specifically.
	//   2) a SIGSEGV inside encoding/json.appendString (a corrupted Go
	//      string header read mid-write), from the unguarded
	//      Encryption/Cipher/Authentication writes a few lines below
	//      the WPS write in the same updateInfo() function -- the WPS
	//      fix above didn't cover these, since they're separate fields
	//      hit by a separate (adjacent) unsynchronized write.
	// A single package-level lock covers all of them: deliberately
	// coarse-grained rather than one lock per Station, since these
	// writes are rare relative to packet-processing throughput (so
	// contention is a non-issue), and a shared lock avoids embedding a
	// sync.Mutex inside Station itself -- Station.BSSID() already takes
	// a value receiver, so copying an embedded live mutex by value
	// would be its own bug.
	stationMu sync.RWMutex
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
	stationMu.RLock()
	defer stationMu.RUnlock()
	return len(s.WPS) > 0
}

// SetWPS safely records a single WPS info element, replacing any direct
// write to s.WPS -- see the stationMu comment for why this can't just be
// a plain map assignment.
func (s *Station) SetWPS(name, value string) {
	stationMu.Lock()
	defer stationMu.Unlock()
	s.WPS[name] = value
}

// WPSInfo returns a snapshot copy of the WPS map, safe to range over
// without holding any lock for the duration of the caller's loop.
func (s *Station) WPSInfo() map[string]string {
	stationMu.RLock()
	defer stationMu.RUnlock()
	cp := make(map[string]string, len(s.WPS))
	for k, v := range s.WPS {
		cp[k] = v
	}
	return cp
}

// SetEncryption safely records encryption/cipher/authentication info,
// replacing the direct 3-field assignment in WiFiModule.updateInfo --
// see the stationMu comment: these three plain string fields were being
// written with no synchronization at all, racing against
// MarshalJSON()/IsOpen() reads.
func (s *Station) SetEncryption(enc, cipher, auth string) {
	stationMu.Lock()
	defer stationMu.Unlock()
	s.Encryption = enc
	s.Cipher = cipher
	s.Authentication = auth
}

// MarshalJSON overrides the default reflection-based encoding so every
// field is read under stationMu -- without this, json.Marshal walks the
// struct's fields directly via reflection, which is exactly the access
// pattern that raced with SetWPS()/SetEncryption() and crashed
// bettercap (twice, on two different fields). stationAlias has the
// identical field layout so the output format is unchanged; it exists
// purely to avoid infinitely recursing back into this same MarshalJSON
// method.
func (s *Station) MarshalJSON() ([]byte, error) {
	stationMu.RLock()
	defer stationMu.RUnlock()
	type stationAlias Station
	return json.Marshal((*stationAlias)(s))
}

func (s *Station) IsOpen() bool {
	stationMu.RLock()
	defer stationMu.RUnlock()
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
