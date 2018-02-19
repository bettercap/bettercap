package network

type Station struct {
	*Endpoint
	IsAP       bool
	Channel    int
	RSSI       int8
	Sent       uint64
	Received   uint64
	Encryption string
}

func NewStation(essid, bssid string, isAp bool, channel int, rssi int8) *Station {
	return &Station{
		Endpoint: NewEndpointNoResolve(MonitorModeAddress, bssid, essid, 0),
		IsAP:     isAp,
		Channel:  channel,
		RSSI:     rssi,
	}
}

func (s Station) BSSID() string {
	return s.HwAddress
}

func (s *Station) ESSID() string {
	return s.Hostname
}
