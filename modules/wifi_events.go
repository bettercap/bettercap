package modules

import (
	"net"

	"github.com/bettercap/bettercap/network"
)

type WiFiClientEvent struct {
	AP     *network.AccessPoint
	Client *network.Station
}

type WiFiProbeEvent struct {
	FromAddr   net.HardwareAddr
	FromVendor string
	FromAlias  string
	SSID       string
	RSSI       int8
}

type WiFiHandshakeEvent struct {
	File       string
	NewPackets int
	AP         net.HardwareAddr
	Station    net.HardwareAddr
	PMKID      []byte
}
