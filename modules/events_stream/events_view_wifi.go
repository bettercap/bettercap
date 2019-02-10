package events_stream

import (
	"fmt"
	"github.com/bettercap/bettercap/modules/wifi"
	"strings"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

func (s *EventsStream) viewWiFiApEvent(e session.Event) {
	ap := e.Data.(*network.AccessPoint)
	vend := ""
	if ap.Vendor != "" {
		vend = fmt.Sprintf(" (%s)", ap.Vendor)
	}
	rssi := ""
	if ap.RSSI != 0 {
		rssi = fmt.Sprintf(" (%d dBm)", ap.RSSI)
	}

	if e.Tag == "wifi.ap.new" {
		fmt.Fprintf(s.output, "[%s] [%s] wifi access point %s%s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			tui.Bold(ap.ESSID()),
			tui.Dim(tui.Yellow(rssi)),
			tui.Green(ap.BSSID()),
			tui.Dim(vend))
	} else if e.Tag == "wifi.ap.lost" {
		fmt.Fprintf(s.output, "[%s] [%s] wifi access point %s (%s) lost.\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			tui.Red(ap.ESSID()),
			ap.BSSID())
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			ap.String())
	}
}

func (s *EventsStream) viewWiFiClientProbeEvent(e session.Event) {
	probe := e.Data.(wifi.WiFiProbeEvent)
	desc := ""
	if probe.FromAlias != "" {
		desc = fmt.Sprintf(" (%s)", probe.FromAlias)
	} else if probe.FromVendor != "" {
		desc = fmt.Sprintf(" (%s)", probe.FromVendor)
	}
	rssi := ""
	if probe.RSSI != 0 {
		rssi = fmt.Sprintf(" (%d dBm)", probe.RSSI)
	}

	fmt.Fprintf(s.output, "[%s] [%s] station %s%s is probing for SSID %s%s\n",
		e.Time.Format(eventTimeFormat),
		tui.Green(e.Tag),
		probe.FromAddr.String(),
		tui.Dim(desc),
		tui.Bold(probe.SSID),
		tui.Yellow(rssi))
}

func (s *EventsStream) viewWiFiHandshakeEvent(e session.Event) {
	hand := e.Data.(wifi.WiFiHandshakeEvent)

	from := hand.Station.String()
	to := hand.AP.String()
	what := "handshake"

	if ap, found := s.Session.WiFi.Get(hand.AP.String()); found {
		to = fmt.Sprintf("%s (%s)", tui.Bold(ap.ESSID()), tui.Dim(ap.BSSID()))
		what = fmt.Sprintf("%s handshake", ap.Encryption)
	}

	if hand.PMKID != nil {
		what = "RSN PMKID"
	}

	fmt.Fprintf(s.output, "[%s] [%s] captured %s -> %s %s to %s\n",
		e.Time.Format(eventTimeFormat),
		tui.Green(e.Tag),
		from,
		to,
		tui.Red(what),
		hand.File)
}

func (s *EventsStream) viewWiFiClientEvent(e session.Event) {
	ce := e.Data.(wifi.WiFiClientEvent)

	ce.Client.Alias = s.Session.Lan.GetAlias(ce.Client.BSSID())

	if e.Tag == "wifi.client.new" {
		fmt.Fprintf(s.output, "[%s] [%s] new station %s detected for %s (%s)\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			ce.Client.String(),
			tui.Bold(ce.AP.ESSID()),
			tui.Dim(ce.AP.BSSID()))
	} else if e.Tag == "wifi.client.lost" {
		fmt.Fprintf(s.output, "[%s] [%s] station %s disconnected from %s (%s)\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			ce.Client.String(),
			tui.Bold(ce.AP.ESSID()),
			tui.Dim(ce.AP.BSSID()))
	}
}

func (s *EventsStream) viewWiFiEvent(e session.Event) {
	if strings.HasPrefix(e.Tag, "wifi.ap.") {
		s.viewWiFiApEvent(e)
	} else if e.Tag == "wifi.client.probe" {
		s.viewWiFiClientProbeEvent(e)
	} else if e.Tag == "wifi.client.handshake" {
		s.viewWiFiHandshakeEvent(e)
	} else if e.Tag == "wifi.client.new" || e.Tag == "wifi.client.lost" {
		s.viewWiFiClientEvent(e)
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), tui.Green(e.Tag), e)
	}
}
