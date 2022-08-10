package events_stream

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/bettercap/modules/net_sniff"
	"github.com/bettercap/bettercap/modules/syn_scan"

	"github.com/google/go-github/github"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewLogEvent(output io.Writer, e session.Event) {
	fmt.Fprintf(output, "[%s] [%s] [%s] %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (mod *EventsStream) viewEndpointEvent(output io.Writer, e session.Event) {
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
		fmt.Fprintf(output, "[%s] [%s] endpoint %s%s detected as %s%s.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Bold(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else if e.Tag == "endpoint.lost" {
		fmt.Fprintf(output, "[%s] [%s] endpoint %s%s %s%s lost.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Red(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else {
		fmt.Fprintf(output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			t.String())
	}
}

func (mod *EventsStream) viewModuleEvent(output io.Writer, e session.Event) {
	if *mod.Session.Options.Debug {
		fmt.Fprintf(output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			e.Data)
	}
}

func (mod *EventsStream) viewSnifferEvent(output io.Writer, e session.Event) {
	if strings.HasPrefix(e.Tag, "net.sniff.http.") {
		mod.viewHttpEvent(output, e)
	} else {
		fmt.Fprintf(output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			e.Data.(net_sniff.SnifferEvent).Message)
	}
}

func (mod *EventsStream) viewSynScanEvent(output io.Writer, e session.Event) {
	se := e.Data.(syn_scan.SynScanEvent)
	fmt.Fprintf(output, "[%s] [%s] found open port %d for %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		se.Port,
		tui.Bold(se.Address))
}

func (mod *EventsStream) viewUpdateEvent(output io.Writer, e session.Event) {
	update := e.Data.(*github.RepositoryRelease)

	fmt.Fprintf(output, "[%s] [%s] an update to version %s is available at %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Bold(tui.Yellow(e.Tag)),
		tui.Bold(*update.TagName),
		*update.HTMLURL)
}

func (mod *EventsStream) Render(output io.Writer, e session.Event) {
	var err error
	if err, mod.timeFormat = mod.StringParam("events.stream.time.format"); err != nil {
		fmt.Fprintf(output, "%v", err)
		mod.timeFormat = "15:04:05"
	}

	if e.Tag == "sys.log" {
		mod.viewLogEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "endpoint.") {
		mod.viewEndpointEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "wifi.") {
		mod.viewWiFiEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "ble.") {
		mod.viewBLEEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "hid.") {
		mod.viewHIDEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "gps.") {
		mod.viewGPSEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "mod.") {
		mod.viewModuleEvent(output, e)
	} else if strings.HasPrefix(e.Tag, "net.sniff.") {
		mod.viewSnifferEvent(output, e)
	} else if e.Tag == "syn.scan" {
		mod.viewSynScanEvent(output, e)
	} else if e.Tag == "update.available" {
		mod.viewUpdateEvent(output, e)
	} else if e.Tag == "gateway.change" {
		mod.viewGatewayEvent(output, e)
	} else if e.Tag != "tick" && e.Tag != "session.started" && e.Tag != "session.stopped" {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}

func (mod *EventsStream) View(e session.Event, refresh bool) {
	mod.Render(mod.output, e)

	if refresh && mod.output == os.Stdout {
		mod.Session.Refresh()
	}

	mod.doRotation()
}
