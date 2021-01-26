package events_stream

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/modules/net_sniff"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

var (
	reJsonKey = regexp.MustCompile(`("[^"]+"):`)
)

func (mod *EventsStream) shouldDumpHttpRequest(req net_sniff.HTTPRequest) bool {
	if mod.dumpHttpReqs {
		// dump all
		return true
	} else if req.Method != "GET" {
		// dump if it's not just a GET
		return true
	}
	// search for interesting headers and cookies
	for name := range req.Headers {
		headerName := strings.ToLower(name)
		if strings.Contains(headerName, "auth") || strings.Contains(headerName, "token") {
			return true
		}
	}
	return false
}

func (mod *EventsStream) shouldDumpHttpResponse(res net_sniff.HTTPResponse) bool {
	if mod.dumpHttpResp {
		return true
	} else if strings.Contains(res.ContentType, "text/plain") {
		return true
	} else if strings.Contains(res.ContentType, "application/json") {
		return true
	} else if strings.Contains(res.ContentType, "text/xml") {
		return true
	}
	// search for interesting headers
	for name := range res.Headers {
		headerName := strings.ToLower(name)
		if strings.Contains(headerName, "auth") || strings.Contains(headerName, "token") || strings.Contains(headerName, "cookie") {
			return true
		}
	}
	return false
}

func (mod *EventsStream) dumpForm(body []byte) string {
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
			form = append(form, tui.Bold(tui.Red(value)))
		}
	}
	return "\n" + strings.Join(form, "&") + "\n"
}

func (mod *EventsStream) dumpText(body []byte) string {
	return "\n" + tui.Bold(tui.Red(string(body))) + "\n"
}

func (mod *EventsStream) dumpGZIP(body []byte) string {
	buffer := bytes.NewBuffer(body)
	uncompressed := bytes.Buffer{}
	reader, err := gzip.NewReader(buffer)
	if mod.dumpFormatHex {
		if err != nil {
			return mod.dumpRaw(body)
		} else if _, err = uncompressed.ReadFrom(reader); err != nil {
			return mod.dumpRaw(body)
		}
		return mod.dumpRaw(uncompressed.Bytes())
	} else {
		if err != nil {
			return mod.dumpText(body)
		} else if _, err = uncompressed.ReadFrom(reader); err != nil {
			return mod.dumpText(body)
		}
		return mod.dumpText(uncompressed.Bytes())
	}
}

func (mod *EventsStream) dumpJSON(body []byte) string {
	var buf bytes.Buffer
	var pretty string

	if err := json.Indent(&buf, body, "", "  "); err != nil {
		pretty = string(body)
	} else {
		pretty = buf.String()
	}

	return "\n" + reJsonKey.ReplaceAllString(pretty, tui.Green(`$1:`)) + "\n"
}

func (mod *EventsStream) dumpXML(body []byte) string {
	// TODO: indent xml
	return "\n" + string(body) + "\n"
}

func (mod *EventsStream) dumpRaw(body []byte) string {
	return "\n" + hex.Dump(body) + "\n"
}

func (mod *EventsStream) viewHttpRequest(output io.Writer, e session.Event) {
	se := e.Data.(net_sniff.SnifferEvent)
	req := se.Data.(net_sniff.HTTPRequest)

	fmt.Fprintf(output, "[%s] [%s] %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		se.Message)

	if mod.shouldDumpHttpRequest(req) {
		dump := fmt.Sprintf("%s %s %s\n", tui.Bold(req.Method), req.URL, tui.Dim(req.Proto))
		dump += fmt.Sprintf("%s: %s\n", tui.Blue("Host"), tui.Yellow(req.Host))
		for name, values := range req.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", tui.Blue(name), tui.Yellow(value))
			}
		}

		if req.Body != nil {
			if req.IsType("application/x-www-form-urlencoded") {
				dump += mod.dumpForm(req.Body)
			} else if req.IsType("text/plain") {
				dump += mod.dumpText(req.Body)
			} else if req.IsType("text/xml") {
				dump += mod.dumpXML(req.Body)
			} else if req.IsType("gzip") {
				dump += mod.dumpGZIP(req.Body)
			} else if req.IsType("application/json") {
				dump += mod.dumpJSON(req.Body)
			} else {
				if mod.dumpFormatHex {
					dump += mod.dumpRaw(req.Body)
				} else {
					dump += mod.dumpText(req.Body)
				}
			}
		}

		fmt.Fprintf(output, "\n%s\n", dump)
	}
}

func (mod *EventsStream) viewHttpResponse(output io.Writer, e session.Event) {
	se := e.Data.(net_sniff.SnifferEvent)
	res := se.Data.(net_sniff.HTTPResponse)

	fmt.Fprintf(output, "[%s] [%s] %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		se.Message)

	if mod.shouldDumpHttpResponse(res) {
		dump := fmt.Sprintf("%s %s\n", tui.Dim(res.Protocol), res.Status)
		for name, values := range res.Headers {
			for _, value := range values {
				dump += fmt.Sprintf("%s: %s\n", tui.Blue(name), tui.Yellow(value))
			}
		}

		if res.Body != nil {
			// TODO: add more interesting response types
			if res.IsType("text/plain") {
				dump += mod.dumpText(res.Body)
			} else if res.IsType("application/json") {
				dump += mod.dumpJSON(res.Body)
			} else if res.IsType("text/xml") {
				dump += mod.dumpXML(res.Body)
			}
		}

		fmt.Fprintf(output, "\n%s\n", dump)
	}
}

func (mod *EventsStream) viewHttpEvent(output io.Writer, e session.Event) {
	if e.Tag == "net.sniff.http.request" {
		mod.viewHttpRequest(output, e)
	} else if e.Tag == "net.sniff.http.response" {
		mod.viewHttpResponse(output, e)
	}
}
