package network

import (
	"encoding/json"
	"sync"
)

type Meta struct {
	sync.Mutex
	m map[string]interface{}
}

// we want to protect concurrent access to the Meta
// object so the m field needs to be unexported, this
// is to have it in JSON regardless.
type metaJSON struct {
	Values map[string]interface{} `json:"values"`
}

func NewMeta() *Meta {
	return &Meta{
		m: make(map[string]interface{}),
	}
}

func (m *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(metaJSON{Values: m.m})
}

func (m *Meta) Set(name string, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.m[name] = value
}

func (m *Meta) Get(name string) interface{} {
	m.Lock()
	defer m.Unlock()

	if v, found := m.m[name]; found == true {
		return v
	}
	return ""
}

func (m *Meta) Each(cb func(name string, value interface{})) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.m {
		cb(k, v)
	}
}

func (m *Meta) Empty() bool {
	m.Lock()
	defer m.Unlock()
	return len(m.m) == 0
}
