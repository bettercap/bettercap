package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
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

func (e Event) Print() {
	fmt.Printf("[%s] [%s] [%s] %+v\n", e.Time, core.Bold("event"), core.Green(e.Tag), e.Data)
}

type EventPool struct {
	events []Event
	lock   *sync.Mutex
}

func NewEventPool() *EventPool {
	return &EventPool{
		events: make([]Event, 0),
		lock:   &sync.Mutex{},
	}
}

func (p *EventPool) Add(tag string, data interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	e := NewEvent(tag, data)
	p.events = append([]Event{e}, p.events...)
	e.Print()
}

func (p *EventPool) Events() []Event {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.events
}
