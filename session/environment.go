package session

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

type Environment struct {
	Padding int
	storage map[string]string
	lock    *sync.Mutex
}

func NewEnvironment() *Environment {
	env := &Environment{
		Padding: 0,
		storage: make(map[string]string),
		lock:    &sync.Mutex{},
	}

	return env
}

func (env *Environment) Storage() *map[string]string {
	return &env.storage
}

func (env *Environment) Has(name string) bool {
	env.lock.Lock()
	defer env.lock.Unlock()

	_, found := env.storage[name]

	return found
}

func (env *Environment) Set(name, value string) string {
	env.lock.Lock()
	defer env.lock.Unlock()

	old, _ := env.storage[name]
	env.storage[name] = value

	width := len(name)
	if width > env.Padding {
		env.Padding = width
	}

	return old
}

func (env *Environment) Get(name string) (bool, string) {
	env.lock.Lock()
	defer env.lock.Unlock()

	if value, found := env.storage[name]; found == true {
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
	for k := range env.storage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
