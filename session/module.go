package session

import (
	"encoding/json"
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

	Extra() map[string]interface{}
	Required() []string
	Running() bool
	Start() error
	Stop() error
}

type ModuleList []Module

type moduleJSON struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Author      string                  `json:"author"`
	Parameters  map[string]*ModuleParam `json:"parameters"`
	Handlers    []ModuleHandler         `json:"handlers"`
	Running     bool                    `json:"running"`
	State       map[string]interface{}  `json:"state"`
}

func (mm ModuleList) MarshalJSON() ([]byte, error) {
	mods := []moduleJSON{}
	for _, m := range mm {
		mJSON := moduleJSON{
			Name:        m.Name(),
			Description: m.Description(),
			Author:      m.Author(),
			Parameters:  m.Parameters(),
			Handlers:    m.Handlers(),
			Running:     m.Running(),
			State:       m.Extra(),
		}
		mods = append(mods, mJSON)
	}
	return json.Marshal(mods)
}

type SessionModule struct {
	Name       string
	Session    *Session
	Started    bool
	StatusLock *sync.RWMutex
	State      *sync.Map

	handlers []ModuleHandler
	params   map[string]*ModuleParam
	requires []string
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
		State:      &sync.Map{},

		requires: make([]string, 0),
		handlers: make([]ModuleHandler, 0),
		params:   make(map[string]*ModuleParam),
		tag:      AsTag(name),
	}

	return m
}

func (m *SessionModule) Extra() map[string]interface{} {
	extra := make(map[string]interface{})
	m.State.Range(func(k, v interface{}) bool {
		extra[k.(string)] = v
		return true
	})
	return extra
}

func (m *SessionModule) InitState(keys ...string) {
	for _, key := range keys {
		m.State.Store(key, nil)
	}
}

func (m *SessionModule) ResetState() {
	m.State.Range(func(k, v interface{}) bool {
		m.State.Store(k, nil)
		return true
	})
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

func (m *SessionModule) Requires(modName string) {
	m.requires = append(m.requires, modName)
}

func (m *SessionModule) Required() []string {
	return m.requires
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
		if v, err := p.Get(m.Session); err != nil {
			return err, ""
		} else {
			return nil, v.(string)
		}
	} else {
		return fmt.Errorf("Parameter %s does not exist.", name), ""
	}
}

func (m SessionModule) IPParam(name string) (net.IP, error) {
	if err, v := m.StringParam(name); err != nil {
		return nil, err
	} else {
		return net.ParseIP(v), err
	}
}

func (m SessionModule) IntParam(name string) (int, error) {
	if p, found := m.params[name]; found {
		if v, err := p.Get(m.Session); err != nil {
			return 0, err
		} else {
			return v.(int), nil
		}

	} else {
		return 0, fmt.Errorf("Parameter %s does not exist.", name)
	}
}

func (m SessionModule) DecParam(name string) (float64, error) {
	if p, found := m.params[name]; found {
		if v, err := p.Get(m.Session); err != nil {
			return 0, err
		} else {
			return v.(float64), err
		}

	} else {
		return 0, fmt.Errorf("Parameter %s does not exist.", name)
	}
}

func (m SessionModule) BoolParam(name string) (bool, error) {
	if v, err := m.params[name].Get(m.Session); err != nil {
		return false, err
	} else {
		return v.(bool), nil
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

func (m *SessionModule) AddObservableParam(p *ModuleParam, cb EnvironmentChangedCallback) *ModuleParam {
	m.params[p.Name] = p
	p.RegisterObserver(m.Session, cb)
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
			return ErrAlreadyStarted(m.Name)
		} else {
			return ErrAlreadyStopped(m.Name)
		}
	}

	if running == true {
		for _, modName := range m.Required() {
			if m.Session.IsOn(modName) == false {
				m.Info("starting %s as a requirement for %s", modName, m.Name)
				if err := m.Session.Run(modName + " on"); err != nil {
					return fmt.Errorf("error while starting module %s as a requirement for %s: %v", modName, m.Name, err)
				}
			}
		}
	}

	m.StatusLock.Lock()
	m.Started = running
	m.StatusLock.Unlock()

	if running {
		m.Session.Events.Add("mod.started", m.Name)
	} else {
		m.Session.Events.Add("mod.stopped", m.Name)
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
