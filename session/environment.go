package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"

	"github.com/evilsocket/islazy/fs"
)

type EnvironmentChangedCallback func(newValue string)

type Environment struct {
	sync.Mutex
	Data map[string]string `json:"data"`
	cbs  map[string]EnvironmentChangedCallback
}

func NewEnvironment(envFile string) (*Environment, error) {
	env := &Environment{
		Data: make(map[string]string),
		cbs:  make(map[string]EnvironmentChangedCallback),
	}

	if envFile != "" {
		envFile, _ := fs.Expand(envFile)
		if fs.Exists(envFile) {
			if err := env.Load(envFile); err != nil {
				return nil, err
			}
		}
	}

	return env, nil
}

func (env *Environment) Load(fileName string) error {
	env.Lock()
	defer env.Unlock()

	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	if len(raw) > 0 {
		return json.Unmarshal(raw, &env.Data)
	}
	return nil
}

func (env *Environment) Save(fileName string) error {
	env.Lock()
	defer env.Unlock()

	raw, err := json.Marshal(env.Data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, raw, 0644)
}

func (env *Environment) Has(name string) bool {
	env.Lock()
	defer env.Unlock()

	_, found := env.Data[name]

	return found
}

func (env *Environment) addCb(name string, cb EnvironmentChangedCallback) {
	env.Lock()
	defer env.Unlock()
	env.cbs[name] = cb
}

func (env *Environment) WithCallback(name, value string, cb EnvironmentChangedCallback) string {
	env.addCb(name, cb)
	ret := env.Set(name, value)
	return ret
}

func (env *Environment) Set(name, value string) string {
	env.Lock()

	old := env.Data[name]
	env.Data[name] = value

	env.Unlock()

	if cb, hasCallback := env.cbs[name]; hasCallback {
		cb(value)
	}

	return old
}

func (env *Environment) GetUnlocked(name string) (bool, string) {
	if value, found := env.Data[name]; found {
		return true, value
	}
	return false, ""
}

func (env *Environment) Get(name string) (bool, string) {
	env.Lock()
	defer env.Unlock()
	return env.GetUnlocked(name)
}

func (env *Environment) GetInt(name string) (error, int) {
	if found, value := env.Get(name); found {
		if i, err := strconv.Atoi(value); err == nil {
			return nil, i
		} else {
			return err, 0
		}
	}

	return fmt.Errorf("Not found."), 0
}

func (env *Environment) Sorted() []string {
	env.Lock()
	defer env.Unlock()

	var keys []string
	for k := range env.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
