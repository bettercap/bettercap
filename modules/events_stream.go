package modules

import (
	"strconv"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

type EventsStream struct {
	session.SessionModule
	filter string
	quit   chan bool
}

func NewEventsStream(s *session.Session) *EventsStream {
	stream := &EventsStream{
		SessionModule: session.NewSessionModule("events.stream", s),
		filter:        "",
		quit:          make(chan bool),
	}

	stream.AddParam(session.NewStringParameter("events.stream.filter",
		"",
		"",
		"If filled, filter events by this prefix type."))

	stream.AddHandler(session.NewModuleHandler("events.stream on", "",
		"Start events stream.",
		func(args []string) error {
			return stream.Start()
		}))

	stream.AddHandler(session.NewModuleHandler("events.stream off", "",
		"Stop events stream.",
		func(args []string) error {
			return stream.Stop()
		}))

	stream.AddHandler(session.NewModuleHandler("events.show LIMIT?", "events.show(\\s\\d+)?",
		"Show events stream.",
		func(args []string) error {
			limit := -1
			if len(args) == 1 {
				arg := core.Trim(args[0])
				limit, _ = strconv.Atoi(arg)
			}
			return stream.Show(limit)
		}))

	stream.AddHandler(session.NewModuleHandler("events.clear", "",
		"Clear events stream.",
		func(args []string) error {
			stream.Session.Events.Clear()
			return nil
		}))

	return stream
}

func (s EventsStream) Name() string {
	return "events.stream"
}

func (s EventsStream) Description() string {
	return "Print events as a continuous stream."
}

func (s EventsStream) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *EventsStream) Configure() error {
	var err error
	if err, s.filter = s.StringParam("events.stream.filter"); err != nil {
		return err
	}
	return nil
}

func (s *EventsStream) Start() error {
	if err := s.Configure(); err != nil {
		return err
	}

	return s.SetRunning(true, func() {
		for {
			var e session.Event
			select {
			case e = <-s.Session.Events.NewEvents:
				s.View(e, true)
				break

			case <-s.quit:
				return
			}
		}
	})
}

func (s *EventsStream) Show(limit int) error {
	events := s.Session.Events.Sorted()
	num := len(events)
	from := 0

	if limit > 0 && num > limit {
		from = num - limit
	}

	for _, e := range events[from:num] {
		s.View(e, false)
	}

	s.Session.Refresh()

	return nil
}

func (s *EventsStream) Stop() error {
	return s.SetRunning(false, func() {
		s.quit <- true
	})
}
