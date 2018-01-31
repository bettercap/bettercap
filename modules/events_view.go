package modules

import (
	"fmt"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

const eventTimeFormat = "2006-01-02 15:04:05"

func (s EventsStream) viewLogEvent(e session.Event) {
	fmt.Printf("[%s] [%s] (%s) %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (s EventsStream) viewSnifferEvent(e session.Event) {
	se := e.Data.(SnifferEvent)
	fmt.Printf("[%s] [%s] %s > %s | %v\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Source,
		se.Destination,
		se.Data)
}

func (s *EventsStream) View(e session.Event, refresh bool) {
	if s.filter == "" || strings.Contains(e.Tag, s.filter) {
		if e.Tag == "sys.log" {
			s.viewLogEvent(e)
		} else if strings.HasPrefix(e.Tag, "net.sniff.") {
			s.viewSnifferEvent(e)
		} else {
			fmt.Printf("[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), core.Green(e.Tag), e)
		}

		if refresh {
			s.Session.Refresh()
		}
	}
}
