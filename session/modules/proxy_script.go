package session_modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"

	"github.com/robertkrimen/otto"
)

type ProxyScript struct {
	Path   string
	Source string
	VM     *otto.Otto
	gil    *sync.Mutex
}

type JSHeader struct {
	Name  string
	Value string
}

type JSRequest struct {
	Method   string
	Version  string
	Path     string
	Hostname string
	Headers  []JSHeader
	Body     string
}

type JSResponse struct {
	Status      int
	ContentType string
	Headers     string
	Body        string

	wasUpdated bool
	resp       *http.Response
}

func (j *JSResponse) Updated() {
	j.wasUpdated = true
}

func (j *JSResponse) ToResponse(req *http.Request) (resp *http.Response) {
	resp = goproxy.NewResponse(req, j.ContentType, j.Status, j.Body)
	if j.Headers != "" {
		for _, header := range strings.Split(j.Headers, "\n") {
			header = strings.Trim(header, "\n\r\t ")
			if header == "" {
				continue
			}
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				resp.Header.Add(parts[0], parts[1])
			} else {
				log.Warningf("Unexpected header '%s'", header)
			}
		}
	}
	return
}

func (j *JSResponse) ReadBody() string {
	defer j.resp.Body.Close()

	raw, err := ioutil.ReadAll(j.resp.Body)
	if err != nil {
		log.Errorf("Could not read response body: %s", err)
		return ""
	}

	j.Body = string(raw)

	return j.Body
}

func (jsr JSRequest) ReadBody() string {
	return "TODO: read body"
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
