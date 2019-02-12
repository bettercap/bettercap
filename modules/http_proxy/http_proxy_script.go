package http_proxy

import (
	"net/http"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/robertkrimen/otto"

	"github.com/evilsocket/islazy/plugin"
)

type HttpProxyScript struct {
	*plugin.Plugin

	doOnRequest  bool
	doOnResponse bool
	doOnCommand  bool
}

func LoadHttpProxyScript(path string, sess *session.Session) (err error, s *HttpProxyScript) {
	log.Debug("loading proxy script %s ...", path)

	plug, err := plugin.Load(path)
	if err != nil {
		return
	}

	// define session pointer
	if err = plug.Set("env", sess.Env.Data); err != nil {
		log.Error("Error while defining environment: %+v", err)
		return
	}

	// run onLoad if defined
	if plug.HasFunc("onLoad") {
		if _, err = plug.Call("onLoad"); err != nil {
			log.Error("Error while executing onLoad callback: %s", "\nTraceback:\n  "+err.(*otto.Error).String())
			return
		}
	}

	s = &HttpProxyScript{
		Plugin:       plug,
		doOnRequest:  plug.HasFunc("onRequest"),
		doOnResponse: plug.HasFunc("onResponse"),
		doOnCommand:  plug.HasFunc("onCommand"),
	}
	return
}

func (s *HttpProxyScript) OnRequest(original *http.Request) (jsreq *JSRequest, jsres *JSResponse) {
	if s.doOnRequest {
		jsreq := NewJSRequest(original)
		jsres := NewJSResponse(nil)

		if _, err := s.Call("onRequest", jsreq, jsres); err != nil {
			log.Error("%s", err)
			return nil, nil
		} else if jsreq.WasModified() {
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
	if s.doOnResponse {
		jsreq := NewJSRequest(res.Request)
		jsres := NewJSResponse(res)

		if _, err := s.Call("onResponse", jsreq, jsres); err != nil {
			log.Error("%s", err)
			return nil, nil
		} else if jsres.WasModified() {
			jsres.UpdateHash()
			return nil, jsres
		}
	}

	return nil, nil
}

func (s *HttpProxyScript) OnCommand(cmd string) bool {
	if s.doOnCommand {
		if ret, err := s.Call("onCommand", cmd); err != nil {
			log.Error("Error while executing onCommand callback: %+v", err)
			return false
		} else if v, ok := ret.(bool); ok {
			return v
		}
	}
	return false
}
