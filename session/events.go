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

const (
	BOLD = "\033[1m"
	DIM  = "\033[2m"

	FG_BLACK = "\033[30m"
	FG_WHITE = "\033[97m"

	BG_DGRAY  = "\033[100m"
	BG_RED    = "\033[41m"
	BG_GREEN  = "\033[42m"
	BG_YELLOW = "\033[43m"
	BG_LBLUE  = "\033[104m"

	RESET = "\033[0m"
)

var (
	labels = map[int]string{
		DEBUG:     "DBG",
		INFO:      "INF",
		IMPORTANT: "IMP",
		WARNING:   "WAR",
		ERROR:     "ERR",
		FATAL:     "!!!",
	}
	colors = map[int]string{
		DEBUG:     DIM + FG_BLACK + BG_DGRAY,
		INFO:      FG_WHITE + BG_GREEN,
		IMPORTANT: FG_WHITE + BG_LBLUE,
		WARNING:   FG_WHITE + BG_YELLOW,
		ERROR:     FG_WHITE + BG_RED,
		FATAL:     FG_WHITE + BG_RED + BOLD,
	}
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
	label := labels[log.Level]
	color := colors[log.Level]
	return color + label + RESET
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

	p.Add("sys.log", LogMessage{
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
