package network

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/v2/core"
)

type Meta struct {
	sync.Mutex
	m map[string]any
}

// we want to protect concurrent access to the Meta
// object so the m field needs to be unexported, this
// is to have it in JSON regardless.
type metaJSON struct {
	Values map[string]any `json:"values"`
}

func NewMeta() *Meta {
	return &Meta{
		m: make(map[string]any),
	}
}

func (m *Meta) MarshalJSON() ([]byte, error) {
	m.Lock()
	defer m.Unlock()
	return json.Marshal(metaJSON{Values: m.m})
}

func (m *Meta) Set(name string, value any) {
	m.Lock()
	defer m.Unlock()
	m.m[name] = value
}

func (m *Meta) Get(name string) any {
	m.Lock()
	defer m.Unlock()

	if v, found := m.m[name]; found {
		return v
	}
	return ""
}

func (m *Meta) GetIntsWith(name string, with int, sorted bool) []int {
	sints := strings.Split(m.Get(name).(string), ",")
	ints := []int{with}

	for _, s := range sints {
		n, err := strconv.Atoi(s)
		if err == nil {
			ints = append(ints, n)
		}
	}

	return core.UniqueInts(ints, sorted)
}

func (m *Meta) SetInts(name string, ints []int) {
	list := make([]string, len(ints))
	for i, n := range ints {
		list[i] = fmt.Sprintf("%d", n)
	}

	m.Set(name, strings.Join(list, ","))
}

func (m *Meta) GetOr(name string, dflt any) any {
	m.Lock()
	defer m.Unlock()

	if v, found := m.m[name]; found {
		return v
	}
	return dflt
}

func (m *Meta) Each(cb func(name string, value any)) {
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
