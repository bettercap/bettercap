package events_stream

import (
	"fmt"
	"io"
	"strings"

	"github.com/bettercap/bettercap/v2/modules/zerogod"
	"github.com/bettercap/bettercap/v2/session"
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
		for _, q := range event.Query.Questions {
			services = append(services, tui.Yellow(string(q.Name)))
		}
		/*
			instances := make([]string, 0)
			answers := append(event.Query.Answers, event.Query.Additionals...)
			for _, answer := range answers {
				if answer.Class == layers.DNSClassIN && answer.Type == layers.DNSTypePTR {
					instances = append(instances, tui.Green(string(answer.PTR)))
				} else {
					instances = append(instances, tui.Green(answer.String()))
				}
			}

			advPart := ""
			if len(instances) > 0 {
				advPart = fmt.Sprintf(" and advertising %s", strings.Join(instances, ", "))
			}
		*/

		fmt.Fprintf(output, "[%s] [%s] %s is browsing for services %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			source,
			strings.Join(services, ", "),
		)
	} else {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
