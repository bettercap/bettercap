package session

import "sync"

type Module interface {
	Name() string
	Description() string
	Author() string
	Handlers() []ModuleHandler
	Parameters() map[string]*ModuleParam

	Running() bool
	Start() error
	Stop() error

	OnSessionStarted(s *Session)
	OnSessionEnded(s *Session)
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

func (m *SessionModule) AddHandler(h ModuleHandler) {
	m.handlers = append(m.handlers, h)
}

func (m *SessionModule) AddParam(p *ModuleParam) {
	m.params[p.Name] = p
	p.Register(m.Session)
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

func (m *SessionModule) OnSessionStarted(s *Session) {

}
