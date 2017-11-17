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
	Session    *Session
	Started    bool
	StatusLock *sync.Mutex

	handlers []ModuleHandler
	params   map[string]*ModuleParam
}

func NewSessionModule(s *Session) SessionModule {
	m := SessionModule{
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
}

func (m *SessionModule) OnSessionStarted(s *Session) {

}
