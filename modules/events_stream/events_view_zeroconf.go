package events_stream

import (
	"fmt"
	"io"
	"strings"

	"github.com/bettercap/bettercap/v2/modules/zerogod"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewZeroConfEvent(output io.Writer, e session.Event) {
	if e.Tag == "zeroconf.service" {
		event := e.Data.(zerogod.ServiceDiscoveryEvent)
		fmt.Fprintf(output, "[%s] [%s] service %s detected for %s (%s):%d with %d records\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Bold(event.Service.ServiceInstanceName()),
			event.Service.AddrIPv4,
			tui.Dim(event.Service.HostName),
			event.Service.Port,
			len(event.Service.Text),
		)
	} else if e.Tag == "zeroconf.browsing" {
		event := e.Data.(zerogod.BrowsingEvent)
		source := event.Source
		if event.Endpoint != nil {
			source = event.Endpoint.ShortString()
		}

		services := make([]string, 0)
		for _, q := range event.Services {
			services = append(services, tui.Yellow(q))
		}

		instPart := ""
		if len(event.Instances) > 0 {
			instances := make([]string, 0)
			for _, q := range event.Instances {
				instances = append(instances, tui.Green(q))
			}
			instPart = fmt.Sprintf(" and instances %s", strings.Join(instances, ", "))
		}

		textPart := ""
		if len(event.Text) > 0 {
			textPart = fmt.Sprintf("\n  text records: %s\n", strings.Join(event.Text, ", "))
		}

		fmt.Fprintf(output, "[%s] [%s] %s is browsing (%s) for services %s%s\n%s",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			source,
			ops.Ternary(event.Query.QR, "RESPONSE", "QUERY"),
			strings.Join(services, ", "),
			instPart,
			textPart,
		)

	} else {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
