package modules

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/session"
)

var (
	cookieFilter = map[string]bool{
		"__cfduid": true,
		"_ga":      true,
		"_gat":     true,
	}

	reJsonKey = regexp.MustCompile(`("[^"]+"):`)
)

func (s *EventsStream) shouldDumpHttpRequest(req HTTPRequest) bool {
	// dump if it's not just a GET
	if req.Method != "GET" {
		return true
	}
	// search for interesting headers and cookies
	for name, values := range req.Headers {
		headerName := strings.ToLower(name)
		if strings.Contains(headerName, "auth") || strings.Contains(headerName, "token") {
			return true
		} else if headerName == "cookie" {
			for _, value := range values {
				cookies := strings.Split(value, ";")
				for _, cookie := range cookies {
					parts := strings.Split(cookie, "=")
					if _, found := cookieFilter[parts[0]]; found == false {
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *EventsStream) shouldDumpHttpResponse(res HTTPResponse) bool {
	if strings.Contains(res.ContentType, "text/plain") {
		return true
	} else if strings.Contains(res.ContentType, "application/json") {
		return true
	} else if strings.Contains(res.ContentType, "text/xml") {
		return true
	}

	// search for interesting headers
	for name, _ := range res.Headers {
		headerName := strings.ToLower(name)
		if strings.Contains(headerName, "auth") || strings.Contains(headerName, "token") || strings.Contains(headerName, "cookie") {
			return true
		}
	}

	return false
}

func (s *EventsStream) dumpForm(body []byte) string {
	form := []string{}
	for _, v := range strings.Split(string(body), "&") {
		if strings.Contains(v, "=") {
			parts := strings.SplitN(v, "=", 2)
			name := parts[0]
			value, err := url.QueryUnescape(parts[1])
			if err != nil {
				value = parts[1]
			}

			form = append(form, fmt.Sprintf("%s=%s", core.Green(name), core.Bold(core.Red(value))))
		} else {
			value, err := url.QueryUnescape(v)
			if err != nil {
				value = v
			}
			form = append(form, fmt.Sprintf("%s", core.Bold(core.Red(value))))
		}
	}
	return "\n" + strings.Join(form, "&") + "\n"
}

func (s *EventsStream) dumpText(body []byte) string {
	return "\n" + core.Bold(core.Red(string(body))) + "\n"
}

func (s *EventsStream) dumpGZIP(body []byte) string {
	buffer := bytes.NewBuffer(body)
	uncompressed := bytes.Buffer{}
	reader, err := gzip.NewReader(buffer)
	if err != nil {
		return s.dumpRaw(body)
	} else if _, err = uncompressed.ReadFrom(reader); err != nil {
		return s.dumpRaw(body)
	}
	return s.dumpRaw(uncompressed.Bytes())
}

func (s *EventsStream) dumpJSON(body []byte) string {
	var buf bytes.Buffer
	var pretty string

	if err := json.Indent(&buf, body, "", "  "); err != nil {
		pretty = string(body)
	} else {
		pretty = string(buf.Bytes())
	}

	return "\n" + reJsonKey.ReplaceAllString(pretty, core.W(core.GREEN, `$1:`)) + "\n"
}

func (s *EventsStream) dumpXML(body []byte) string {
	// TODO: indent xml
	return "\n" + string(body) + "\n"
}

func (s *EventsStream) dumpRaw(body []byte) string {
	return "\n" + hex.Dump(body) + "\n"
}

func (s *EventsStream) viewHttpRequest(e session.Event) {
	se := e.Data.(SnifferEvent)
	req := se.Data.(HTTPRequest)

	fmt.Fprintf(s.output, "[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Message)

	if s.shouldDumpHttpRequest(req) {
		dump := fmt.Sprintf("%s %s %s\n", core.Bold(req.Method), req.URL, core.Dim(req.Proto))
		dump += fmt.Sprintf("%s: %s\n", core.Blue("Host"), core.Yellow(req.Host))
		for name, values := range req.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", core.Blue(name), core.Yellow(value))
			}
		}

		if req.Body != nil {
			if req.IsType("application/x-www-form-urlencoded") {
				dump += s.dumpForm(req.Body)
			} else if req.IsType("text/plain") {
				dump += s.dumpText(req.Body)
			} else if req.IsType("text/xml") {
				dump += s.dumpXML(req.Body)
			} else if req.IsType("gzip") {
				dump += s.dumpGZIP(req.Body)
			} else if req.IsType("application/json") {
				dump += s.dumpJSON(req.Body)
			} else {
				dump += s.dumpRaw(req.Body)
			}
		}

		fmt.Fprintf(s.output, "\n%s\n", dump)
	}
}

func (s *EventsStream) viewHttpResponse(e session.Event) {
	se := e.Data.(SnifferEvent)
	res := se.Data.(HTTPResponse)

	fmt.Fprintf(s.output, "[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Message)

	if s.shouldDumpHttpResponse(res) {
		dump := fmt.Sprintf("%s %s\n", core.Dim(res.Protocol), res.Status)
		for name, values := range res.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", core.Blue(name), core.Yellow(value))
			}
		}

		if res.Body != nil {
			// TODO: add more interesting response types
			if res.IsType("text/plain") {
				dump += s.dumpText(res.Body)
			} else if res.IsType("application/json") {
				dump += s.dumpJSON(res.Body)
			} else if res.IsType("text/xml") {
				dump += s.dumpXML(res.Body)
			}
		}

		fmt.Fprintf(s.output, "\n%s\n", dump)
	}
}

func (s *EventsStream) viewHttpEvent(e session.Event) {
	if e.Tag == "net.sniff.http.request" {
		s.viewHttpRequest(e)
	} else if e.Tag == "net.sniff.http.response" {
		s.viewHttpResponse(e)
	}
}
