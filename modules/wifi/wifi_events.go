package wifi

import (
	"net"

	"github.com/bettercap/bettercap/network"
)

type ClientEvent struct {
	AP     *network.AccessPoint
	Client *network.Station
}

type ProbeEvent struct {
	FromAddr   string `json:"mac"`
	FromVendor string `json:"vendor"`
	FromAlias  string `json:"alias"`
	SSID       string `json:"essid"`
	RSSI       int8   `json:"rssi"`
}

type HandshakeEvent struct {
	File       string
	NewPackets int
	AP         net.HardwareAddr
	Station    net.HardwareAddr
	PMKID      []byte
}
