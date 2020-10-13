package http_proxy

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
	Headers     string
	Body        string

	refHash   string
	resp      *http.Response
	bodyRead  bool
	bodyClear bool
}

func NewJSResponse(res *http.Response) *JSResponse {
	cType := ""
	headers := ""
	code := 200

	if res != nil {
		code = res.StatusCode
		for name, values := range res.Header {
			for _, value := range values {
				headers += name + ": " + value + "\r\n"

				if strings.ToLower(name) == "content-type" {
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
		bodyClear:   false,
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
	if j.bodyRead {
		// body was read
		return true
	} else if j.bodyClear {
		// body was cleared manually
		return true
	} else if j.Body != "" {
		// body was not read but just set
		return true
	}
	// check if any of the fields has been changed
	return j.NewHash() != j.refHash
}

func (j *JSResponse) GetHeader(name, deflt string) string {
	headers := strings.Split(j.Headers, "\r\n")
	for i := 0; i < len(headers); i++ {
		if headers[i] != "" {
			header_parts := header_regexp.FindAllSubmatch([]byte(headers[i]), 1)
			if len(header_parts) != 0 && len(header_parts[0]) == 3 {
				header_name := string(header_parts[0][1])
				header_value := string(header_parts[0][2])

				if strings.ToLower(name) == strings.ToLower(header_name) {
					return header_value
				}
			}
		}
	}
	return deflt
}

func (j *JSResponse) SetHeader(name, value string) {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	if strings.ToLower(name) == "content-type" {
		j.ContentType = value
	}

	headers := strings.Split(j.Headers, "\r\n")
	for i := 0; i < len(headers); i++ {
		if headers[i] != "" {
			header_parts := header_regexp.FindAllSubmatch([]byte(headers[i]), 1)
			if len(header_parts) != 0 && len(header_parts[0]) == 3 {
				header_name := string(header_parts[0][1])
				header_value := string(header_parts[0][2])

				if strings.ToLower(name) == strings.ToLower(header_name) {
					old_header := header_name + ": " + header_value + "\r\n"
					new_header := name + ": " + value + "\r\n"
					j.Headers = strings.Replace(j.Headers, old_header, new_header, 1)
					return
				}
			}
		}
	}
	j.Headers += name + ": " + value + "\r\n"
}

func (j *JSResponse) RemoveHeader(name string) {
	headers := strings.Split(j.Headers, "\r\n")
	for i := 0; i < len(headers); i++ {
		if headers[i] != "" {
			header_parts := header_regexp.FindAllSubmatch([]byte(headers[i]), 1)
			if len(header_parts) != 0 && len(header_parts[0]) == 3 {
				header_name := string(header_parts[0][1])
				header_value := string(header_parts[0][2])

				if strings.ToLower(name) == strings.ToLower(header_name) {
					removed_header := header_name + ": " + header_value + "\r\n"
					j.Headers = strings.Replace(j.Headers, removed_header, "", 1)
					return
				}
			}
		}
	}
}

func (j *JSResponse) ClearBody() {
	j.Body = ""
	j.bodyClear = true
}

func (j *JSResponse) ToResponse(req *http.Request) (resp *http.Response) {
	resp = goproxy.NewResponse(req, j.ContentType, j.Status, j.Body)

	headers := strings.Split(j.Headers, "\r\n")
	for i := 0; i < len(headers); i++ {
		if headers[i] != "" {
			header_parts := header_regexp.FindAllSubmatch([]byte(headers[i]), 1)
			if len(header_parts) != 0 && len(header_parts[0]) == 3 {
				header_name := string(header_parts[0][1])
				header_value := string(header_parts[0][2])

				resp.Header.Add(header_name, header_value)
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
	j.bodyClear = false
	// reset the response body to the original unread state
	j.resp.Body = ioutil.NopCloser(bytes.NewBuffer(raw))

	return j.Body
}
