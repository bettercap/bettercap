package modules

import (
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

type AnyProxy struct {
	session.SessionModule
	Redirection *firewall.Redirection
}

func NewAnyProxy(s *session.Session) *AnyProxy {
	p := &AnyProxy{
		SessionModule: session.NewSessionModule("any.proxy", s),
	}

	p.AddParam(session.NewStringParameter("any.proxy.iface",
		session.ParamIfaceName,
		"",
		"Interface to redirect packets from."))

	p.AddParam(session.NewStringParameter("any.proxy.protocol",
		"TCP",
		"(TCP|UDP)",
		"Proxy protocol."))

	p.AddParam(session.NewIntParameter("any.proxy.src_port",
		"80",
		"Remote port to redirect when the module is activated."))

	p.AddParam(session.NewStringParameter("any.proxy.src_address",
		"",
		"",
		"Leave empty to intercept any source address."))

	p.AddParam(session.NewStringParameter("any.proxy.dst_address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address where the proxy is listening."))

	p.AddParam(session.NewIntParameter("any.proxy.dst_port",
		"8080",
		"Port where the proxy is listening."))

	p.AddHandler(session.NewModuleHandler("any.proxy on", "",
		"Start the custom proxy redirection.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("any.proxy off", "",
		"Stop the custom proxy redirection.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p *AnyProxy) Name() string {
	return "any.proxy"
}

func (p *AnyProxy) Description() string {
	return "A firewall redirection to any custom proxy."
}

func (p *AnyProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *AnyProxy) Configure() error {
	var err error
	var srcPort int
	var dstPort int
	var iface string
	var protocol string
	var srcAddress string
	var dstAddress string

	if p.Running() {
		return session.ErrAlreadyStarted
	}
	if err, iface = p.StringParam("any.proxy.iface"); err != nil {
		return err
	}
	if err, protocol = p.StringParam("any.proxy.protocol"); err != nil {
		return err
	}
	if err, srcPort = p.IntParam("any.proxy.src_port"); err != nil {
		return err
	}
	if err, dstPort = p.IntParam("any.proxy.dst_port"); err != nil {
		return err
	}
	if err, srcAddress = p.StringParam("any.proxy.src_address"); err != nil {
		return err
	}
	if err, dstAddress = p.StringParam("any.proxy.dst_address"); err != nil {
		return err
	}

	if !p.Session.Firewall.IsForwardingEnabled() {
		log.Info("Enabling forwarding.")
		p.Session.Firewall.EnableForwarding(true)
	}

	p.Redirection = firewall.NewRedirection(iface,
		protocol,
		srcPort,
		dstAddress,
		dstPort)

	if srcAddress != "" {
		p.Redirection.SrcAddress = srcAddress
	}

	if err := p.Session.Firewall.EnableRedirection(p.Redirection, true); err != nil {
		return err
	}

	log.Info("Applied redirection %s", p.Redirection.String())

	return nil
}

func (p *AnyProxy) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {})
}

func (p *AnyProxy) Stop() error {
	if p.Redirection != nil {
		log.Info("Disabling redirection %s", p.Redirection.String())
		if err := p.Session.Firewall.EnableRedirection(p.Redirection, false); err != nil {
			return err
		}
		p.Redirection = nil
	}

	return p.SetRunning(false, func() {})
}
