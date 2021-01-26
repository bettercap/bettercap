package events_stream

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

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
	timeFormat    string
	outputName    string
	output        io.Writer
	rotation      rotation
	triggerList   *TriggerList
	waitFor       string
	waitChan      chan *session.Event
	eventListener <-chan session.Event
	quit          chan bool
	dumpHttpReqs  bool
	dumpHttpResp  bool
	dumpFormatHex bool
}

func NewEventsStream(s *session.Session) *EventsStream {
	mod := &EventsStream{
		SessionModule: session.NewSessionModule("events.stream", s),
		output:        os.Stdout,
		timeFormat:    "15:04:05",
		quit:          make(chan bool),
		waitChan:      make(chan *session.Event),
		waitFor:       "",
		triggerList:   NewTriggerList(),
	}

	mod.State.Store("ignoring", &mod.Session.EventsIgnoreList)

	mod.AddHandler(session.NewModuleHandler("events.stream on", "",
		"Start events stream.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("events.stream off", "",
		"Stop events stream.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("events.show LIMIT?", "events.show(\\s\\d+)?",
		"Show events stream.",
		func(args []string) error {
			limit := -1
			if len(args) == 1 {
				arg := str.Trim(args[0])
				limit, _ = strconv.Atoi(arg)
			}
			return mod.Show(limit)
		}))

	on := session.NewModuleHandler("events.on TAG COMMANDS", `events\.on ([^\s]+) (.+)`,
		"Run COMMANDS when an event with the specified TAG is triggered.",
		func(args []string) error {
			return mod.addTrigger(args[0], args[1])
		})

	on.Complete("events.on", s.EventsCompleter)

	mod.AddHandler(on)

	mod.AddHandler(session.NewModuleHandler("events.triggers", "",
		"Show the list of event triggers created by the events.on command.",
		func(args []string) error {
			return mod.showTriggers()
		}))

	onClear := session.NewModuleHandler("events.trigger.delete TRIGGER_ID", `events\.trigger\.delete ([^\s]+)`,
		"Remove an event trigger given its TRIGGER_ID (use events.triggers to see the list of triggers).",
		func(args []string) error {
			return mod.clearTrigger(args[0])
		})

	onClear.Complete("events.trigger.delete", mod.triggerList.Completer)

	mod.AddHandler(onClear)

	mod.AddHandler(session.NewModuleHandler("events.triggers.clear", "",
		"Remove all event triggers (use events.triggers to see the list of triggers).",
		func(args []string) error {
			return mod.clearTrigger("")
		}))

	mod.AddHandler(session.NewModuleHandler("events.waitfor TAG TIMEOUT?", `events.waitfor ([^\s]+)([\s\d]*)`,
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
			return mod.startWaitingFor(tag, timeout)
		}))

	ignore := session.NewModuleHandler("events.ignore FILTER", "events.ignore ([^\\s]+)",
		"Events with an identifier matching this filter will not be shown (use multiple times to add more filters).",
		func(args []string) error {
			return mod.Session.EventsIgnoreList.Add(args[0])
		})

	ignore.Complete("events.ignore", s.EventsCompleter)

	mod.AddHandler(ignore)

	include := session.NewModuleHandler("events.include FILTER", "events.include ([^\\s]+)",
		"Used to remove filters passed with the events.ignore command.",
		func(args []string) error {
			return mod.Session.EventsIgnoreList.Remove(args[0])
		})

	include.Complete("events.include", s.EventsCompleter)

	mod.AddHandler(include)

	mod.AddHandler(session.NewModuleHandler("events.filters", "",
		"Print the list of filters used to ignore events.",
		func(args []string) error {
			if mod.Session.EventsIgnoreList.Empty() {
				mod.Printf("Ignore filters list is empty.\n")
			} else {
				mod.Session.EventsIgnoreList.RLock()
				defer mod.Session.EventsIgnoreList.RUnlock()

				for _, filter := range mod.Session.EventsIgnoreList.Filters() {
					mod.Printf("  '%s'\n", string(filter))
				}
			}
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("events.filters.clear", "",
		"Clear the list of filters passed with the events.ignore command.",
		func(args []string) error {
			mod.Session.EventsIgnoreList.Clear()
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("events.clear", "",
		"Clear events stream.",
		func(args []string) error {
			mod.Session.Events.Clear()
			return nil
		}))

	mod.AddParam(session.NewStringParameter("events.stream.output",
		"",
		"",
		"If not empty, events will be written to this file instead of the standard output."))

	mod.AddParam(session.NewStringParameter("events.stream.time.format",
		mod.timeFormat,
		"",
		"Date and time format to use for events reporting."))

	mod.AddParam(session.NewBoolParameter("events.stream.output.rotate",
		"true",
		"If true will enable log rotation."))

	mod.AddParam(session.NewBoolParameter("events.stream.output.rotate.compress",
		"true",
		"If true will enable log rotation compression."))

	mod.AddParam(session.NewStringParameter("events.stream.output.rotate.how",
		"size",
		"(size|time)",
		"Rotate by 'size' or 'time'."))

	mod.AddParam(session.NewStringParameter("events.stream.output.rotate.format",
		"2006-01-02 15:04:05",
		"",
		"Datetime format to use for log rotation file names."))

	mod.AddParam(session.NewDecimalParameter("events.stream.output.rotate.when",
		"10",
		"File size (in MB) or time duration (in seconds) for log rotation."))

	mod.AddParam(session.NewBoolParameter("events.stream.http.request.dump",
		"false",
		"If true all HTTP requests will be dumped."))

	mod.AddParam(session.NewBoolParameter("events.stream.http.response.dump",
		"false",
		"If true all HTTP responses will be dumped."))

	mod.AddParam(session.NewBoolParameter("events.stream.http.format.hex",
		"true",
		"If true dumped HTTP bodies will be in hexadecimal format."))

	return mod
}

func (mod *EventsStream) Name() string {
	return "events.stream"
}

func (mod *EventsStream) Description() string {
	return "Print events as a continuous stream."
}

func (mod *EventsStream) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *EventsStream) Configure() (err error) {
	var output string

	if err, output = mod.StringParam("events.stream.output"); err == nil {
		if output == "" {
			mod.output = os.Stdout
		} else if mod.outputName, err = fs.Expand(output); err == nil {
			mod.output, err = os.OpenFile(mod.outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
		}
	}

	if err, mod.rotation.Enabled = mod.BoolParam("events.stream.output.rotate"); err != nil {
		return err
	} else if err, mod.timeFormat = mod.StringParam("events.stream.time.format"); err != nil {
		return err
	} else if err, mod.rotation.Compress = mod.BoolParam("events.stream.output.rotate.compress"); err != nil {
		return err
	} else if err, mod.rotation.Format = mod.StringParam("events.stream.output.rotate.format"); err != nil {
		return err
	} else if err, mod.rotation.How = mod.StringParam("events.stream.output.rotate.how"); err != nil {
		return err
	} else if err, mod.rotation.Period = mod.DecParam("events.stream.output.rotate.when"); err != nil {
		return err
	}

	if err, mod.dumpHttpReqs = mod.BoolParam("events.stream.http.request.dump"); err != nil {
		return err
	} else if err, mod.dumpHttpResp = mod.BoolParam("events.stream.http.response.dump"); err != nil {
		return err
	} else if err, mod.dumpFormatHex = mod.BoolParam("events.stream.http.format.hex"); err != nil {
		return err
	}

	return err
}

func (mod *EventsStream) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.eventListener = mod.Session.Events.Listen()
		defer mod.Session.Events.Unlisten(mod.eventListener)

		for {
			var e session.Event
			select {
			case e = <-mod.eventListener:
				if e.Tag == mod.waitFor {
					mod.waitFor = ""
					mod.waitChan <- &e
				}

				if !mod.Session.EventsIgnoreList.Ignored(e) {
					mod.View(e, true)
				}

				// this could generate sys.log events and lock the whole
				// events.stream, make it async
				go mod.dispatchTriggers(e)

			case <-mod.quit:
				return
			}
		}
	})
}

func (mod *EventsStream) Show(limit int) error {
	events := mod.Session.Events.Sorted()
	num := len(events)

	selected := []session.Event{}
	for i := range events {
		e := events[num-1-i]
		if !mod.Session.EventsIgnoreList.Ignored(e) {
			selected = append(selected, e)
			if len(selected) == limit {
				break
			}
		}
	}

	if numSelected := len(selected); numSelected > 0 {
		mod.Printf("\n")
		for i := range selected {
			mod.View(selected[numSelected-1-i], false)
		}
		mod.Session.Refresh()
	}

	return nil
}

func (mod *EventsStream) startWaitingFor(tag string, timeout int) error {
	if timeout == 0 {
		mod.Info("waiting for event %s ...", tui.Green(tag))
	} else {
		mod.Info("waiting for event %s for %d seconds ...", tui.Green(tag), timeout)
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			mod.waitFor = ""
			mod.waitChan <- nil
		}()
	}

	mod.waitFor = tag
	event := <-mod.waitChan

	if event == nil {
		return fmt.Errorf("'events.waitFor %s %d' timed out.", tag, timeout)
	} else {
		mod.Debug("got event: %v", event)
	}

	return nil
}

func (mod *EventsStream) Stop() error {
	return mod.SetRunning(false, func() {
		mod.quit <- true
		if mod.output != os.Stdout {
			if fp, ok := mod.output.(*os.File); ok {
				fp.Close()
			}
		}
	})
}
