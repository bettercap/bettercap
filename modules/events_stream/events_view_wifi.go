package events_stream

import (
	"fmt"
	"github.com/bettercap/bettercap/modules/wifi"
	"io"
	"strings"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewWiFiApEvent(output io.Writer, e session.Event) {
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
		fmt.Fprintf(output, "[%s] [%s] wifi access point %s%s detected as %s%s.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Bold(ap.ESSID()),
			tui.Dim(tui.Yellow(rssi)),
			tui.Green(ap.BSSID()),
			tui.Dim(vend))
	} else if e.Tag == "wifi.ap.lost" {
		fmt.Fprintf(output, "[%s] [%s] wifi access point %s (%s) lost.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Red(ap.ESSID()),
			ap.BSSID())
	} else {
		fmt.Fprintf(output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			ap.String())
	}
}

func (mod *EventsStream) viewWiFiClientProbeEvent(output io.Writer, e session.Event) {
	probe := e.Data.(wifi.ProbeEvent)
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

	fmt.Fprintf(output, "[%s] [%s] station %s%s is probing for SSID %s%s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		probe.FromAddr,
		tui.Dim(desc),
		tui.Bold(probe.SSID),
		tui.Yellow(rssi))
}

func (mod *EventsStream) viewWiFiHandshakeEvent(output io.Writer, e session.Event) {
	hand := e.Data.(wifi.HandshakeEvent)

	from := hand.Station
	to := hand.AP
	what := "handshake"

	if ap, found := mod.Session.WiFi.Get(hand.AP); found {
		to = fmt.Sprintf("%s (%s)", tui.Bold(ap.ESSID()), tui.Dim(ap.BSSID()))
		what = fmt.Sprintf("%s handshake", ap.Encryption)
	}

	if hand.PMKID != nil {
		what = "RSN PMKID"
	} else if hand.Full {
		what += " (full)"
	} else if hand.Half {
		what += " (half)"
	}

	fmt.Fprintf(output, "[%s] [%s] captured %s -> %s %s to %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		from,
		to,
		tui.Red(what),
		hand.File)
}

func (mod *EventsStream) viewWiFiClientEvent(output io.Writer, e session.Event) {
	ce := e.Data.(wifi.ClientEvent)

	ce.Client.Alias = mod.Session.Lan.GetAlias(ce.Client.BSSID())

	if e.Tag == "wifi.client.new" {
		fmt.Fprintf(output, "[%s] [%s] new station %s detected for %s (%s)\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			ce.Client.String(),
			tui.Bold(ce.AP.ESSID()),
			tui.Dim(ce.AP.BSSID()))
	} else if e.Tag == "wifi.client.lost" {
		fmt.Fprintf(output, "[%s] [%s] station %s disconnected from %s (%s)\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			ce.Client.String(),
			tui.Bold(ce.AP.ESSID()),
			tui.Dim(ce.AP.BSSID()))
	}
}

func (mod *EventsStream) viewWiFiDeauthEvent(output io.Writer, e session.Event) {
	deauth := e.Data.(wifi.DeauthEvent)

	fmt.Fprintf(output, "[%s] [%s] a1=%s a2=%s a3=%s reason=%s (%d dBm)\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		deauth.Address1,
		deauth.Address2,
		deauth.Address3,
		tui.Bold(deauth.Reason),
		deauth.RSSI)
}

func (mod *EventsStream) viewWiFiEvent(output io.Writer, e session.Event) {
	if strings.HasPrefix(e.Tag, "wifi.ap.") {
		mod.viewWiFiApEvent(output, e)
	} else if e.Tag == "wifi.deauthentication" {
		mod.viewWiFiDeauthEvent(output, e)
	} else if e.Tag == "wifi.client.probe" {
		mod.viewWiFiClientProbeEvent(output, e)
	} else if e.Tag == "wifi.client.handshake" {
		mod.viewWiFiHandshakeEvent(output, e)
	} else if e.Tag == "wifi.client.new" || e.Tag == "wifi.client.lost" {
		mod.viewWiFiClientEvent(output, e)
	} else {
		fmt.Fprintf(output, "[%s] [%s] %#v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
