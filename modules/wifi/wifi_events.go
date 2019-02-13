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
	FromAddr   net.HardwareAddr
	FromVendor string
	FromAlias  string
	SSID       string
	RSSI       int8
}

type HandshakeEvent struct {
	File       string
	NewPackets int
	AP         net.HardwareAddr
	Station    net.HardwareAddr
	PMKID      []byte
}
