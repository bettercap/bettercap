package modules

import (
	"io/ioutil"
	"net/http"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/robertkrimen/otto"
)

type HttpProxyScript struct {
	*ProxyScript
	onRequestScript  *otto.Script
	onResponseScript *otto.Script
}

func LoadHttpProxyScriptSource(path, source string, sess *session.Session) (err error, s *HttpProxyScript) {
	err, ps := LoadProxyScriptSource(path, source, sess)
	if err != nil {
		return
	}

	s = &HttpProxyScript{
		ProxyScript:      ps,
		onRequestScript:  nil,
		onResponseScript: nil,
	}

	// compile call to onRequest if defined
	if s.hasCallback("onRequest") {
		s.onRequestScript, err = s.VM.Compile("", "onRequest(req, res)")
		if err != nil {
			log.Error("Error while compiling onRequest callback: %s", err)
			return
		}
	}

	// compile call to onResponse if defined
	if s.hasCallback("onResponse") {
		s.onResponseScript, err = s.VM.Compile("", "onResponse(req, res)")
		if err != nil {
			log.Error("Error while compiling onResponse callback: %s", err)
			return
		}
	}

	return
}

func LoadHttpProxyScript(path string, sess *session.Session) (err error, s *HttpProxyScript) {
	log.Info("Loading proxy script %s ...", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return LoadHttpProxyScriptSource(path, string(raw), sess)
}

func (s *HttpProxyScript) doRequestDefines(req *http.Request) (err error, jsres *JSResponse) {
	// convert request and define empty response to be optionally filled
	jsreq := NewJSRequest(req)
	if err = s.VM.Set("req", jsreq); err != nil {
		log.Error("Error while defining request: %s", err)
		return
	}

	jsres = NewJSResponse(nil)
	if err = s.VM.Set("res", jsres); err != nil {
		log.Error("Error while defining response: %s", err)
		return
	}
	return
}

func (s *HttpProxyScript) doResponseDefines(res *http.Response) (err error, jsres *JSResponse) {
	// convert both request and response
	jsreq := NewJSRequest(res.Request)
	if err = s.VM.Set("req", jsreq); err != nil {
		log.Error("Error while defining request: %s", err)
		return
	}

	jsres = NewJSResponse(res)
	if err = s.VM.Set("res", jsres); err != nil {
		log.Error("Error while defining response: %s", err)
		return
	}

	return
}

func (s *HttpProxyScript) OnRequest(req *http.Request) *JSResponse {
	if s.onRequestScript != nil {
		s.Lock()
		defer s.Unlock()

		err, jsres := s.doRequestDefines(req)
		if err != nil {
			log.Error("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run(s.onRequestScript)
		if err != nil {
			log.Error("Error while executing onRequest callback: %s", err)
			return nil
		}

		if jsres.WasModified() {
			jsres.UpdateHash()
			return jsres
		}
	}

	return nil
}

func (s *HttpProxyScript) OnResponse(res *http.Response) *JSResponse {
	if s.onResponseScript != nil {
		s.Lock()
		defer s.Unlock()

		err, jsres := s.doResponseDefines(res)
		if err != nil {
			log.Error("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run(s.onResponseScript)
		if err != nil {
			log.Error("Error while executing onRequest callback: %s", err)
			return nil
		}

		if jsres.WasModified() {
			jsres.UpdateHash()
			return jsres
		}
	}

	return nil
}
