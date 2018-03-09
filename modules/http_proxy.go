package modules

import (
	"github.com/bettercap/bettercap/session"
)

type HttpProxy struct {
	session.SessionModule
	proxy *HTTPProxy
}

func NewHttpProxy(s *session.Session) *HttpProxy {
	p := &HttpProxy{
		SessionModule: session.NewSessionModule("http.proxy", s),
		proxy:         NewHTTPProxy(s),
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

	p.AddParam(session.NewBoolParameter("http.proxy.sslstrip",
		"false",
		"Enable or disable SSL stripping."))

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

	return p
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
	var address string
	var proxyPort int
	var httpPort int
	var scriptPath string
	var stripSSL bool

	if p.Running() == true {
		return session.ErrAlreadyStarted
	} else if err, address = p.StringParam("http.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = p.IntParam("http.proxy.port"); err != nil {
		return err
	} else if err, httpPort = p.IntParam("http.port"); err != nil {
		return err
	} else if err, scriptPath = p.StringParam("http.proxy.script"); err != nil {
		return err
	} else if err, stripSSL = p.BoolParam("http.proxy.sslstrip"); err != nil {
		return err
	}

	return p.proxy.Configure(address, proxyPort, httpPort, scriptPath, stripSSL)
}

func (p *HttpProxy) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		p.proxy.Start()
	})
}

func (p *HttpProxy) Stop() error {
	return p.SetRunning(false, func() {
		p.proxy.Stop()
	})
}
