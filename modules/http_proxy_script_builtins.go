package modules

import (
	"fmt"
	"io/ioutil"

	"github.com/evilsocket/bettercap-ng/log"

	"github.com/robertkrimen/otto"
)

// define functions available to proxy scripts
func (s *ProxyScript) defineBuiltins() error {
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

	s.VM.Set("log", func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			fmt.Printf("%s", v.String())
		}
		fmt.Println()
		s.sess.Refresh()

		return otto.Value{}
	})

	return nil
}
