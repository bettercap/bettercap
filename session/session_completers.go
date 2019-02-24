package session

import (
	"strings"

	"github.com/bettercap/bettercap/network"
)

func prefixMatches(prefix, what string) bool {
	return prefix == "" || strings.HasPrefix(what, prefix)
}

func addIfMatches(to *[]string, prefix string, what string) {
	if prefixMatches(prefix, what) {
		*to = append(*to, what)
	}
}

func (s *Session) LANCompleter(prefix string) []string {
	macs := []string{""}
	s.Lan.EachHost(func(mac string, e *network.Endpoint) {
		addIfMatches(&macs, prefix, mac)
	})
	return macs
}

func (s *Session) WiFiCompleter(prefix string) []string {
	macs := []string{""}
	s.WiFi.EachAccessPoint(func(mac string, ap *network.AccessPoint) {
		addIfMatches(&macs, prefix, mac)
	})
	return macs
}

func (s *Session) WiFiCompleterFull(prefix string) []string {
	macs := []string{""}
	s.WiFi.EachAccessPoint(func(mac string, ap *network.AccessPoint) {
		addIfMatches(&macs, prefix, mac)
		ap.EachClient(func(mac string, c *network.Station) {
			addIfMatches(&macs, prefix, mac)
		})
	})
	return macs
}

func (s *Session) BLECompleter(prefix string) []string {
	macs := []string{""}
	s.BLE.EachDevice(func(mac string, dev *network.BLEDevice) {
		addIfMatches(&macs, prefix, mac)
	})
	return macs
}

func (s *Session) HIDCompleter(prefix string) []string {
	macs := []string{""}
	s.HID.EachDevice(func(mac string, dev *network.HIDDevice) {
		addIfMatches(&macs, prefix, mac)
	})
	return macs
}

func (s *Session) EventsCompleter(prefix string) []string {
	events := []string{""}
	all := []string{
		"sys.log",
		"session.started",
		"session.closing",
		"update.available",
		"mod.started",
		"mod.stopped",
		"endpoint.new",
		"endpoint.lost",
		"wifi.client.lost",
		"wifi.client.probe",
		"wifi.client.new",
		"wifi.client.handshake",
		"wifi.ap.new",
		"wifi.ap.lost",
		"ble.device.service.discovered",
		"ble.device.characteristic.discovered",
		"ble.device.connected",
		"ble.device.new",
		"ble.device.lost",
		"ble.connection.timeout",
		"hid.device.new",
		"hid.device.lost",
		"http.spoofed-request",
		"http.spoofed-response",
		"https.spoofed-request",
		"https.spoofed-response",
		"syn.scan",
		"net.sniff.mdns",
		"net.sniff.mdns",
		"net.sniff.dot11",
		"net.sniff.tcp",
		"net.sniff.upnp",
		"net.sniff.ntlm",
		"net.sniff.ftp",
		"net.sniff.udp",
		"net.sniff.krb5",
		"net.sniff.dns",
		"net.sniff.teamviewer",
		"net.sniff.http.request",
		"net.sniff.http.response",
		"net.sniff.sni",
	}

	for _, e := range all {
		addIfMatches(&events, prefix, e)
	}

	return events
}
