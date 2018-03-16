package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/google/go-github/github"
)

const eventTimeFormat = "15:04:05"

func (s *EventsStream) viewLogEvent(e session.Event) {
	fmt.Fprintf(s.output, "[%s] [%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (s *EventsStream) viewWiFiEvent(e session.Event) {
	if strings.HasPrefix(e.Tag, "wifi.ap.") {
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
			fmt.Fprintf(s.output, "[%s] [%s] WiFi access point %s%s detected as %s%s.\n",
				e.Time.Format(eventTimeFormat),
				core.Green(e.Tag),
				core.Bold(ap.ESSID()),
				core.Dim(core.Yellow(rssi)),
				core.Green(ap.BSSID()),
				core.Dim(vend))
		} else if e.Tag == "wifi.ap.lost" {
			fmt.Fprintf(s.output, "[%s] [%s] WiFi access point %s (%s) lost.\n",
				e.Time.Format(eventTimeFormat),
				core.Green(e.Tag),
				core.Red(ap.ESSID()),
				ap.BSSID())
		} else {
			fmt.Fprintf(s.output, "[%s] [%s] %s\n",
				e.Time.Format(eventTimeFormat),
				core.Green(e.Tag),
				ap.String())
		}
	} else if e.Tag == "wifi.client.probe" {
		probe := e.Data.(WiFiProbe)
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

		fmt.Fprintf(s.output, "[%s] [%s] Station %s%s is probing for SSID %s%s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			probe.FromAddr.String(),
			core.Dim(desc),
			core.Bold(probe.SSID),
			core.Yellow(rssi))
	}
}

func (s *EventsStream) viewEndpointEvent(e session.Event) {
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
		fmt.Fprintf(s.output, "[%s] [%s] Endpoint %s%s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Bold(t.IpAddress),
			core.Dim(name),
			core.Green(t.HwAddress),
			core.Dim(vend))
	} else if e.Tag == "endpoint.lost" {
		fmt.Fprintf(s.output, "[%s] [%s] Endpoint %s%s lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Red(t.IpAddress),
			core.Dim(vend))
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			t.String())
	}
}

func (s *EventsStream) viewModuleEvent(e session.Event) {
	fmt.Fprintf(s.output, "[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Data)
}

func (s *EventsStream) viewSnifferEvent(e session.Event) {
	se := e.Data.(SnifferEvent)
	misc := ""

	if e.Tag == "net.sniff.leak.http" {
		req := se.Data.(*http.Request)
		if req.Method != "GET" {
			misc += "\n\n"
			misc += fmt.Sprintf("  Method: %s\n", core.Yellow(req.Method))
			misc += fmt.Sprintf("  URL: %s\n", core.Yellow(req.URL.String()))
			misc += fmt.Sprintf("  Headers:\n")
			for name, values := range req.Header {
				misc += fmt.Sprintf("    %s => %s\n", core.Green(name), strings.Join(values, ", "))
			}

			if err := req.ParseForm(); err == nil {
				misc += "  \n  Form:\n\n"
				for key, values := range req.Form {
					misc += fmt.Sprintf("    %s => %s\n", core.Green(key), core.Bold(strings.Join(values, ", ")))
				}
			} else if req.Body != nil {
				b, _ := ioutil.ReadAll(req.Body)
				misc += fmt.Sprintf("  \n  %s:\n\n    %s\n", core.Bold("Body"), string(b))
			}
		}
	} else if se.Data != nil {
		misc = fmt.Sprintf("%s", se.Data)
	}

	fmt.Fprintf(s.output, "[%s] [%s] %s %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Message,
		misc)
}

func (s *EventsStream) viewSynScanEvent(e session.Event) {
	se := e.Data.(SynScanEvent)
	fmt.Fprintf(s.output, "[%s] [%s] Found open port %d for %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Port,
		core.Bold(se.Address))
}

func (s *EventsStream) viewUpdateEvent(e session.Event) {
	update := e.Data.(*github.RepositoryRelease)

	fmt.Fprintf(s.output, "[%s] [%s] An update to version %s is available at %s\n",
		e.Time.Format(eventTimeFormat),
		core.Bold(core.Yellow(e.Tag)),
		core.Bold(*update.TagName),
		*update.HTMLURL)
}

func (s *EventsStream) View(e session.Event, refresh bool) {
	if e.Tag == "sys.log" {
		s.viewLogEvent(e)
	} else if strings.HasPrefix(e.Tag, "endpoint.") {
		s.viewEndpointEvent(e)
	} else if strings.HasPrefix(e.Tag, "wifi.") {
		s.viewWiFiEvent(e)
	} else if strings.HasPrefix(e.Tag, "ble.") {
		s.viewBLEEvent(e)
	} else if strings.HasPrefix(e.Tag, "mod.") {
		s.viewModuleEvent(e)
	} else if strings.HasPrefix(e.Tag, "net.sniff.") {
		s.viewSnifferEvent(e)
	} else if e.Tag == "syn.scan" {
		s.viewSynScanEvent(e)
	} else if e.Tag == "update.available" {
		s.viewUpdateEvent(e)
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), core.Green(e.Tag), e)
	}

	if refresh && s.output == os.Stdout {
		s.Session.Refresh()
	}
}
