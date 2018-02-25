package network

import (
	"strconv"
)

type Station struct {
	*Endpoint
	Frequency      int    `json:"frequency"`
	RSSI           int8   `json:"rssi"`
	Sent           uint64 `json:"sent"`
	Received       uint64 `json:"received"`
	Encryption     string `json:"encryption"`
	Cipher         string `json:"cipher"`
	Authentication string `json:"authentication"`
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
		RSSI:      rssi,
	}
}

func (s Station) BSSID() string {
	return s.HwAddress
}

func (s *Station) ESSID() string {
	return s.Hostname
}
