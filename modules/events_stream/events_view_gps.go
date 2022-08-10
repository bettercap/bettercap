package events_stream

import (
	"fmt"
	"io"

	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewGPSEvent(output io.Writer, e session.Event) {
	if e.Tag == "gps.new" {
		gps := e.Data.(session.GPS)

		fmt.Fprintf(output, "[%s] [%s] latitude:%f longitude:%f quality:%s satellites:%d altitude:%f\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			gps.Latitude,
			gps.Longitude,
			gps.FixQuality,
			gps.NumSatellites,
			gps.Altitude)
	}
}
