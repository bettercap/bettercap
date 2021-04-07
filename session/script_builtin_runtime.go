package session

import (
	"encoding/json"

	"github.com/bettercap/bettercap/js"
	"github.com/evilsocket/islazy/log"
	"github.com/robertkrimen/otto"
)

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

				// lock vm
				I.script.Lock()
				if _, err := cb.Call(otto.NullValue(), opaque); err != nil {
					I.Events.Log(log.ERROR, "error dispatching event %s: %v", event.Tag, err)
				}
				I.script.Unlock()
			}
		}
	}(filterExpr, cb)

	return js.NullValue
}
