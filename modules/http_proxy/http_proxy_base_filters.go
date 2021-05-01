package http_proxy

import (
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"

	"github.com/elazarl/goproxy"

	"github.com/evilsocket/islazy/tui"
)

func (p *HTTPProxy) fixRequestHeaders(req *http.Request) {
	req.Header.Del("Accept-Encoding")
	req.Header.Del("If-None-Match")
	req.Header.Del("If-Modified-Since")
	req.Header.Del("Upgrade-Insecure-Requests")
	req.Header.Set("Pragma", "no-cache")
}

func (p *HTTPProxy) onRequestFilter(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if p.shouldProxy(req) {
		p.Debug("< %s %s %s%s", req.RemoteAddr, req.Method, req.Host, req.URL.Path)

		p.fixRequestHeaders(req)

		redir := p.Stripper.Preprocess(req, ctx)
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
	}

	return req, nil
}

func (p *HTTPProxy) getHeader(res *http.Response, header string) string {
	header = strings.ToLower(header)
	for name, values := range res.Header {
		for _, value := range values {
			if strings.ToLower(name) == header {
				return value
			}
		}
	}
	return ""
}

func (p *HTTPProxy) isScriptInjectable(res *http.Response) (bool, string) {
	if p.jsHook == "" {
		return false, ""
	} else if contentType := p.getHeader(res, "Content-Type"); strings.Contains(contentType, "text/html") {
		return true, contentType
	}
	return false, ""
}

func (p *HTTPProxy) doScriptInjection(res *http.Response, cType string) (error) {
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	} else if html := string(raw); strings.Contains(html, "</head>") {
		p.Info("> injecting javascript (%d bytes) into %s (%d bytes) for %s",
			len(p.jsHook),
			tui.Yellow(res.Request.Host+res.Request.URL.Path),
			len(raw),
			tui.Bold(strings.Split(res.Request.RemoteAddr, ":")[0]))

		html = strings.Replace(html, "</head>", p.jsHook, -1)
		res.Header.Set("Content-Length", strconv.Itoa(len(html)))

		// reset the response body to the original unread state
		res.Body = ioutil.NopCloser(strings.NewReader(html))

		return nil
	}

	return nil
}

func (p *HTTPProxy) onResponseFilter(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// sometimes it happens ¯\_(ツ)_/¯
	if res == nil {
		return nil
	}

	if p.shouldProxy(res.Request) {
		p.Debug("> %s %s %s%s", res.Request.RemoteAddr, res.Request.Method, res.Request.Host, res.Request.URL.Path)

		p.Stripper.Process(res, ctx)

		// do we have a proxy script?
		if p.Script != nil {
			_, jsres := p.Script.OnResponse(res)
			if jsres != nil {
				// the response has been changed by the script
				p.logResponseAction(res.Request, jsres)
				return jsres.ToResponse(res.Request)
			}
		}

		// inject javascript code if specified and needed
		if doInject, cType := p.isScriptInjectable(res); doInject {
			if err := p.doScriptInjection(res, cType); err != nil {
				p.Error("error while injecting javascript: %s", err)
			}
		}
	}

	return res
}

func (p *HTTPProxy) logRequestAction(req *http.Request, jsreq *JSRequest) {
	p.Sess.Events.Add(p.Name+".spoofed-request", struct {
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
	p.Sess.Events.Add(p.Name+".spoofed-response", struct {
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
