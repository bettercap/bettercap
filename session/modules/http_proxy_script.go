package session_modules

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/robertkrimen/otto"
)

type ProxyScript struct {
	Path   string
	Source string
	VM     *otto.Otto

	gil              *sync.Mutex
	onRequestScript  *otto.Script
	onResponseScript *otto.Script
	cbCacheLock      *sync.Mutex
	cbCache          map[string]bool
}

func LoadProxyScript(path string) (err error, s *ProxyScript) {
	log.Infof("Loading proxy script %s ...", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	s = &ProxyScript{
		Path:             path,
		Source:           string(raw),
		VM:               otto.New(),
		gil:              &sync.Mutex{},
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

	// run onLoad if defined
	if s.hasCallback("onLoad") {
		_, err = s.VM.Run("onLoad()")
		if err != nil {
			log.Errorf("Error while executing onLoad callback: %s", err)
			return
		}
	}

	// compile call to onRequest if defined
	if s.hasCallback("onRequest") {
		s.onRequestScript, err = s.VM.Compile("", "onRequest(req, res)")
		if err != nil {
			log.Errorf("Error while compiling onRequest callback: %s", err)
			return
		}
	}

	// compile call to onResponse if defined
	if s.hasCallback("onResponse") {
		s.onResponseScript, err = s.VM.Compile("", "onResponse(req, res)")
		if err != nil {
			log.Errorf("Error while compiling onResponse callback: %s", err)
			return
		}
	}

	return
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
	if err = s.VM.Set("req", jsreq); err != nil {
		log.Errorf("Error while defining request: %s", err)
		return
	}

	jsres = &JSResponse{}
	if err = s.VM.Set("res", jsres); err != nil {
		log.Errorf("Error while defining response: %s", err)
		return
	}

	return
}

func (s *ProxyScript) doResponseDefines(res *http.Response) (err error, jsres *JSResponse) {
	// convert both request and response
	jsreq := NewJSRequest(res.Request)
	if err = s.VM.Set("req", jsreq); err != nil {
		log.Errorf("Error while defining request: %s", err)
		return
	}

	jsres = NewJSResponse(res)
	if err = s.VM.Set("res", jsres); err != nil {
		log.Errorf("Error while defining response: %s", err)
		return
	}

	return
}

func (s *ProxyScript) OnRequest(req *http.Request) *JSResponse {
	if s.onRequestScript != nil {
		s.gil.Lock()
		defer s.gil.Unlock()

		err, jsres := s.doRequestDefines(req)
		if err != nil {
			log.Errorf("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run(s.onRequestScript)
		if err != nil {
			log.Errorf("Error while executing onRequest callback: %s", err)
			return nil
		}

		if jsres.wasUpdated == true {
			return jsres
		}
	}

	return nil
}

func (s *ProxyScript) OnResponse(res *http.Response) *JSResponse {
	if s.onResponseScript != nil {
		s.gil.Lock()
		defer s.gil.Unlock()

		err, jsres := s.doResponseDefines(res)
		if err != nil {
			log.Errorf("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run(s.onResponseScript)
		if err != nil {
			log.Errorf("Error while executing onRequest callback: %s", err)
			return nil
		}

		if jsres.wasUpdated == true {
			return jsres
		}
	}

	return nil
}
