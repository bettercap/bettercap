package network

type Station struct {
	*Endpoint
	IsAP       bool
	Frequency  int
	RSSI       int8
	Sent       uint64
	Received   uint64
	Encryption string
}

func NewStation(essid, bssid string, isAp bool, frequency int, rssi int8) *Station {
	return &Station{
		Endpoint:  NewEndpointNoResolve(MonitorModeAddress, bssid, essid, 0),
		IsAP:      isAp,
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
