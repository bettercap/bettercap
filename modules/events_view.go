package modules

import (
	"fmt"
	// "sort"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/session"
)

const eventTimeFormat = "2006-01-02 15:04:05"

func (s EventsStream) viewLogEvent(e session.Event) {
	fmt.Printf("[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (s EventsStream) viewApEvent(e session.Event) {
	ap := e.Data.(*network.AccessPoint)
	vend := ""
	if ap.Vendor != "" {
		vend = fmt.Sprintf(" (%s)", ap.Vendor)
	}

	if e.Tag == "wifi.ap.new" {
		fmt.Printf("[%s] WiFi access point %s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			core.Bold(ap.ESSID()),
			core.Green(ap.BSSID()),
			vend)
	} else if e.Tag == "wifi.ap.lost" {
		fmt.Printf("[%s] WiFi access point %s (%s) lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Red(ap.ESSID()),
			ap.BSSID())
	} else {
		fmt.Printf("[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			ap.String())
	}
}

func (s EventsStream) viewEndpointEvent(e session.Event) {
	t := e.Data.(*network.Endpoint)
	vend := ""
	name := ""

	if t.Vendor != "" {
		vend = fmt.Sprintf(" (%s)", t.Vendor)
	}

	if t.Alias != "" {
		name = fmt.Sprintf(" (%s)", t.Alias)
	} else if t.Hostname != "" {
		name = fmt.Sprintf(" (%s)", t.Hostname)
	}

	if e.Tag == "endpoint.new" {
		fmt.Printf("[%s] Endpoint %s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			core.Bold(t.IpAddress),
			core.Green(t.HwAddress),
			vend)
	} else if e.Tag == "endpoint.lost" {
		fmt.Printf("[%s] Endpoint %s%s lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Red(t.IpAddress),
			name)
	} else {
		fmt.Printf("[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			t.String())
	}
}

func (s EventsStream) viewModuleEvent(e session.Event) {
	fmt.Printf("[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Data)
}

func (s EventsStream) viewSnifferEvent(e session.Event) {
	se := e.Data.(SnifferEvent)
	fmt.Printf("[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Message)
}

func (s *EventsStream) View(e session.Event, refresh bool) {
	if s.filter == "" || strings.Contains(e.Tag, s.filter) {
		if e.Tag == "sys.log" {
			s.viewLogEvent(e)
		} else if strings.HasPrefix(e.Tag, "endpoint.") {
			s.viewEndpointEvent(e)
		} else if strings.HasPrefix(e.Tag, "wifi.ap.") {
			s.viewApEvent(e)
		} else if strings.HasPrefix(e.Tag, "mod.") {
			s.viewModuleEvent(e)
		} else if strings.HasPrefix(e.Tag, "net.sniff.") {
			s.viewSnifferEvent(e)
		} else {
			fmt.Printf("[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), core.Green(e.Tag), e)
		}

		if refresh {
			s.Session.Refresh()
		}
	}
}
