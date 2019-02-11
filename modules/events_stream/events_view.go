package events_stream

import (
	"fmt"
	"github.com/bettercap/bettercap/modules/net_sniff"
	"os"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/bettercap/modules/syn_scan"

	"github.com/google/go-github/github"

	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/islazy/zip"
)

const eventTimeFormat = "15:04:05"

func (s *EventsStream) viewLogEvent(e session.Event) {
	fmt.Fprintf(s.output, "[%s] [%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		tui.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
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
		fmt.Fprintf(s.output, "[%s] [%s] endpoint %s%s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			tui.Bold(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else if e.Tag == "endpoint.lost" {
		fmt.Fprintf(s.output, "[%s] [%s] endpoint %s%s %s%s lost.\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			tui.Red(t.IpAddress),
			tui.Dim(name),
			tui.Green(t.HwAddress),
			tui.Dim(vend))
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			t.String())
	}
}

func (s *EventsStream) viewModuleEvent(e session.Event) {
	fmt.Fprintf(s.output, "[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		tui.Green(e.Tag),
		e.Data)
}

func (s *EventsStream) viewSnifferEvent(e session.Event) {
	if strings.HasPrefix(e.Tag, "net.sniff.http.") {
		s.viewHttpEvent(e)
	} else {
		fmt.Fprintf(s.output, "[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			tui.Green(e.Tag),
			e.Data.(net_sniff.SnifferEvent).Message)
	}
}

func (s *EventsStream) viewSynScanEvent(e session.Event) {
	se := e.Data.(syn_scan.SynScanEvent)
	fmt.Fprintf(s.output, "[%s] [%s] found open port %d for %s\n",
		e.Time.Format(eventTimeFormat),
		tui.Green(e.Tag),
		se.Port,
		tui.Bold(se.Address))
}

func (s *EventsStream) viewUpdateEvent(e session.Event) {
	update := e.Data.(*github.RepositoryRelease)

	fmt.Fprintf(s.output, "[%s] [%s] an update to version %s is available at %s\n",
		e.Time.Format(eventTimeFormat),
		tui.Bold(tui.Yellow(e.Tag)),
		tui.Bold(*update.TagName),
		*update.HTMLURL)
}

func (s *EventsStream) doRotation() {
	if s.output == os.Stdout {
		return
	} else if !s.rotation.Enabled {
		return
	}

	s.rotation.Lock()
	defer s.rotation.Unlock()

	doRotate := false
	if info, err := s.output.Stat(); err == nil {
		if s.rotation.How == "size" {
			doRotate = float64(info.Size()) >= float64(s.rotation.Period*1024*1024)
		} else if s.rotation.How == "time" {
			doRotate = info.ModTime().Unix()%int64(s.rotation.Period) == 0
		}
	}

	if doRotate {
		var err error

		name := fmt.Sprintf("%s-%s", s.outputName, time.Now().Format(s.rotation.Format))

		if err := s.output.Close(); err != nil {
			fmt.Printf("could not close log for rotation: %s\n", err)
			return
		}

		if err := os.Rename(s.outputName, name); err != nil {
			fmt.Printf("could not rename %s to %s: %s\n", s.outputName, name, err)
		} else if s.rotation.Compress {
			zipName := fmt.Sprintf("%s.zip", name)
			if err = zip.Files(zipName, []string{name}); err != nil {
				fmt.Printf("error creating %s: %s", zipName, err)
			} else if err = os.Remove(name); err != nil {
				fmt.Printf("error deleting %s: %s", name, err)
			}
		}

		s.output, err = os.OpenFile(s.outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("could not open %s: %s", s.outputName, err)
		}
	}
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
		fmt.Fprintf(s.output, "[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), tui.Green(e.Tag), e)
	}

	if refresh && s.output == os.Stdout {
		s.Session.Refresh()
	}

	s.doRotation()
}
