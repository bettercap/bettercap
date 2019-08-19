package wifi

import (
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
	File       string `json:"file"`
	NewPackets int    `json:"new_packets"`
	AP         string `json:"ap"`
	Station    string `json:"station"`
	Half       bool   `json:"half"`
	Full       bool   `json:"full"`
	PMKID      []byte `json:"pmkid"`
}
