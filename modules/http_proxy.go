package modules

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"

	"github.com/evilsocket/bettercap-ng/firewall"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"
)

type HttpProxy struct {
	session.SessionModule

	address     string
	redirection *firewall.Redirection
	server      http.Server
	proxy       *goproxy.ProxyHttpServer
	script      *ProxyScript
}

func (p *HttpProxy) logAction(req *http.Request, jsres *JSResponse) {
	p.Session.Events.Add("http.proxy.spoofed-response", struct {
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

func NewHttpProxy(s *session.Session) *HttpProxy {
	p := &HttpProxy{
		SessionModule: session.NewSessionModule("http.proxy", s),
		proxy:         nil,
		address:       "",
		redirection:   nil,
		script:        nil,
	}

	p.AddParam(session.NewIntParameter("http.port",
		"80",
		"HTTP port to redirect when the proxy is activated."))

	p.AddParam(session.NewStringParameter("http.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the HTTP proxy to."))

	p.AddParam(session.NewIntParameter("http.proxy.port",
		"8080",
		"Port to bind the HTTP proxy to."))

	p.AddParam(session.NewStringParameter("http.proxy.script",
		"",
		"",
		"Path of a proxy JS script."))

	p.AddHandler(session.NewModuleHandler("http.proxy on", "",
		"Start HTTP proxy.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("http.proxy off", "",
		"Stop HTTP proxy.",
		func(args []string) error {
			return p.Stop()
		}))

	proxy := goproxy.NewProxyHttpServer()
	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if p.doProxy(req) == true {
			req.URL.Scheme = "http"
			req.URL.Host = req.Host
			p.proxy.ServeHTTP(w, req)
		}
	})

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if p.script != nil {
			jsres := p.script.OnRequest(req)
			if jsres != nil {
				p.logAction(req, jsres)
				return req, jsres.ToResponse(req)
			}
		}
		return req, nil
	})

	proxy.OnResponse().DoFunc(func(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if p.script != nil {
			jsres := p.script.OnResponse(res)
			if jsres != nil {
				p.logAction(res.Request, jsres)
				return jsres.ToResponse(res.Request)
			}
		}
		return res
	})

	p.proxy = proxy

	return p
}

func (p *HttpProxy) doProxy(req *http.Request) bool {
	blacklist := []string{
		"localhost",
		"127.0.0.1",
		// p.address,
	}

	if req.Host == "" {
		log.Error("Got request with empty host: %v", req)
		return false
	}

	for _, blacklisted := range blacklist {
		if strings.HasPrefix(req.Host, blacklisted) {
			log.Error("Got request with blacklisted host: %s", req.Host)
			return false
		}
	}

	return true
}

func (p *HttpProxy) Name() string {
	return "http.proxy"
}

func (p *HttpProxy) Description() string {
	return "A full featured HTTP proxy that can be used to inject malicious contents into webpages, all HTTP traffic will be redirected to it."
}

func (p *HttpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *HttpProxy) Configure() error {
	var err error
	var http_port int
	var proxy_port int
	var scriptPath string

	if err, p.address = p.StringParam("http.proxy.address"); err != nil {
		return err
	}

	if err, proxy_port = p.IntParam("http.proxy.port"); err != nil {
		return err
	}

	if err, http_port = p.IntParam("http.port"); err != nil {
		return err
	}

	if err, scriptPath = p.StringParam("http.proxy.script"); err != nil {
		return err
	} else if scriptPath != "" {
		if err, p.script = LoadProxyScript(scriptPath, p.Session); err != nil {
			return err
		} else {
			log.Debug("Proxy script %s loaded.", scriptPath)
		}
	}

	p.server = http.Server{Addr: fmt.Sprintf("%s:%d", p.address, proxy_port), Handler: p.proxy}

	if p.Session.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		p.Session.Firewall.EnableForwarding(true)
	}

	p.redirection = firewall.NewRedirection(p.Session.Interface.Name(),
		"TCP",
		http_port,
		p.address,
		proxy_port)

	if err := p.Session.Firewall.EnableRedirection(p.redirection, true); err != nil {
		return err
	}

	log.Debug("Applied redirection %s", p.redirection.String())

	return nil
}

func (p *HttpProxy) Start() error {
	if p.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := p.Configure(); err != nil {
		return err
	}

	p.SetRunning(true)
	go func() {
		if err := p.server.ListenAndServe(); err != nil {
			p.SetRunning(false)
			log.Warning("%s", err)
		}
	}()

	return nil
}

func (p *HttpProxy) Stop() error {
	if p.Running() == false {
		return session.ErrAlreadyStopped
	}
	p.SetRunning(false)

	if p.redirection != nil {
		log.Debug("Disabling redirection %s", p.redirection.String())
		if err := p.Session.Firewall.EnableRedirection(p.redirection, false); err != nil {
			return err
		}
		p.redirection = nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}
