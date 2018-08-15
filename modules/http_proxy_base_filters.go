package modules

import (
	"io/ioutil"
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

func (p *HTTPProxy) doScriptInjection(res *http.Response, cType string) (error, *http.Response) {
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, nil
	} else if html := string(raw); strings.Contains(html, "</head>") {
		log.Info("(%s) > injecting javascript (%d bytes) into %s for %s",
			core.Green(p.Name),
			len(p.jsHook),
			core.Yellow(res.Request.Host+res.Request.URL.Path),
			core.Bold(res.Request.RemoteAddr))

		html = strings.Replace(html, "</head>", p.jsHook, -1)
		newResp := goproxy.NewResponse(res.Request, cType, res.StatusCode, html)
		for k, vv := range res.Header {
			for _, v := range vv {
				newResp.Header.Add(k, v)
			}
		}

		return nil, newResp
	}

	return nil, nil
}

func (p *HTTPProxy) onResponseFilter(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// sometimes it happens ¯\_(ツ)_/¯
	if res == nil {
		return nil
	}

	log.Debug("(%s) > %s %s %s%s", core.Green(p.Name), res.Request.RemoteAddr, res.Request.Method, res.Request.Host, res.Request.URL.Path)

	p.stripper.Process(res, ctx)

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
		if err, injectedResponse := p.doScriptInjection(res, cType); err != nil {
			log.Error("(%s) error while injecting javascript: %s", p.Name, err)
		} else if injectedResponse != nil {
			return injectedResponse
		}
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
