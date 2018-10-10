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

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

var (
	reJsonKey = regexp.MustCompile(`("[^"]+"):`)
)

func (s *EventsStream) shouldDumpHttpRequest(req HTTPRequest) bool {
	if s.dumpHttpReqs {
		// dump all
		return true
	} else if req.Method != "GET" {
		// dump if it's not just a GET
		return true
	}
	// search for interesting headers and cookies
	for name, _ := range req.Headers {
		headerName := strings.ToLower(name)
		if strings.Contains(headerName, "auth") || strings.Contains(headerName, "token") {
			return true
		}
	}
	return false
}

func (s *EventsStream) shouldDumpHttpResponse(res HTTPResponse) bool {
	if s.dumpHttpResp {
		return true
	} else if strings.Contains(res.ContentType, "text/plain") {
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

			form = append(form, fmt.Sprintf("%s=%s", tui.Green(name), tui.Bold(tui.Red(value))))
		} else {
			value, err := url.QueryUnescape(v)
			if err != nil {
				value = v
			}
			form = append(form, fmt.Sprintf("%s", tui.Bold(tui.Red(value))))
		}
	}
	return "\n" + strings.Join(form, "&") + "\n"
}

func (s *EventsStream) dumpText(body []byte) string {
	return "\n" + tui.Bold(tui.Red(string(body))) + "\n"
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

	return "\n" + reJsonKey.ReplaceAllString(pretty, tui.Green(`$1:`)) + "\n"
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
		tui.Green(e.Tag),
		se.Message)

	if s.shouldDumpHttpRequest(req) {
		dump := fmt.Sprintf("%s %s %s\n", tui.Bold(req.Method), req.URL, tui.Dim(req.Proto))
		dump += fmt.Sprintf("%s: %s\n", tui.Blue("Host"), tui.Yellow(req.Host))
		for name, values := range req.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", tui.Blue(name), tui.Yellow(value))
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
		tui.Green(e.Tag),
		se.Message)

	if s.shouldDumpHttpResponse(res) {
		dump := fmt.Sprintf("%s %s\n", tui.Dim(res.Protocol), res.Status)
		for name, values := range res.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", tui.Blue(name), tui.Yellow(value))
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
