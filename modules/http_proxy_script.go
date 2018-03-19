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
	onCommandScript  *otto.Script
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
		onCommandScript:  nil,
	}

	if s.hasCallback("onRequest") {
		s.onRequestScript, err = s.VM.Compile("", "onRequest(req, res)")
		if err != nil {
			log.Error("Error while compiling onRequest callback: %s", err)
			return
		}
	}

	if s.hasCallback("onResponse") {
		s.onResponseScript, err = s.VM.Compile("", "onResponse(req, res)")
		if err != nil {
			log.Error("Error while compiling onResponse callback: %s", err)
			return
		}
	}

	if s.hasCallback("onCommand") {
		s.onCommandScript, err = s.VM.Compile("", "onCommand(cmd)")
		if err != nil {
			log.Error("Error while compiling onCommand callback: %s", err)
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

func (s *HttpProxyScript) doRequestDefines(req *http.Request) (err error, jsreq *JSRequest, jsres *JSResponse) {
	jsreq = NewJSRequest(req)
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

func (s *HttpProxyScript) doResponseDefines(res *http.Response) (err error, jsreq *JSRequest, jsres *JSResponse) {
	jsreq = NewJSRequest(res.Request)
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

func (s *HttpProxyScript) doCommandDefines(cmd string) (err error) {
	if err = s.VM.Set("cmd", cmd); err != nil {
		log.Error("Error while defining cmd: %s", err)
	}
	return
}

func (s *HttpProxyScript) OnRequest(original *http.Request) (jsreq *JSRequest, jsres *JSResponse) {
	var err error

	if s.onRequestScript != nil {
		s.Lock()
		defer s.Unlock()

		if err, jsreq, jsres = s.doRequestDefines(original); err != nil {
			log.Error("Error while running bootstrap definitions: %s", err)
			return nil, nil
		}

		if _, err = s.VM.Run(s.onRequestScript); err != nil {
			log.Error("Error while executing onRequest callback: %s", err)
			return nil, nil
		}

		if jsreq.WasModified() {
			jsreq.UpdateHash()
			return jsreq, nil
		} else if jsres.WasModified() {
			jsres.UpdateHash()
			return nil, jsres
		}
	}

	return nil, nil
}

func (s *HttpProxyScript) OnResponse(res *http.Response) (jsreq *JSRequest, jsres *JSResponse) {
	var err error

	if s.onResponseScript != nil {
		s.Lock()
		defer s.Unlock()

		if err, jsreq, jsres = s.doResponseDefines(res); err != nil {
			log.Error("Error while running bootstrap definitions: %s", err)
			return nil, nil
		}

		if _, err = s.VM.Run(s.onResponseScript); err != nil {
			log.Error("Error while executing onRequest callback: %s", err)
			return nil, nil
		}

		if jsres.WasModified() {
			jsres.UpdateHash()
			return nil, jsres
		}
	}

	return nil, nil
}

func (s *HttpProxyScript) OnCommand(cmd string) bool {
	if s.onCommandScript != nil {
		s.Lock()
		defer s.Unlock()

		if err := s.doCommandDefines(cmd); err != nil {
			log.Error("Error while running bootstrap onCommand definitions: %s", err)
			return false
		}

		if ret, err := s.VM.Run(s.onCommandScript); err != nil {
			log.Error("Error while executing onCommand callback: %s", err)
			return false
		} else if v, err := ret.ToBoolean(); err == nil {
			return v
		}
	}

	return false
}
