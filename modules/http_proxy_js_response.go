package modules

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
)

type JSResponse struct {
	Status      int
	ContentType string
	Headers     []JSHeader
	Body        string

	refHash  string
	resp     *http.Response
	bodyRead bool
}

func NewJSResponse(res *http.Response) *JSResponse {
	cType := ""
	headers := make([]JSHeader, 0)
	code := 200

	if res != nil {
		code = res.StatusCode
		for name, values := range res.Header {
			for _, value := range values {
				headers = append(headers, JSHeader{name, value})

				if name == "Content-Type" {
					cType = value
				}
			}
		}
	}

	resp := &JSResponse{
		Status:      code,
		ContentType: cType,
		Headers:     headers,
		resp:        res,
		bodyRead:    false,
	}
	resp.UpdateHash()

	return resp
}

func (j *JSResponse) NewHash() string {
	hash := fmt.Sprintf("%d.%s", j.Status, j.ContentType)
	for _, h := range j.Headers {
		hash += fmt.Sprintf(".%s-%s", h.Name, h.Value)
	}
	return hash
}

func (j *JSResponse) UpdateHash() {
	j.refHash = j.NewHash()
}

func (j *JSResponse) WasModified() bool {
	if j.bodyRead == true {
		// body was read
		return true
	} else if j.Body != "" {
		// body was not read but just set
		return true
	}

	// check if any of the fields has been changed
	newHash := j.NewHash()
	if newHash != j.refHash {
		return true
	}
	return false
}

func (j *JSResponse) GetHeader(name, deflt string) string {
	name = strings.ToLower(name)
	for _, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			return h.Value
		}
	}
	return deflt
}

func (j *JSResponse) SetHeader(name, value string) {
	name = strings.ToLower(name)
	for i, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			j.Headers[i].Value = value
			return
		}
	}
	j.Headers = append(j.Headers, JSHeader{name, value})
}

func (j *JSResponse) RemoveHeader(name string) {
	name = strings.ToLower(name)
	for i, h := range j.Headers {
		if name == strings.ToLower(h.Name) {
			j.Headers = append(j.Headers[:i], j.Headers[i+1:]...)
		}
	}
}

func (j *JSResponse) ToResponse(req *http.Request) (resp *http.Response) {
	resp = goproxy.NewResponse(req, j.ContentType, j.Status, j.Body)
	if len(j.Headers) > 0 {
		for _, h := range j.Headers {
			resp.Header.Add(h.Name, h.Value)
		}
	}
	return
}

func (j *JSResponse) ReadBody() string {
	defer j.resp.Body.Close()

	raw, err := ioutil.ReadAll(j.resp.Body)
	if err != nil {
		return ""
	}

	j.Body = string(raw)
	j.bodyRead = true
	// reset the response body to the original unread state
	j.resp.Body = ioutil.NopCloser(bytes.NewBuffer(raw))

	return j.Body
}
