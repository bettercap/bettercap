package js

import (
	"github.com/evilsocket/islazy/log"
	"github.com/robertkrimen/otto"
)

func flog(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Info("%s", v.String())
	}
	return otto.Value{}
}

func log_debug(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Debug("%s", v.String())
	}
	return otto.Value{}
}

func log_info(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Info("%s", v.String())
	}
	return otto.Value{}
}

func log_warn(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Warning("%s", v.String())
	}
	return otto.Value{}
}

func log_error(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Error("%s", v.String())
	}
	return otto.Value{}
}

func log_fatal(call otto.FunctionCall) otto.Value {
	for _, v := range call.ArgumentList {
		log.Fatal("%s", v.String())
	}
	return otto.Value{}
}
