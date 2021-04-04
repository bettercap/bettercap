package session

import (
	"github.com/bettercap/bettercap/js"
	"github.com/robertkrimen/otto"
)

func jsEnvFunc(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)

	if argc == 1 {
		// get
		varName := call.Argument(0).String()
		if found, varValue := I.Env.Get(varName); found {
			v, err := otto.ToValue(varValue)
			if err != nil {
				return js.ReportError("could not convert to string: %s", varValue)
			}
			return v
		}

	} else if argc == 2 {
		// set
		varName := call.Argument(0).String()
		varValue := call.Argument(1).String()
		I.Env.Set(varName, varValue)
	} else {
		return js.ReportError("env: expected 1 or 2 arguments, %d given instead.", argc)
	}
	return js.NullValue
}

