package network

type WiFiStation struct {
	*Endpoint
	IsAP       bool
	Channel    int
	RSSI       int8
	Sent       uint64
	Received   uint64
	Encryption string
}

func NewWiFiStation(essid, bssid string, isAp bool, channel int, rssi int8) *WiFiStation {
	return &WiFiStation{
		Endpoint: NewEndpointNoResolve(MonitorModeAddress, bssid, essid, 0),
		IsAP:     isAp,
		Channel:  channel,
		RSSI:     rssi,
	}
}

func (s WiFiStation) BSSID() string {
	return s.HwAddress
}

func (s *WiFiStation) ESSID() string {
	return s.Hostname
}
