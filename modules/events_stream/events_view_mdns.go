package events_stream

import (
	"fmt"
	"io"

	"github.com/bettercap/bettercap/v2/modules/zerogod"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewMDNSEvent(output io.Writer, e session.Event) {
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
}
