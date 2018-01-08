package session_modules

import (
	"fmt"
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
	}
}

func (j *JSRequest) ReadBody() string {
	return "TODO: read body"
}
