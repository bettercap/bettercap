package modules

import (
	"github.com/bettercap/bettercap/session"
)

type CustomHttpProxy struct {
	session.SessionModule
	proxy *CustomProxy
}

func NewCustomHttpProxy(s *session.Session) *CustomHttpProxy {
	p := &CustomHttpProxy{
		SessionModule: session.NewSessionModule("custom.http.proxy", s),
		proxy:         NewCustomProxy(s),
	}

	p.AddParam(session.NewStringParameter("custom.http.port",
		"80", session.PortListValidator,
		"HTTP port to redirect when the proxy is activated."))

	p.AddParam(session.NewStringParameter("custom.http.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the Custom HTTP proxy to."))

	p.AddParam(session.NewIntParameter("custom.http.proxy.port",
		"8080",
		"Port to bind the Custom HTTP proxy to."))

	p.AddParam(session.NewBoolParameter("custom.http.proxy.sslstrip",
		"false",
		"Enable or disable SSL stripping."))

	p.AddHandler(session.NewModuleHandler("custom.http.proxy on", "",
		"Start Custom HTTP proxy.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("custom.http.proxy off", "",
		"Stop Custom HTTP proxy.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p *CustomHttpProxy) Name() string {
	return "custom.http.proxy"
}

func (p *CustomHttpProxy) Description() string {
	return "Use a custom HTTP proxy server instead of the builtin one."
}

func (p *CustomHttpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *CustomHttpProxy) Configure() error {
	var err error
	var address string
	var proxyPort int
	var httpPort []string
	var stripSSL bool

	if p.Running() {
		return session.ErrAlreadyStarted
	} else if err, address = p.StringParam("custom.http.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = p.IntParam("custom.http.proxy.port"); err != nil {
		return err
	} else if err, httpPort = p.ListParam("custom.http.port"); err != nil {
		return err
	} else if err, stripSSL = p.BoolParam("custom.http.proxy.sslstrip"); err != nil {
		return err
	}

	return p.proxy.Configure(address, proxyPort, httpPort, stripSSL)
}

func (p *CustomHttpProxy) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		p.proxy.Start()
	})
}

func (p *CustomHttpProxy) Stop() error {
	return p.SetRunning(false, func() {
		p.proxy.Stop()
	})
}

