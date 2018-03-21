package modules

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

	req      *http.Request
	refHash  string
	bodyRead bool
}

func NewJSRequest(req *http.Request) *JSRequest {
	headers := make([]JSHeader, 0)
	cType := ""

	for name, values := range req.Header {
		for _, value := range values {
			headers = append(headers, JSHeader{name, value})

			if name == "Content-Type" {
				cType = value
			}
		}
	}

	jreq := &JSRequest{
		Client:      strings.Split(req.RemoteAddr, ":")[0],
		Method:      req.Method,
		Version:     fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor),
		Hostname:    req.Host,
		Path:        req.URL.Path,
		Query:       req.URL.RawQuery,
		ContentType: cType,
		Headers:     headers,

		req:      req,
		bodyRead: false,
	}
	jreq.UpdateHash()

	return jreq
}

func (j *JSRequest) NewHash() string {
	hash := fmt.Sprintf("%s.%s.%s.%s.%s.%s.%s", j.Client, j.Method, j.Version, j.Hostname, j.Path, j.Query, j.ContentType)
	for _, h := range j.Headers {
		hash += fmt.Sprintf(".%s-%s", h.Name, h.Value)
	}
	hash += "." + j.Body
	return hash
}

func (j *JSRequest) UpdateHash() {
	j.refHash = j.NewHash()
}

func (j *JSRequest) WasModified() bool {
	// body was read
	if j.bodyRead == true {
		return true
	}
	// check if any of the fields has been changed
	newHash := j.NewHash()
	if newHash != j.refHash {
		return true
	}
	return false
}

func (j *JSRequest) GetHeader(name, deflt string) string {
	name = strings.ToLower(name)
	for _, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			return h.Value
		}
	}
	return deflt
}

func (j *JSRequest) SetHeader(name, value string) {
	name = strings.ToLower(name)
	for i, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			j.Headers[i].Value = value
			return
		}
	}
	j.Headers = append(j.Headers, JSHeader{name, value})
}

func (j *JSRequest) RemoveHeader(name string) {
	name = strings.ToLower(name)
	for i, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			j.Headers = append(j.Headers[:i], j.Headers[i+1:]...)
		}
	}
}

func (j *JSRequest) ReadBody() string {
	raw, err := ioutil.ReadAll(j.req.Body)
	if err != nil {
		return ""
	}

	j.Body = string(raw)
	j.bodyRead = true
	// reset the request body to the original unread state
	j.req.Body = ioutil.NopCloser(bytes.NewBuffer(raw))

	return j.Body
}

func (j *JSRequest) ParseForm() map[string]string {
	if j.Body == "" {
		j.Body = j.ReadBody()
	}

	form := make(map[string]string, 0)
	parts := strings.Split(j.Body, "&")

	for _, part := range parts {
		nv := strings.SplitN(part, "=", 2)
		if len(nv) == 2 {
			unescaped, err := url.QueryUnescape(nv[1])
			if err == nil {
				form[nv[0]] = unescaped
			} else {
				form[nv[0]] = nv[1]
			}
		}
	}

	return form
}

func (j *JSRequest) ToRequest() (req *http.Request) {
	url := fmt.Sprintf("%s://%s:%s%s?%s", j.req.URL.Scheme, j.Hostname, j.req.URL.Port(), j.Path, j.Query)
	if j.Body == "" {
		req, _ = http.NewRequest(j.Method, url, j.req.Body)
	} else {
		req, _ = http.NewRequest(j.Method, url, strings.NewReader(j.Body))
	}

	hadType := false
	for _, h := range j.Headers {
		req.Header.Set(h.Name, h.Value)
		if h.Name == "Content-Type" {
			hadType = true
		}
	}

	if hadType == false && j.ContentType != "" {
		req.Header.Set("Content-Type", j.ContentType)
	}

	return
}
