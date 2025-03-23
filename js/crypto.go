package js

import (
	"crypto/sha1"

	"github.com/robertkrimen/otto"
)

func cryptoSha1(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("Crypto.sha1: expected 1 argument, %d given instead.", argc)
	}

	arg := argv[0]
	if (!arg.IsString()) {
		return ReportError("Crypto.sha1: single argument must be a string.")
	}

	hasher := sha1.New()
	hasher.Write([]byte(arg.String()))
	v, err := otto.ToValue(string(hasher.Sum(nil)))
	if err != nil {
		return ReportError("Crypto.sha1: could not convert to string: %s", err)
	}

	return v
}
