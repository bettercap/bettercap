package session

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	DEBUG = iota
	INFO
	IMPORTANT
	WARNING
	ERROR
	FATAL
)

type Event struct {
	Tag  string      `json:"tag"`
	Time time.Time   `json:"time"`
	Data interface{} `json:"data"`
}

func NewEvent(tag string, data interface{}) Event {
	return Event{
		Tag:  tag,
		Time: time.Now(),
		Data: data,
	}
}

type EventPool struct {
	NewEvents chan Event
	debug     bool
	silent    bool
	events    []Event
	lock      *sync.Mutex
}

func NewEventPool(debug bool, silent bool) *EventPool {
	return &EventPool{
		NewEvents: make(chan Event),
		debug:     debug,
		silent:    silent,
		events:    make([]Event, 0),
		lock:      &sync.Mutex{},
	}
}

func (p *EventPool) Add(tag string, data interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	e := NewEvent(tag, data)
	p.events = append([]Event{e}, p.events...)

	select {
	case p.NewEvents <- e:
		break
	default:
	}
}

func (p *EventPool) Log(level int, format string, args ...interface{}) {
	if level == DEBUG && p.debug == false {
		return
	} else if level < ERROR && p.silent == true {
		return
	}

	p.Add("sys.log", struct {
		Level   int
		Message string
	}{
		level,
		fmt.Sprintf(format, args...),
	})

	if level == FATAL {
		os.Exit(1)
	}
}

func (p *EventPool) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.events = make([]Event, 0)
}

func (p *EventPool) Events() []Event {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.events
}
