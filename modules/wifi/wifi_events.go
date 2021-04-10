package wifi

import (
	"github.com/bettercap/bettercap/network"
)

type ClientEvent struct {
	AP     *network.AccessPoint
	Client *network.Station
}

type DeauthEvent struct {
	RSSI     int8                 `json:"rssi"`
	AP       *network.AccessPoint `json:"ap"`
	Address1 string               `json:"address1"`
	Address2 string               `json:"address2"`
	Address3 string               `json:"address3"`
	Reason   string               `json:"reason"`
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
