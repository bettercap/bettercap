package session

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

type Module interface {
	Name() string
	Description() string
	Author() string
	Handlers() []ModuleHandler
	Parameters() map[string]*ModuleParam

	Running() bool
	Start() error
	Stop() error
}

type SessionModule struct {
	Name       string        `json:"name"`
	Session    *Session      `json:"-"`
	Started    bool          `json:"started"`
	StatusLock *sync.RWMutex `json:"-"`

	handlers []ModuleHandler
	params   map[string]*ModuleParam
	tag      string
}

func AsTag(name string) string {
	return fmt.Sprintf("%s ", tui.Wrap(tui.BACKLIGHTBLUE, tui.Wrap(tui.FOREBLACK, name)))
}

func NewSessionModule(name string, s *Session) SessionModule {
	m := SessionModule{
		Name:       name,
		Session:    s,
		Started:    false,
		StatusLock: &sync.RWMutex{},

		handlers: make([]ModuleHandler, 0),
		params:   make(map[string]*ModuleParam),
		tag:      AsTag(name),
	}

	return m
}

func (m *SessionModule) Debug(format string, args ...interface{}) {
	m.Session.Events.Log(log.DEBUG, m.tag+format, args...)
}

func (m *SessionModule) Info(format string, args ...interface{}) {
	m.Session.Events.Log(log.INFO, m.tag+format, args...)
}

func (m *SessionModule) Warning(format string, args ...interface{}) {
	m.Session.Events.Log(log.WARNING, m.tag+format, args...)
}

func (m *SessionModule) Error(format string, args ...interface{}) {
	m.Session.Events.Log(log.ERROR, m.tag+format, args...)
}

func (m *SessionModule) Fatal(format string, args ...interface{}) {
	m.Session.Events.Log(log.FATAL, m.tag+format, args...)
}

func (m *SessionModule) Handlers() []ModuleHandler {
	return m.handlers
}

func (m *SessionModule) Parameters() map[string]*ModuleParam {
	return m.params
}

func (m *SessionModule) Param(name string) *ModuleParam {
	return m.params[name]
}

func (m SessionModule) ListParam(name string) (err error, values []string) {
	values = make([]string, 0)
	list := ""
	if err, list = m.StringParam(name); err != nil {
		return
	} else {
		parts := strings.Split(list, ",")
		for _, part := range parts {
			part = str.Trim(part)
			if part != "" {
				values = append(values, part)
			}
		}
	}
	return
}

func (m SessionModule) StringParam(name string) (error, string) {
	if p, found := m.params[name]; found {
		if err, v := p.Get(m.Session); err != nil {
			return err, ""
		} else {
			return nil, v.(string)
		}
	} else {
		return fmt.Errorf("Parameter %s does not exist.", name), ""
	}
}

func (m SessionModule) IPParam(name string) (error, net.IP) {
	if err, v := m.StringParam(name); err != nil {
		return err, nil
	} else {
		return nil, net.ParseIP(v)
	}
}

func (m SessionModule) IntParam(name string) (error, int) {
	if p, found := m.params[name]; found {
		if err, v := p.Get(m.Session); err != nil {
			return err, 0
		} else {
			return nil, v.(int)
		}

	} else {
		return fmt.Errorf("Parameter %s does not exist.", name), 0
	}
}

func (m SessionModule) DecParam(name string) (error, float64) {
	if p, found := m.params[name]; found {
		if err, v := p.Get(m.Session); err != nil {
			return err, 0
		} else {
			return nil, v.(float64)
		}

	} else {
		return fmt.Errorf("Parameter %s does not exist.", name), 0
	}
}

func (m SessionModule) BoolParam(name string) (error, bool) {
	if err, v := m.params[name].Get(m.Session); err != nil {
		return err, false
	} else {
		return nil, v.(bool)
	}
}

func (m *SessionModule) AddHandler(h ModuleHandler) {
	m.handlers = append(m.handlers, h)
}

func (m *SessionModule) AddParam(p *ModuleParam) *ModuleParam {
	m.params[p.Name] = p
	p.Register(m.Session)
	return p
}

func (m *SessionModule) Running() bool {
	m.StatusLock.RLock()
	defer m.StatusLock.RUnlock()
	return m.Started
}

func (m *SessionModule) SetRunning(running bool, cb func()) error {
	if running == m.Running() {
		if m.Started {
			return ErrAlreadyStarted
		} else {
			return ErrAlreadyStopped
		}
	}

	m.StatusLock.Lock()
	m.Started = running
	m.StatusLock.Unlock()

	if *m.Session.Options.Debug {
		if running {
			m.Session.Events.Add("mod.started", m.Name)
		} else {
			m.Session.Events.Add("mod.stopped", m.Name)
		}
	}

	if cb != nil {
		if running {
			// this is the worker, start async
			go cb()
		} else {
			// stop callback, this is sync with a 10 seconds timeout
			done := make(chan bool, 1)
			go func() {
				cb()
				done <- true
			}()

			select {
			case <-done:
				return nil
			case <-time.After(10 * time.Second):
				fmt.Printf("%s: Stopping module %s timed out.\n", tui.Yellow(tui.Bold("WARNING")), m.Name)
			}
		}
	}

	return nil
}
