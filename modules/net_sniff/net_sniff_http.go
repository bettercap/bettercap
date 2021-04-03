package net_sniff

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
)

type HTTPRequest struct {
	Method      string      `json:"method"`
	Proto       string      `json:"proto"`
	Host        string      `json:"host"`
	URL         string      `json:"url:"`
	Headers     http.Header `json:"headers"`
	ContentType string      `json:"content_type"`
	Body        []byte      `json:"body"`
}

func (r HTTPRequest) IsType(ctype string) bool {
	return strings.Contains(r.ContentType, ctype)
}

type HTTPResponse struct {
	Protocol         string      `json:"protocol"`
	Status           string      `json:"status"`
	StatusCode       int         `json:"status_code"`
	Headers          http.Header `json:"headers"`
	Body             []byte      `json:"body"`
	ContentLength    int64       `json:"content_length"`
	ContentType      string      `json:"content_type"`
	TransferEncoding []string    `json:"transfer_encoding"`
}

func (r HTTPResponse) IsType(ctype string) bool {
	return strings.Contains(r.ContentType, ctype)
}

func toSerializableRequest(req *http.Request) HTTPRequest {
	body := []byte(nil)
	ctype := "?"
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
	}

	for name, values := range req.Header {
		if strings.ToLower(name) == "content-type" {
			for _, value := range values {
				ctype = value
			}
		}
	}

	return HTTPRequest{
		Method:      req.Method,
		Proto:       req.Proto,
		Host:        req.Host,
		URL:         req.URL.String(),
		Headers:     req.Header,
		ContentType: ctype,
		Body:        body,
	}
}

func toSerializableResponse(res *http.Response) HTTPResponse {
	body := []byte(nil)
	ctype := "?"
	cenc := ""
	for name, values := range res.Header {
		name = strings.ToLower(name)
		if name == "content-type" {
			for _, value := range values {
				ctype = value
			}
		} else if name == "content-encoding" {
			for _, value := range values {
				cenc = value
			}
		}
	}

	if res.Body != nil {
		body, _ = ioutil.ReadAll(res.Body)
	}

	// attempt decompression, but since this has been parsed by just
	// a tcp packet, it will probably fail
	if body != nil && strings.Contains(cenc, "gzip") {
		buffer := bytes.NewBuffer(body)
		uncompressed := bytes.Buffer{}
		if reader, err := gzip.NewReader(buffer); err == nil {
			if _, err = uncompressed.ReadFrom(reader); err == nil {
				body = uncompressed.Bytes()
			}
		}
	}

	return HTTPResponse{
		Protocol:         res.Proto,
		Status:           res.Status,
		StatusCode:       res.StatusCode,
		Headers:          res.Header,
		Body:             body,
		ContentLength:    res.ContentLength,
		ContentType:      ctype,
		TransferEncoding: res.TransferEncoding,
	}
}

func httpParser(srcIP, dstIP net.IP, payload []byte, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	if req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data))); err == nil {
		if user, pass, ok := req.BasicAuth(); ok {
			NewSnifferEvent(
				pkt.Metadata().Timestamp,
				"http.request",
				srcIP.String(),
				req.Host,
				toSerializableRequest(req),
				"%s %s %s %s%s - %s %s, %s %s",
				tui.Wrap(tui.BACKRED+tui.FOREBLACK, "http"),
				vIP(srcIP),
				tui.Wrap(tui.BACKLIGHTBLUE+tui.FOREBLACK, req.Method),
				tui.Yellow(req.Host),
				vURL(req.URL.String()),
				tui.Bold("USER"),
				tui.Red(user),
				tui.Bold("PASS"),
				tui.Red(pass),
			).Push()
		} else {
			NewSnifferEvent(
				pkt.Metadata().Timestamp,
				"http.request",
				srcIP.String(),
				req.Host,
				toSerializableRequest(req),
				"%s %s %s %s%s",
				tui.Wrap(tui.BACKRED+tui.FOREBLACK, "http"),
				vIP(srcIP),
				tui.Wrap(tui.BACKLIGHTBLUE+tui.FOREBLACK, req.Method),
				tui.Yellow(req.Host),
				vURL(req.URL.String()),
			).Push()
		}

		return true
	} else if res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil); err == nil {
		sres := toSerializableResponse(res)
		NewSnifferEvent(
			pkt.Metadata().Timestamp,
			"http.response",
			srcIP.String(),
			dstIP.String(),
			sres,
			"%s %s:%d %s -> %s (%s %s)",
			tui.Wrap(tui.BACKRED+tui.FOREBLACK, "http"),
			vIP(srcIP),
			tcp.SrcPort,
			tui.Bold(res.Status),
			vIP(dstIP),
			tui.Dim(humanize.Bytes(uint64(len(sres.Body)))),
			tui.Yellow(sres.ContentType),
		).Push()

		return true
	}

	return false
}
