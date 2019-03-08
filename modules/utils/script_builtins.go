package utils

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/plugin"

	"github.com/robertkrimen/otto"
)

var nullOtto = otto.Value{}

func errOtto(format string, args ...interface{}) otto.Value {
	log.Error(format, args...)
	return nullOtto
}

func init() {
	// used to read a directory (returns string array)
	plugin.Defines["readDir"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("readDir: expected 1 argument, %d given instead.", argc)
		}

		path := argv[0].String()
		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return errOtto("Could not read directory %s: %s", path, err)
		}

		entry_list := []string{}
		for _, file := range dir {
			entry_list = append( entry_list, file.Name() )
		}

		v, err := otto.Otto.ToValue(*call.Otto, entry_list)
		if err != nil {
			return errOtto("Could not convert to array: %s", err)
		}

		return v
	}

	// used to read a file ... doh
	plugin.Defines["readFile"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("readFile: expected 1 argument, %d given instead.", argc)
		}

		filename := argv[0].String()
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			return errOtto("Could not read file %s: %s", filename, err)
		}

		v, err := otto.ToValue(string(raw))
		if err != nil {
			return errOtto("Could not convert to string: %s", err)
		}
		return v
	}

	plugin.Defines["writeFile"] = func(call otto.FunctionCall) otto.Value {
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
	}

	// log something
	plugin.Defines["log"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	}

	// log debug
	plugin.Defines["log_debug"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Debug("%s", v.String())
		}
		return otto.Value{}
	}

	// log info
	plugin.Defines["log_info"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	}

	// log warning
	plugin.Defines["log_warn"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Warning("%s", v.String())
		}
		return otto.Value{}
	}

	// log error
	plugin.Defines["log_error"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Error("%s", v.String())
		}
		return otto.Value{}
	}

	// log fatal
	plugin.Defines["log_fatal"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Fatal("%s", v.String())
		}
		return otto.Value{}
	}

	// javascript btoa function
	plugin.Defines["btoa"] = func(call otto.FunctionCall) otto.Value {
		varValue := base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String()))
		v, err := otto.ToValue(varValue)
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	}

	// javascript atob function
	plugin.Defines["atob"] = func(call otto.FunctionCall) otto.Value {
		varValue, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return errOtto("Could not decode string: %s", call.Argument(0).String())
		}
		v, err := otto.ToValue(string(varValue))
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	}

	// read or write environment variable
	plugin.Defines["env"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)

		if argc == 1 {
			// get
			varName := call.Argument(0).String()
			if found, varValue := session.I.Env.Get(varName); found {
				v, err := otto.ToValue(varValue)
				if err != nil {
					return errOtto("Could not convert to string: %s", varValue)
				}
				return v
			}

		} else if argc == 2 {
			// set
			varName := call.Argument(0).String()
			varValue := call.Argument(1).String()
			session.I.Env.Set(varName, varValue)
		} else {
			return errOtto("env: expected 1 or 2 arguments, %d given instead.", argc)
		}

		return nullOtto
	}
}
