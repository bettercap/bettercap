package network

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	pathNameCleaner = regexp.MustCompile("[^a-zA-Z0-9]+")
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
	return len(s.WPS) > 0
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
