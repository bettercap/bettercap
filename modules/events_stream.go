package modules

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

type EventsStream struct {
	session.SessionModule
	output        *os.File
	ignoreList    *IgnoreList
	waitFor       string
	waitChan      chan *session.Event
	eventListener <-chan session.Event
	quit          chan bool
}

func NewEventsStream(s *session.Session) *EventsStream {
	stream := &EventsStream{
		SessionModule: session.NewSessionModule("events.stream", s),
		output:        os.Stdout,
		quit:          make(chan bool),
		waitChan:      make(chan *session.Event),
		waitFor:       "",
		ignoreList:    NewIgnoreList(),
	}

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

	stream.AddHandler(session.NewModuleHandler("events.waitfor TAG TIMEOUT?", `events.waitfor ([^\s]+)([\s\d]*)`,
		"Wait for an event with the given tag either forever or for a timeout in seconds.",
		func(args []string) error {
			tag := args[0]
			timeout := 0
			if len(args) == 2 {
				t := core.Trim(args[1])
				if t != "" {
					n, err := strconv.Atoi(t)
					if err != nil {
						return err
					}
					timeout = n
				}
			}
			return stream.startWaitingFor(tag, timeout)
		}))

	stream.AddHandler(session.NewModuleHandler("events.ignore FILTER", "events.ignore ([^\\s]+)",
		"Events with an identifier matching this filter will not be shown (use multiple times to add more filters).",
		func(args []string) error {
			return stream.ignoreList.Add(args[0])
		}))

	stream.AddHandler(session.NewModuleHandler("events.include FILTER", "events.include ([^\\s]+)",
		"Used to remove filters passed with the events.ignore command.",
		func(args []string) error {
			return stream.ignoreList.Remove(args[0])
		}))

	stream.AddHandler(session.NewModuleHandler("events.filters", "",
		"Print the list of filters used to ignore events.",
		func(args []string) error {
			if stream.ignoreList.Empty() {
				fmt.Printf("Ignore filters list is empty.\n")
			} else {
				stream.ignoreList.RLock()
				defer stream.ignoreList.RUnlock()

				for _, filter := range stream.ignoreList.Filters() {
					fmt.Printf("  '%s'\n", string(filter))
				}
			}
			return nil
		}))

	stream.AddHandler(session.NewModuleHandler("events.clear", "",
		"Clear events stream.",
		func(args []string) error {
			stream.Session.Events.Clear()
			return nil
		}))

	stream.AddParam(session.NewStringParameter("events.stream.output",
		"",
		"",
		"If not empty, events will be written to this file instead of the standard output."))

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

func (s *EventsStream) Configure() (err error) {
	var output string

	if err, output = s.StringParam("events.stream.output"); err == nil {
		if output == "" {
			s.output = os.Stdout
		} else if output, err = core.ExpandPath(output); err == nil {
			s.output, err = os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		}
	}

	return err
}

func (s *EventsStream) Start() error {
	if err := s.Configure(); err != nil {
		return err
	}

	return s.SetRunning(true, func() {
		s.eventListener = s.Session.Events.Listen()
		for {
			var e session.Event
			select {
			case e = <-s.eventListener:
				if e.Tag == s.waitFor {
					s.waitFor = ""
					s.waitChan <- &e
				}

				if s.ignoreList.Ignored(e) == false {
					s.View(e, true)
				} else {
					log.Debug("Skipping ignored event %v", e)
				}
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

	selected := events[from:num]
	if len(selected) > 0 {
		fmt.Println()

		for _, e := range selected {
			s.View(e, false)
		}

		s.Session.Refresh()
	}

	return nil
}

func (s *EventsStream) startWaitingFor(tag string, timeout int) error {
	if timeout == 0 {
		log.Info("Waiting for event %s ...", core.Green(tag))
	} else {
		log.Info("Waiting for event %s for %d seconds ...", core.Green(tag), timeout)
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			s.waitFor = ""
			s.waitChan <- nil
		}()
	}

	s.waitFor = tag
	event := <-s.waitChan

	if event == nil {
		return fmt.Errorf("'events.waitFor %s %d' timed out.", tag, timeout)
	} else {
		log.Debug("Got event: %v", event)
	}

	return nil
}

func (s *EventsStream) Stop() error {
	return s.SetRunning(false, func() {
		s.quit <- true
		if s.output != os.Stdout {
			s.output.Close()
		}
	})
}
