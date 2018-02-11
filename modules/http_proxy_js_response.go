package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"

	"github.com/evilsocket/bettercap-ng/core"
)

type JSResponse struct {
	Status      int
	ContentType string
	Headers     string
	Body        string

	refHash  string
	resp     *http.Response
	bodyRead bool
}

func NewJSResponse(res *http.Response) *JSResponse {
	cType := ""
	headers := ""
	code := 200

	if res != nil {
		code = res.StatusCode
		for name, values := range res.Header {
			for _, value := range values {
				if name == "Content-Type" {
					cType = value
				}
				headers += name + ": " + value + "\r\n"
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
	return fmt.Sprintf("%d.%s.%s", j.Status, j.ContentType, j.Headers)
}

func (j *JSResponse) UpdateHash() {
	j.refHash = j.NewHash()
}

func (j *JSResponse) WasModified() bool {
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

func (j *JSResponse) ToResponse(req *http.Request) (resp *http.Response) {
	resp = goproxy.NewResponse(req, j.ContentType, j.Status, j.Body)
	if j.Headers != "" {
		for _, header := range strings.Split(j.Headers, "\n") {
			header = core.Trim(header)
			if header == "" {
				continue
			}
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				resp.Header.Add(parts[0], parts[1])
			}
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

	return j.Body
}
