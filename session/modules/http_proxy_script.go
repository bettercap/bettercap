package session_modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/robertkrimen/otto"
)

type ProxyScript struct {
	Path   string
	Source string
	VM     *otto.Otto
	gil    *sync.Mutex
}

func LoadProxyScript(path string) (err error, s *ProxyScript) {
	log.Infof("Loading proxy script %s ...", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	s = &ProxyScript{
		Path:   path,
		Source: string(raw),
		VM:     otto.New(),
		gil:    &sync.Mutex{},
	}

	_, err = s.VM.Run(s.Source)
	if err == nil {
		cb, err := s.VM.Get("onLoad")
		if err == nil && cb.IsFunction() {
			_, err = s.VM.Run("onLoad()")
			if err != nil {
				log.Errorf("Error while executing onLoad callback: %s", err)
				return err, nil
			}
		}
	}

	return
}

func (s ProxyScript) reqToJS(req *http.Request) JSRequest {
	headers := make([]JSHeader, 0)
	for key, values := range req.Header {
		for _, value := range values {
			headers = append(headers, JSHeader{key, value})
		}
	}

	return JSRequest{
		Method:   req.Method,
		Version:  fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor),
		Path:     req.URL.Path,
		Hostname: req.Host,
		Headers:  headers,
	}
}

func (s ProxyScript) resToJS(res *http.Response) *JSResponse {
	cType := ""
	headers := ""

	for name, values := range res.Header {
		for _, value := range values {
			if name == "Content-Type" {
				cType = value
			}
			headers += name + ": " + value + "\r\n"
		}
	}

	return &JSResponse{
		Status:      res.StatusCode,
		ContentType: cType,
		Headers:     headers,
		resp:        res,
	}
}

func (s *ProxyScript) doRequestDefines(req *http.Request) (err error, jsres *JSResponse) {
	jsreq := s.reqToJS(req)
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
	jsreq := s.reqToJS(res.Request)
	if err = s.VM.Set("req", jsreq); err != nil {
		log.Errorf("Error while defining request: %s", err)
		return
	}

	jsres = s.resToJS(res)
	if err = s.VM.Set("res", jsres); err != nil {
		log.Errorf("Error while defining response: %s", err)
		return
	}

	return
}

func (s *ProxyScript) OnRequest(req *http.Request) *JSResponse {
	cb, err := s.VM.Get("onRequest")
	if err == nil && cb.IsFunction() {
		s.gil.Lock()
		defer s.gil.Unlock()

		err, jsres := s.doRequestDefines(req)
		if err != nil {
			log.Errorf("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run("onRequest(req, res)")
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
	cb, err := s.VM.Get("onResponse")
	if err == nil && cb.IsFunction() {
		s.gil.Lock()
		defer s.gil.Unlock()

		err, jsres := s.doResponseDefines(res)
		if err != nil {
			log.Errorf("Error while running bootstrap definitions: %s", err)
			return nil
		}

		_, err = s.VM.Run("onResponse(req, res)")
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
