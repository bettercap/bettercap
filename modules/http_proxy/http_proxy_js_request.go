package http_proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/session"
)

type JSRequest struct {
	Client      map[string]string
	Method      string
	Version     string
	Scheme      string
	Path        string
	Query       string
	Hostname    string
	Port        string
	ContentType string
	Headers     string
	Body        string

	req      *http.Request
	refHash  string
	bodyRead bool
}

var header_regexp = regexp.MustCompile(`^\s*(.*?)\s*:\s*(.*)\s*$`)

func NewJSRequest(req *http.Request) *JSRequest {
	headers := ""
	cType := ""

	for name, values := range req.Header {
		for _, value := range values {
			headers += name + ": " + value + "\r\n"

			if strings.ToLower(name) == "content-type" {
				cType = value
			}
		}
	}

	client_ip := strings.Split(req.RemoteAddr, ":")[0]
	client_mac := ""
	client_alias := ""
	if endpoint := session.I.Lan.GetByIp(client_ip); endpoint != nil {
		client_mac = endpoint.HwAddress
		client_alias = endpoint.Alias
	}

	jreq := &JSRequest{
		Client:      map[string]string{"IP": client_ip, "MAC": client_mac, "Alias": client_alias},
		Method:      req.Method,
		Version:     fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor),
		Scheme:      req.URL.Scheme,
		Hostname:    req.URL.Hostname(),
		Port:        req.URL.Port(),
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
	hash := fmt.Sprintf("%s.%s.%s.%s.%s.%s.%s.%s.%s.%s",
		j.Client["IP"],
		j.Method,
		j.Version,
		j.Scheme,
		j.Hostname,
		j.Port,
		j.Path,
		j.Query,
		j.ContentType,
		j.Headers)
	hash += "." + j.Body
	return hash
}

func (j *JSRequest) UpdateHash() {
	j.refHash = j.NewHash()
}

func (j *JSRequest) WasModified() bool {
	// body was read
	if j.bodyRead {
		return true
	}
	// check if any of the fields has been changed
	return j.NewHash() != j.refHash
}

func (j *JSRequest) GetHeader(name, deflt string) string {
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

func (j *JSRequest) SetHeader(name, value string) {
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

func (j *JSRequest) RemoveHeader(name string) {
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

	form := make(map[string]string)
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
	portPart := ""
	if j.Port != "" {
		portPart = fmt.Sprintf(":%s", j.Port)
	}

	url := fmt.Sprintf("%s://%s%s%s?%s", j.Scheme, j.Hostname, portPart, j.Path, j.Query)
	if j.Body == "" {
		req, _ = http.NewRequest(j.Method, url, j.req.Body)
	} else {
		req, _ = http.NewRequest(j.Method, url, strings.NewReader(j.Body))
	}

	headers := strings.Split(j.Headers, "\r\n")
	for i := 0; i < len(headers); i++ {
		if headers[i] != "" {
			header_parts := header_regexp.FindAllSubmatch([]byte(headers[i]), 1)
			if len(header_parts) != 0 && len(header_parts[0]) == 3 {
				header_name := string(header_parts[0][1])
				header_value := string(header_parts[0][2])

				if strings.ToLower(header_name) == "content-type" {
					if header_value != j.ContentType {
						req.Header.Set(header_name, j.ContentType)
						continue
					}
				}
				req.Header.Set(header_name, header_value)
			}
		}
	}

	req.RemoteAddr = j.Client["IP"]

	return
}
