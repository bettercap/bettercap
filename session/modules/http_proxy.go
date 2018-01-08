package session_modules

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/op/go-logging"

	"github.com/elazarl/goproxy"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/firewall"
	"github.com/evilsocket/bettercap-ng/session"
)

var log = logging.MustGetLogger("mitm")

type HttpProxy struct {
	session.SessionModule

	address     string
	redirection *firewall.Redirection
	server      http.Server
	proxy       *goproxy.ProxyHttpServer
	script      *ProxyScript
}

func (p HttpProxy) logAction(req *http.Request, jsres *JSResponse) {
	fmt.Printf("[%s] %s > '%s %s%s' | Sending %d bytes of spoofed response.\n",
		core.Green("http.proxy"),
		core.Bold(strings.Split(req.RemoteAddr, ":")[0]),
		req.Method,
		req.Host,
		req.URL.Path,
		len(jsres.Body))
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
		"<interface address>",
		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`,
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

func (p HttpProxy) Name() string {
	return "HTTP Proxy"
}

func (p HttpProxy) Description() string {
	return "A full featured HTTP proxy that can be used to inject malicious contents into webpages, all HTTP traffic will be redirected to it."
}

func (p HttpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p HttpProxy) OnSessionStarted(s *session.Session) {
	// refresh the address after session has been created
	s.Env.Set("http.proxy.address", s.Interface.IpAddress)
}

func (p HttpProxy) OnSessionEnded(s *session.Session) {
	if p.Running() {
		p.Stop()
	}
}

func (p *HttpProxy) Start() error {
	var http_port int
	var proxy_port int

	if p.Running() == true {
		return fmt.Errorf("HTTP proxy already started.")
	}

	if err, v := p.Param("http.proxy.address").Get(p.Session); err != nil {
		return err
	} else {
		p.address = v.(string)
	}

	if err, v := p.Param("http.proxy.port").Get(p.Session); err != nil {
		return err
	} else {
		proxy_port = v.(int)
	}

	if err, v := p.Param("http.port").Get(p.Session); err != nil {
		return err
	} else {
		http_port = v.(int)
	}

	if err, v := p.Param("http.proxy.script").Get(p.Session); err != nil {
		return err
	} else {
		scriptPath := v.(string)
		if scriptPath != "" {
			if err, p.script = LoadProxyScript(scriptPath, p.Session); err != nil {
				return err
			} else {
				log.Debugf("Proxy script %s loaded.", scriptPath)
			}
		}
	}

	if p.Session.Firewall.IsForwardingEnabled() == false {
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

	address := fmt.Sprintf("%s:%d", p.address, proxy_port)
	log.Infof("Starting proxy on %s.\n", address)

	p.server = http.Server{Addr: address, Handler: p.proxy}
	go func() {
		p.SetRunning(true)
		if err := p.server.ListenAndServe(); err != nil {
			p.SetRunning(false)
			log.Warning(err)
		}
	}()

	return nil
}

func (p *HttpProxy) Stop() error {
	if p.Running() == true {
		p.SetRunning(false)
		p.server.Shutdown(nil)
		log.Info("HTTP proxy stopped.\n")
		if p.redirection != nil {
			if err := p.Session.Firewall.EnableRedirection(p.redirection, false); err != nil {
				return err
			}
			p.redirection = nil
		}
		return nil
	} else {
		return fmt.Errorf("HTTP proxy stopped.")
	}
}

func (p HttpProxy) doProxy(req *http.Request) bool {
	blacklist := []string{
		"localhost",
		"127.0.0.1",
		p.address,
	}

	if req.Host == "" {
		log.Errorf("Got request with empty host: %v\n", req)
		return false
	}

	for _, blacklisted := range blacklist {
		if strings.HasPrefix(req.Host, blacklisted) {
			log.Errorf("Got request with blacklisted host: %s\n", req.Host)
			return false
		}
	}

	return true
}
