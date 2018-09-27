package modules

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/bettercap/bettercap/core"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type HTTPRequest struct {
	Method  string      `json:"method"`
	Host    string      `json:"host"`
	URL     string      `json:"url:"`
	Headers http.Header `json:"headers"`
	Form    url.Values  `json:"form"`
	Body    []byte      `json:"body"`
}

func toSerializableRequest(req *http.Request) HTTPRequest {
	body := []byte(nil)
	form := (url.Values)(nil)

	if err := req.ParseForm(); err == nil {
		form = req.Form
	} else if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
	}

	return HTTPRequest{
		Method:  req.Method,
		Host:    req.Host,
		URL:     req.URL.String(),
		Headers: req.Header,
		Form:    form,
		Body:    body,
	}
}

func httpParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	reader := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(reader)

	if err == nil {
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"http",
			ip.SrcIP.String(),
			req.Host,
			toSerializableRequest(req),
			"%s %s %s %s%s %s",
			core.W(core.BG_RED+core.FG_BLACK, "http"),
			vIP(ip.SrcIP),
			core.W(core.BG_LBLUE+core.FG_BLACK, req.Method),
			core.Yellow(req.Host),
			vURL(req.URL.String()),
			core.Dim(req.UserAgent()),
		).Push()

		return true
	}

	return false
}
