package modules

import (
	"net/http"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"

	"github.com/elazarl/goproxy"
)

func (p *HTTPProxy) onRequestFilter(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	log.Debug("(%s) < %s %s %s%s", core.Green(p.Name), req.RemoteAddr, req.Method, req.Host, req.URL.Path)

	redir := p.stripper.Preprocess(req, ctx)
	if redir != nil {
		// we need to redirect the user in order to make
		// some session cookie expire
		return req, redir
	}

	// do we have a proxy script?
	if p.Script == nil {
		return req, nil
	}

	// run the module OnRequest callback if defined
	jsreq, jsres := p.Script.OnRequest(req)
	if jsreq != nil {
		// the request has been changed by the script
		p.logRequestAction(req, jsreq)
		return jsreq.ToRequest(), nil
	} else if jsres != nil {
		// a fake response has been returned by the script
		p.logResponseAction(req, jsres)
		return req, jsres.ToResponse(req)
	}

	return req, nil
}

func (p *HTTPProxy) onResponseFilter(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// sometimes it happens ¯\_(ツ)_/¯
	if res == nil {
		return nil
	}

	req := res.Request
	log.Debug("(%s) > %s %s %s%s", core.Green(p.Name), req.RemoteAddr, req.Method, req.Host, req.URL.Path)

	p.stripper.Process(res, ctx)

	// do we have a proxy script?
	if p.Script == nil {
		return res
	}

	_, jsres := p.Script.OnResponse(res)
	if jsres != nil {
		// the response has been changed by the script
		p.logResponseAction(res.Request, jsres)
		return jsres.ToResponse(res.Request)
	}

	return res
}

func (p *HTTPProxy) logRequestAction(req *http.Request, jsreq *JSRequest) {
	p.sess.Events.Add(p.Name+".spoofed-request", struct {
		To     string
		Method string
		Host   string
		Path   string
		Size   int
	}{
		strings.Split(req.RemoteAddr, ":")[0],
		jsreq.Method,
		jsreq.Hostname,
		jsreq.Path,
		len(jsreq.Body),
	})
}

func (p *HTTPProxy) logResponseAction(req *http.Request, jsres *JSResponse) {
	p.sess.Events.Add(p.Name+".spoofed-response", struct {
		To     string
		Method string
		Host   string
		Path   string
		Size   int
	}{
		strings.Split(req.RemoteAddr, ":")[0],
		req.Method,
		req.Host,
		req.URL.Path,
		len(jsres.Body),
	})
}
