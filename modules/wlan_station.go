package modules

import (
	"github.com/evilsocket/bettercap-ng/network"
)

type WirelessStation struct {
	*network.Endpoint
	IsAP    bool
	Channel int
}

func NewWirelessStation(essid, bssid string, isAp bool, channel int) *WirelessStation {
	return &WirelessStation{
		Endpoint: network.NewEndpointNoResolve(network.MonitorModeAddress, bssid, essid, 0),
		IsAP:     isAp,
		Channel:  channel,
	}
}

func (s WirelessStation) BSSID() string {
	return s.HwAddress
}

func (s *WirelessStation) ESSID() string {
	return s.Hostname
}
