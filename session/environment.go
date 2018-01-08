package session

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

type Environment struct {
	Padding int               `json:"-"`
	Storage map[string]string `json:"storage"`
	lock    *sync.Mutex
	sess    *Session
}

func NewEnvironment(s *Session) *Environment {
	env := &Environment{
		Padding: 0,
		Storage: make(map[string]string),
		lock:    &sync.Mutex{},
		sess:    s,
	}

	return env
}

func (env *Environment) Has(name string) bool {
	env.lock.Lock()
	defer env.lock.Unlock()

	_, found := env.Storage[name]

	return found
}

func (env *Environment) Set(name, value string) string {
	env.lock.Lock()
	defer env.lock.Unlock()

	old, _ := env.Storage[name]
	env.Storage[name] = value

	env.sess.Events.Add("env.change", struct {
		name  string
		value string
	}{
		name,
		value,
	})

	width := len(name)
	if width > env.Padding {
		env.Padding = width
	}

	return old
}

func (env *Environment) Get(name string) (bool, string) {
	env.lock.Lock()
	defer env.lock.Unlock()

	if value, found := env.Storage[name]; found == true {
		return true, value
	}

	return false, ""
}

func (env *Environment) GetInt(name string) (error, int) {
	if found, value := env.Get(name); found == true {
		if i, err := strconv.Atoi(value); err == nil {
			return nil, i
		} else {
			return err, 0
		}
	}

	return fmt.Errorf("Not found."), 0
}

func (env *Environment) Sorted() []string {
	env.lock.Lock()
	defer env.lock.Unlock()

	var keys []string
	for k := range env.Storage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
