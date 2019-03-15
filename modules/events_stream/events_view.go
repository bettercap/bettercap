package events_stream

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/bettercap/modules/net_sniff"
	"github.com/bettercap/bettercap/modules/syn_scan"

	"github.com/google/go-github/github"

	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/islazy/zip"
)

func (mod *EventsStream) viewLogEvent(e session.Event) {
	fmt.Fprintf(mod.output, "[%s] [%s] [%s] %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (mod *EventsStream) viewEndpointEvent(e session.Event) {
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
		fmt.Fprintf(mod.output, "[%s] [%s] endpoint %s%s detected as %s%s.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Bold(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else if e.Tag == "endpoint.lost" {
		fmt.Fprintf(mod.output, "[%s] [%s] endpoint %s%s %s%s lost.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Red(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else {
		fmt.Fprintf(mod.output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			t.String())
	}
}

func (mod *EventsStream) viewModuleEvent(e session.Event) {
	if *mod.Session.Options.Debug {
		fmt.Fprintf(mod.output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			e.Data)
	}
}

func (mod *EventsStream) viewSnifferEvent(e session.Event) {
	if strings.HasPrefix(e.Tag, "net.sniff.http.") {
		mod.viewHttpEvent(e)
	} else {
		fmt.Fprintf(mod.output, "[%s] [%s] %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			e.Data.(net_sniff.SnifferEvent).Message)
	}
}

func (mod *EventsStream) viewSynScanEvent(e session.Event) {
	se := e.Data.(syn_scan.SynScanEvent)
	fmt.Fprintf(mod.output, "[%s] [%s] found open port %d for %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		se.Port,
		tui.Bold(se.Address))
}

func (mod *EventsStream) viewUpdateEvent(e session.Event) {
	update := e.Data.(*github.RepositoryRelease)

	fmt.Fprintf(mod.output, "[%s] [%s] an update to version %s is available at %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Bold(tui.Yellow(e.Tag)),
		tui.Bold(*update.TagName),
		*update.HTMLURL)
}

func (mod *EventsStream) doRotation() {
	if mod.output == os.Stdout {
		return
	} else if !mod.rotation.Enabled {
		return
	}

	mod.rotation.Lock()
	defer mod.rotation.Unlock()

	doRotate := false
	if info, err := mod.output.Stat(); err == nil {
		if mod.rotation.How == "size" {
			doRotate = float64(info.Size()) >= float64(mod.rotation.Period*1024*1024)
		} else if mod.rotation.How == "time" {
			doRotate = info.ModTime().Unix()%int64(mod.rotation.Period) == 0
		}
	}

	if doRotate {
		var err error

		name := fmt.Sprintf("%s-%s", mod.outputName, time.Now().Format(mod.rotation.Format))

		if err := mod.output.Close(); err != nil {
			fmt.Printf("could not close log for rotation: %s\n", err)
			return
		}

		if err := os.Rename(mod.outputName, name); err != nil {
			fmt.Printf("could not rename %s to %s: %s\n", mod.outputName, name, err)
		} else if mod.rotation.Compress {
			zipName := fmt.Sprintf("%s.zip", name)
			if err = zip.Files(zipName, []string{name}); err != nil {
				fmt.Printf("error creating %s: %s", zipName, err)
			} else if err = os.Remove(name); err != nil {
				fmt.Printf("error deleting %s: %s", name, err)
			}
		}

		mod.output, err = os.OpenFile(mod.outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("could not open %s: %s", mod.outputName, err)
		}
	}
}

func (mod *EventsStream) View(e session.Event, refresh bool) {
	var err error
	if err, mod.timeFormat = mod.StringParam("events.stream.time.format"); err != nil {
		fmt.Fprintf(mod.output, "%v", err)
		mod.timeFormat = "15:04:05"
	}

	if e.Tag == "sys.log" {
		mod.viewLogEvent(e)
	} else if strings.HasPrefix(e.Tag, "endpoint.") {
		mod.viewEndpointEvent(e)
	} else if strings.HasPrefix(e.Tag, "wifi.") {
		mod.viewWiFiEvent(e)
	} else if strings.HasPrefix(e.Tag, "ble.") {
		mod.viewBLEEvent(e)
	} else if strings.HasPrefix(e.Tag, "hid.") {
		mod.viewHIDEvent(e)
	} else if strings.HasPrefix(e.Tag, "mod.") {
		mod.viewModuleEvent(e)
	} else if strings.HasPrefix(e.Tag, "net.sniff.") {
		mod.viewSnifferEvent(e)
	} else if e.Tag == "syn.scan" {
		mod.viewSynScanEvent(e)
	} else if e.Tag == "update.available" {
		mod.viewUpdateEvent(e)
	} else {
		fmt.Fprintf(mod.output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}

	if refresh && mod.output == os.Stdout {
		mod.Session.Refresh()
	}

	mod.doRotation()
}
