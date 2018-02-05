package session

import (
	"fmt"
	"strings"
	"sync"

	"github.com/evilsocket/bettercap-ng/core"
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
	Name       string      `json:"name"`
	Session    *Session    `json:"-"`
	Started    bool        `json:"started"`
	StatusLock *sync.Mutex `json:"-"`

	handlers []ModuleHandler
	params   map[string]*ModuleParam
}

func NewSessionModule(name string, s *Session) SessionModule {
	m := SessionModule{
		Name:       name,
		Session:    s,
		Started:    false,
		StatusLock: &sync.Mutex{},

		handlers: make([]ModuleHandler, 0),
		params:   make(map[string]*ModuleParam),
	}

	return m
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
			part = core.Trim(part)
			if part != "" {
				values = append(values, part)
			}
		}
	}
	return
}

func (m SessionModule) StringParam(name string) (error, string) {
	if p, found := m.params[name]; found == true {
		if err, v := p.Get(m.Session); err != nil {
			return err, ""
		} else {
			return nil, v.(string)
		}
	} else {
		return fmt.Errorf("Parameter %s does not exist.", name), ""
	}
}

func (m SessionModule) IntParam(name string) (error, int) {
	if p, found := m.params[name]; found == true {
		if err, v := p.Get(m.Session); err != nil {
			return err, 0
		} else {
			return nil, v.(int)
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
	m.StatusLock.Lock()
	defer m.StatusLock.Unlock()
	return m.Started
}

func (m *SessionModule) SetRunning(running bool) {
	m.StatusLock.Lock()
	defer m.StatusLock.Unlock()
	m.Started = running

	if running {
		m.Session.Events.Add("mod.started", m.Name)
	} else {
		m.Session.Events.Add("mod.stopped", m.Name)
	}
}
