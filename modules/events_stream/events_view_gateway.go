package events_stream

import (
	"fmt"
	"io"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewGatewayEvent(output io.Writer, e session.Event) {
	change := e.Data.(session.GatewayChange)

	fmt.Fprintf(output, "[%s] [%s] %s gateway changed: '%s' (%s) -> '%s' (%s)\n",
		e.Time.Format(mod.timeFormat),
		tui.Red(e.Tag),
		string(change.Type),
		change.Prev.IP,
		change.Prev.MAC,
		change.New.IP,
		change.New.MAC)
}