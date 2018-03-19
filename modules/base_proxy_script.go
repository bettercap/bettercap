package modules

import (
	"encoding/base64"
	"io/ioutil"
	"sync"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/robertkrimen/otto"
)

var nullOtto = otto.Value{}

func errOtto(format string, args ...interface{}) otto.Value {
	log.Error(format, args...)
	return nullOtto
}

type ProxyScript struct {
	sync.Mutex

	Path   string
	Source string
	VM     *otto.Otto

	sess        *session.Session
	cbCacheLock *sync.Mutex
	cbCache     map[string]bool
}

func LoadProxyScriptSource(path, source string, sess *session.Session) (err error, s *ProxyScript) {
	s = &ProxyScript{
		Path:        path,
		Source:      source,
		VM:          otto.New(),
		sess:        sess,
		cbCacheLock: &sync.Mutex{},
		cbCache:     make(map[string]bool),
	}

	// this will define callbacks and global objects
	_, err = s.VM.Run(s.Source)
	if err != nil {
		return
	}

	// define session pointer
	err = s.VM.Set("env", sess.Env.Data)
	if err != nil {
		log.Error("Error while defining environment: %s", err)
		return
	}

	err = s.defineBuiltins()
	if err != nil {
		log.Error("Error while defining builtin functions: %s", err)
		return
	}

	// run onLoad if defined
	if s.hasCallback("onLoad") {
		_, err = s.VM.Run("onLoad()")
		if err != nil {
			log.Error("Error while executing onLoad callback: %s", err)
			return
		}
	}

	return
}

func LoadProxyScript(path string, sess *session.Session) (err error, s *ProxyScript) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return LoadProxyScriptSource(path, string(raw), sess)
}

func (s *ProxyScript) defineBuiltins() error {
	// used to read a file ... doh
	s.VM.Set("readFile", func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("readFile: expected 1 argument, %d given instead.", argc)
		}

		filename := argv[0].String()
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			return errOtto("Could not read %s: %s", filename, err)
		}

		v, err := s.VM.ToValue(string(raw))
		if err != nil {
			return errOtto("Could not convert to string: %s", err)
		}
		return v
	})

	s.VM.Set("writeFile", func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 2 {
			return errOtto("writeFile: expected 2 arguments, %d given instead.", argc)
		}

		filename := argv[0].String()
		data := argv[1].String()

		err := ioutil.WriteFile(filename, []byte(data), 0644)
		if err != nil {
			return errOtto("Could not write %d bytes to %s: %s", len(data), filename, err)
		}

		return otto.NullValue()
	})

	// log something
	s.VM.Set("log", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	})

	// log debug
	s.VM.Set("log_debug", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Debug("%s", v.String())
		}
		return otto.Value{}
	})

	// log info
	s.VM.Set("log_info", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	})

	// log warning
	s.VM.Set("log_warn", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Warning("%s", v.String())
		}
		return otto.Value{}
	})

	// log error
	s.VM.Set("log_error", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Error("%s", v.String())
		}
		return otto.Value{}
	})

	// log fatal
	s.VM.Set("log_fatal", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Fatal("%s", v.String())
		}
		return otto.Value{}
	})

	// javascript btoa function
	s.VM.Set("btoa", func(call otto.FunctionCall) otto.Value {
		varValue := base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String()))
		v, err := s.VM.ToValue(varValue)
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	})

	// javascript atob function
	s.VM.Set("atob", func(call otto.FunctionCall) otto.Value {
		varValue, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return errOtto("Could not decode string: %s", call.Argument(0).String())
		}
		v, err := s.VM.ToValue(string(varValue))
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	})

	// read or write environment variable
	s.VM.Set("env", func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)

		if argc == 1 {
			// get
			varName := call.Argument(0).String()
			if found, varValue := s.sess.Env.Get(varName); found == true {
				v, err := s.VM.ToValue(varValue)
				if err != nil {
					return errOtto("Could not convert to string: %s", varValue)
				}
				return v
			}

		} else if argc == 2 {
			// set
			varName := call.Argument(0).String()
			varValue := call.Argument(1).String()
			s.sess.Env.Set(varName, varValue)
		} else {
			return errOtto("env: expected 1 or 2 arguments, %d given instead.", argc)
		}

		return nullOtto
	})

	return nil
}

func (s *ProxyScript) hasCallback(name string) bool {
	s.cbCacheLock.Lock()
	defer s.cbCacheLock.Unlock()

	// check the cache
	has, found := s.cbCache[name]
	if found == false {
		// check the VM
		cb, err := s.VM.Get(name)
		if err == nil && cb.IsFunction() {
			has = true
		} else {
			has = false
		}
		s.cbCache[name] = has
	}

	return has
}
