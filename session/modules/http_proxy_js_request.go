package session_modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	req      *http.Request
}

func NewJSRequest(req *http.Request) JSRequest {
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
		req:      req,
	}
}

func (j *JSRequest) ReadBody() string {
	raw, err := ioutil.ReadAll(j.req.Body)
	if err != nil {
		log.Errorf("Could not read request body: %s", err)
		return ""
	}

	j.Body = string(raw)

	return j.Body
}
