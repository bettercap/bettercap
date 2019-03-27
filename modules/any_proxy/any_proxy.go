package any_proxy

import (
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/session"
)

type AnyProxy struct {
	session.SessionModule
	Redirection *firewall.Redirection
}

func NewAnyProxy(s *session.Session) *AnyProxy {
	mod := &AnyProxy{
		SessionModule: session.NewSessionModule("any.proxy", s),
	}

	mod.AddParam(session.NewStringParameter("any.proxy.iface",
		session.ParamIfaceName,
		"",
		"Interface to redirect packets from."))

	mod.AddParam(session.NewStringParameter("any.proxy.protocol",
		"TCP",
		"(TCP|UDP)",
		"Proxy protocol."))

	mod.AddParam(session.NewIntParameter("any.proxy.src_port",
		"80",
		"Remote port to redirect when the module is activated."))

	mod.AddParam(session.NewStringParameter("any.proxy.src_address",
		"",
		"",
		"Leave empty to intercept any source address."))

	mod.AddParam(session.NewStringParameter("any.proxy.dst_address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address where the proxy is listening."))

	mod.AddParam(session.NewIntParameter("any.proxy.dst_port",
		"8080",
		"Port where the proxy is listening."))

	mod.AddHandler(session.NewModuleHandler("any.proxy on", "",
		"Start the custom proxy redirection.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("any.proxy off", "",
		"Stop the custom proxy redirection.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *AnyProxy) Name() string {
	return "any.proxy"
}

func (mod *AnyProxy) Description() string {
	return "A firewall redirection to any custom proxy."
}

func (mod *AnyProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *AnyProxy) Configure() error {
	var err error
	var srcPort int
	var dstPort int
	var iface string
	var protocol string
	var srcAddress string
	var dstAddress string

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, iface = mod.StringParam("any.proxy.iface"); err != nil {
		return err
	} else if err, protocol = mod.StringParam("any.proxy.protocol"); err != nil {
		return err
	} else if err, srcPort = mod.IntParam("any.proxy.src_port"); err != nil {
		return err
	} else if err, dstPort = mod.IntParam("any.proxy.dst_port"); err != nil {
		return err
	} else if err, srcAddress = mod.StringParam("any.proxy.src_address"); err != nil {
		return err
	} else if err, dstAddress = mod.StringParam("any.proxy.dst_address"); err != nil {
		return err
	}

	if !mod.Session.Firewall.IsForwardingEnabled() {
		mod.Info("Enabling forwarding.")
		mod.Session.Firewall.EnableForwarding(true)
	}

	mod.Redirection = firewall.NewRedirection(iface,
		protocol,
		srcPort,
		dstAddress,
		dstPort)

	if srcAddress != "" {
		mod.Redirection.SrcAddress = srcAddress
	}

	if err := mod.Session.Firewall.EnableRedirection(mod.Redirection, true); err != nil {
		return err
	}

	mod.Info("Applied redirection %s", mod.Redirection.String())

	return nil
}

func (mod *AnyProxy) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {})
}

func (mod *AnyProxy) Stop() error {
	if mod.Redirection != nil {
		mod.Info("Disabling redirection %s", mod.Redirection.String())
		if err := mod.Session.Firewall.EnableRedirection(mod.Redirection, false); err != nil {
			return err
		}
		mod.Redirection = nil
	}

	return mod.SetRunning(false, func() {})
}
