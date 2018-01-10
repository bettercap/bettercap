package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type JSHeader struct {
	Name  string
	Value string
}

type JSRequest struct {
	Client      string
	Method      string
	Version     string
	Path        string
	Query       string
	Hostname    string
	ContentType string
	Headers     []JSHeader
	Body        string
	req         *http.Request
}

func NewJSRequest(req *http.Request) JSRequest {
	headers := make([]JSHeader, 0)
	cType := ""

	for key, values := range req.Header {
		for _, value := range values {
			headers = append(headers, JSHeader{key, value})

			if key == "Content-Type" {
				cType = value
			}
		}
	}

	return JSRequest{
		Client:      strings.Split(req.RemoteAddr, ":")[0],
		Method:      req.Method,
		Version:     fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor),
		Hostname:    req.Host,
		Path:        req.URL.Path,
		Query:       req.URL.RawQuery,
		ContentType: cType,
		Headers:     headers,

		req: req,
	}
}

func (j *JSRequest) ReadBody() string {
	raw, err := ioutil.ReadAll(j.req.Body)
	if err != nil {
		return ""
	}

	j.Body = string(raw)

	return j.Body
}
