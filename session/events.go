package session

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
)

type Event struct {
	Tag  string      `json:"tag"`
	Time time.Time   `json:"time"`
	Data interface{} `json:"data"`
}

type LogMessage struct {
	Level   int
	Message string
}

func NewEvent(tag string, data interface{}) Event {
	return Event{
		Tag:  tag,
		Time: time.Now(),
		Data: data,
	}
}

func (e Event) Label() string {
	log := e.Data.(LogMessage)
	label := core.LogLabels[log.Level]
	color := core.LogColors[log.Level]
	return color + label + core.RESET
}

type EventPool struct {
	sync.Mutex

	NewEvents chan Event
	debug     bool
	silent    bool
	events    []Event
}

func NewEventPool(debug bool, silent bool) *EventPool {
	return &EventPool{
		NewEvents: make(chan Event, 0xff),
		debug:     debug,
		silent:    silent,
		events:    make([]Event, 0),
	}
}

func (p *EventPool) Add(tag string, data interface{}) {
	p.Lock()
	defer p.Unlock()
	e := NewEvent(tag, data)
	p.events = append([]Event{e}, p.events...)
	p.NewEvents <- e
}

func (p *EventPool) Log(level int, format string, args ...interface{}) {
	if level == core.DEBUG && p.debug == false {
		return
	} else if level < core.ERROR && p.silent == true {
		return
	}

	message := fmt.Sprintf(format, args...)

	p.Add("sys.log", LogMessage{
		level,
		message,
	})

	if level == core.FATAL {
		fmt.Fprintf(os.Stderr, "%s\n", message)
		os.Exit(1)
	}
}

func (p *EventPool) Clear() {
	p.Lock()
	defer p.Unlock()
	p.events = make([]Event, 0)
}

func (p *EventPool) Events() []Event {
	p.Lock()
	defer p.Unlock()
	return p.events
}

func (p *EventPool) Sorted() []Event {
	p.Lock()
	defer p.Unlock()

	sort.Slice(p.events, func(i, j int) bool {
		return p.events[i].Time.Before(p.events[j].Time)
	})

	return p.events
}
