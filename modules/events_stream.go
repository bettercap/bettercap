package modules

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

type rotation struct {
	sync.Mutex
	Enabled  bool
	Compress bool
	Format   string
	How      string
	Period   float64
}

type EventsStream struct {
	session.SessionModule
	outputName    string
	output        *os.File
	rotation      rotation
	ignoreList    *IgnoreList
	waitFor       string
	waitChan      chan *session.Event
	eventListener <-chan session.Event
	quit          chan bool
	dumpHttpReqs  bool
	dumpHttpResp  bool
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
				arg := str.Trim(args[0])
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
				t := str.Trim(args[1])
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

	stream.AddParam(session.NewBoolParameter("events.stream.output.rotate",
		"true",
		"If true will enable log rotation."))

	stream.AddParam(session.NewBoolParameter("events.stream.output.rotate.compress",
		"true",
		"If true will enable log rotation compression."))

	stream.AddParam(session.NewStringParameter("events.stream.output.rotate.how",
		"size",
		"(size|time)",
		"Rotate by 'size' or 'time'."))

	stream.AddParam(session.NewStringParameter("events.stream.output.rotate.format",
		"2006-01-02 15:04:05",
		"",
		"Datetime format to use for log rotation file names."))

	stream.AddParam(session.NewDecimalParameter("events.stream.output.rotate.when",
		"10",
		"File size (in MB) or time duration (in seconds) for log rotation."))

	stream.AddParam(session.NewBoolParameter("events.stream.http.request.dump",
		"false",
		"If true all HTTP requests will be dumped."))

	stream.AddParam(session.NewBoolParameter("events.stream.http.response.dump",
		"false",
		"If true all HTTP responses will be dumped."))

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
		} else if s.outputName, err = fs.Expand(output); err == nil {
			s.output, err = os.OpenFile(s.outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		}
	}

	if err, s.rotation.Enabled = s.BoolParam("events.stream.output.rotate"); err != nil {
		return err
	} else if err, s.rotation.Compress = s.BoolParam("events.stream.output.rotate.compress"); err != nil {
		return err
	} else if err, s.rotation.Format = s.StringParam("events.stream.output.rotate.format"); err != nil {
		return err
	} else if err, s.rotation.How = s.StringParam("events.stream.output.rotate.how"); err != nil {
		return err
	} else if err, s.rotation.Period = s.DecParam("events.stream.output.rotate.when"); err != nil {
		return err
	}

	if err, s.dumpHttpReqs = s.BoolParam("events.stream.http.request.dump"); err != nil {
		return err
	} else if err, s.dumpHttpResp = s.BoolParam("events.stream.http.response.dump"); err != nil {
		return err
	}

	return err
}

func (s *EventsStream) Start() error {
	if err := s.Configure(); err != nil {
		return err
	}

	return s.SetRunning(true, func() {
		s.eventListener = s.Session.Events.Listen()
		defer s.Session.Events.Unlisten(s.eventListener)

		for {
			var e session.Event
			select {
			case e = <-s.eventListener:
				if e.Tag == s.waitFor {
					s.waitFor = ""
					s.waitChan <- &e
				}

				if !s.ignoreList.Ignored(e) {
					s.View(e, true)
				} else {
					log.Debug("skipping ignored event %v", e)
				}

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

	selected := []session.Event{}
	for _, e := range events[from:] {
		if !s.ignoreList.Ignored(e) {
			selected = append(selected, e)
			if len(selected) == limit {
				break
			}
		}
	}

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
		log.Info("waiting for event %s ...", tui.Green(tag))
	} else {
		log.Info("waiting for event %s for %d seconds ...", tui.Green(tag), timeout)
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
		log.Debug("got event: %v", event)
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
