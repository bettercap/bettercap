package events_stream

import (
	"fmt"
	"io"

	"github.com/bettercap/bettercap/v2/modules/mdns"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewMDNSEvent(output io.Writer, e session.Event) {
	event := e.Data.(mdns.ServiceDiscoveryEvent)
	fmt.Fprintf(output, "[%s] [%s] service %s detected for %s (%s):%d : %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		tui.Bold(event.Service.Name),
		event.Service.AddrV4.String(),
		tui.Dim(event.Service.Host),
		event.Service.Port,
		event.Service.Info,
	)
}
