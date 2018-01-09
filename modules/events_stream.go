package modules

import (
	"fmt"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

type EventsStream struct {
	session.SessionModule
	quit chan bool
}

func NewEventsStream(s *session.Session) *EventsStream {
	stream := &EventsStream{
		SessionModule: session.NewSessionModule("events.stream", s),
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

	stream.AddHandler(session.NewModuleHandler("events.clear", "",
		"Clear events stream.",
		func(args []string) error {
			stream.Session.Events.Clear()
			return nil
		}))

	return stream
}

func (s EventsStream) Name() string {
	return "Events Stream"
}

func (s EventsStream) Description() string {
	return "Print events as a continuous stream."
}

func (s EventsStream) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *EventsStream) Start() error {
	if s.Running() == false {
		filter := ""

		if err, v := s.Param("events.stream.filter").Get(s.Session); err != nil {
			return err
		} else {
			filter = v.(string)
		}

		s.SetRunning(true)

		go func() {
			for {
				var e session.Event
				select {
				case e = <-s.Session.Events.NewEvents:
					if filter == "" || strings.Contains(e.Tag, filter) {
						tm := e.Time.Format("2006-01-02 15:04:05")

						if e.Tag == "sys.log" {
							fmt.Printf("[%s] [%s] (%s) %s\n", tm, core.Green(e.Tag), e.Label(), e.Data.(session.LogMessage).Message)
						} else {
							fmt.Printf("[%s] [%s] %v\n", tm, core.Green(e.Tag), e.Data)
						}

						s.Session.Input.Refresh()
					}
					break

				case <-s.quit:
					return
				}
			}
		}()

		return nil
	}

	return fmt.Errorf("Events stream already started.")
}

func (s *EventsStream) Stop() error {
	if s.Running() == true {
		s.SetRunning(false)
		s.quit <- true
		return nil
	}
	return fmt.Errorf("Events stream already stopped.")
}

func (s *EventsStream) OnSessionEnded(sess *session.Session) {
	if s.Running() {
		s.Stop()
	}
}
