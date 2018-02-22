package modules

import (
	"io/ioutil"

	"github.com/evilsocket/bettercap-ng/log"

	"github.com/robertkrimen/otto"
)

// define functions available to proxy scripts
func (s *ProxyScript) defineBuiltins() error {
	// used to read a file ... doh
	s.VM.Set("readFile", func(call otto.FunctionCall) otto.Value {
		filename := call.Argument(0).String()
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Error("Could not read %s: %s", filename, err)
			return otto.Value{}
		}

		v, err := s.VM.ToValue(string(raw))
		if err != nil {
			log.Error("Could not convert to string: %s", err)
			return otto.Value{}
		}
		return v
	})

	// log something
	s.VM.Set("log", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
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
					log.Error("Could not convert to string: %s", varValue)
					return otto.Value{}
				}
				return v
			}

		} else if argc == 2 {
			// set
			varName := call.Argument(0).String()
			varValue := call.Argument(1).String()
			s.sess.Env.Set(varName, varValue)
		}

		return otto.Value{}
	})

	return nil
}
