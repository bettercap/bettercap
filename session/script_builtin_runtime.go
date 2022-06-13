package session

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/bettercap/bettercap/js"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/robertkrimen/otto"
)

// see https://github.com/robertkrimen/otto/issues/213
var jsRuntime = otto.New()

func jsRunFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return js.ReportError("run accepts one string argument")
	} else if argv[0].IsString() == false {
		return js.ReportError("run accepts one string argument")
	}

	for _, cmd := range ParseCommands(argv[0].String()) {
		if err := I.Run(cmd); err != nil {
			return js.ReportError("error running '%s': %v", cmd, err)
		}
	}

	return js.NullValue
}

func jsOnEventFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	cb := otto.NullValue()
	filterExpr := ""

	// just one argument, a function to receive all events
	if argc == 1 {
		if argv[0].IsFunction() == false {
			return js.ReportError("the single argument must be a function")
		}
		cb = argv[0]
	} else {
		if argc != 2 {
			return js.ReportError("expected two arguments (event_name, callback), got %d", argc)
		} else if argv[0].IsString() == false {
			return js.ReportError("first argument must be a string")
		} else if argv[1].IsFunction() == false {
			return js.ReportError("second argument must be a function")
		}

		filterExpr = argv[0].String()
		cb = argv[1]
	}

	// start a go routine for this event listener
	go func(expr string, cb otto.Value) {
		listener := I.Events.Listen()
		defer I.Events.Unlisten(listener)

		for event := range listener {
			if expr == "" || event.Tag == expr {
				// some objects don't do well with js, so convert them to a generic map
				// before passing them to the callback
				var opaque interface{}
				if raw, err := json.Marshal(event); err != nil {
					I.Events.Log(log.ERROR, "error serializing event %s: %v", event.Tag, err)
				} else if err = json.Unmarshal(raw, &opaque); err != nil {
					I.Events.Log(log.ERROR, "error serializing event %s: %v", event.Tag, err)
				}

				// lock vm if ready and available
				locked := false
				if I.script != nil {
					I.script.Lock()
					locked = true
				}

				if _, err := cb.Call(otto.NullValue(), opaque); err != nil {
					I.Events.Log(log.ERROR, "error dispatching event %s: %v", event.Tag, err)
				}

				// unlock vm if ready and available
				if locked {
					I.script.Unlock()
				}
			}
		}
	}(filterExpr, cb)

	return js.NullValue
}

func jsSaveJSONFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 2 {
		return js.ReportError("saveJSON accepts one object and one string arguments")
	} else if argv[0].IsObject() == false {
		return js.ReportError("saveJSON accepts one object and one string arguments")
	} else if argv[1].IsString() == false {
		return js.ReportError("saveJSON accepts one object and one string arguments")
	}

	obj := argv[0]
	if fileName, err := fs.Expand(argv[1].String()); err != nil {
		return js.ReportError("can't expand '%s': %v", fileName, err)
	} else if exp, err := obj.Export(); err != nil {
		return js.ReportError("error exporting object: %v", err)
	} else if raw, err := json.Marshal(exp); err != nil {
		return js.ReportError("error serializing object: %v", err)
	} else if err = ioutil.WriteFile(fileName, raw, os.ModePerm); err != nil {
		return js.ReportError("error writing to '%s': %v", fileName, err)
	}

	return js.NullValue
}

func jsLoadJSONFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return js.ReportError("loadJSON accepts one string argument")
	} else if argv[0].IsString() == false {
		return js.ReportError("loadJSON accepts one string argument")
	}

	var obj interface{}

	if fileName, err := fs.Expand(argv[0].String()); err != nil {
		return js.ReportError("can't expand '%s': %v", fileName, err)
	} else if rawData, err := ioutil.ReadFile(fileName); err != nil {
		return js.ReportError("can't read '%s': %v", fileName, err)
	} else if err = json.Unmarshal(rawData, &obj); err != nil {
		return js.ReportError("can't parse '%s': %v", fileName, err)
	} else if v, err := jsRuntime.ToValue(obj); err != nil {
		return js.ReportError("could not convert '%s' to javascript object: %s", fileName, err)
	} else {
		return v
	}
}

func jsFileExistsFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return js.ReportError("fileExists accepts one string argument")
	} else if argv[0].IsString() == false {
		return js.ReportError("fileExists accepts one string argument")
	} else if fileName, err := fs.Expand(argv[0].String()); err != nil {
		return js.ReportError("can't expand '%s': %v", fileName, err)
	} else if fs.Exists(fileName) {
		return otto.TrueValue()
	}

	return otto.FalseValue()
}
