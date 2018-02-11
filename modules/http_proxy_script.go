package modules

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/robertkrimen/otto"
)

type ProxyScript struct {
	sync.Mutex

	Path   string
	Source string
	VM     *otto.Otto

	sess             *session.Session
	onRequestScript  *otto.Script
	onResponseScript *otto.Script
	cbCacheLock      *sync.Mutex
	cbCache          map[string]bool
}

func LoadProxyScriptSource(path, source string, sess *session.Session) (err error, s *ProxyScript) {
	s = &ProxyScript{
		Path:   path,
		Source: source,
		VM:     otto.New(),

		sess:             sess,
		onRequestScript:  nil,
		onResponseScript: nil,
		cbCacheLock:      &sync.Mutex{},
		cbCache:          make(map[string]bool),
	}

	// this will define callbacks and global objects
	_, err = s.VM.Run(s.Source)
	if err != nil {
		return
	}

	// define session pointer
	err = s.VM.Set("env", sess.Env.Storage)
	if err != nil {
		log.Error("Error while defining environment: %s", err)
		return
	}

	err = s.defineBuiltins()
	if err != nil {
		log.Error("Error while defining builtin functions: %s", err)
		return
	}

	// run onLoad if defined
	if s.hasCallback("onLoad") {
		_, err = s.VM.Run("onLoad()")
		if err != nil {
			log.Error("Error while executing onLoad callback: %s", err)
			return
		}
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

func LoadProxyScript(path string, sess *session.Session) (err error, s *ProxyScript) {
	log.Info("Loading proxy script %s ...", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return LoadProxyScriptSource(path, string(raw), sess)
}

func (s *ProxyScript) hasCallback(name string) bool {
	s.cbCacheLock.Lock()
	defer s.cbCacheLock.Unlock()

	// check the cache
	has, found := s.cbCache[name]
	if found == false {
		// check the VM
		cb, err := s.VM.Get(name)
		if err == nil && cb.IsFunction() {
			has = true
		} else {
			has = false
		}
		s.cbCache[name] = has
	}

	return has
}

func (s *ProxyScript) doRequestDefines(req *http.Request) (err error, jsres *JSResponse) {
	// convert request and define empty response to be optionally filled
	jsreq := NewJSRequest(req)
	if err = s.VM.Set("req", &jsreq); err != nil {
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

func (s *ProxyScript) doResponseDefines(res *http.Response) (err error, jsres *JSResponse) {
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

func (s *ProxyScript) OnRequest(req *http.Request) *JSResponse {
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

func (s *ProxyScript) OnResponse(res *http.Response) *JSResponse {
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
