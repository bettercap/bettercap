package dns_proxy

import (
	"github.com/bettercap/bettercap/v2/session"
)

type DnsProxy struct {
	session.SessionModule
	proxy *DNSProxy
}

func (mod *DnsProxy) Author() string {
	return "Yarwin Kolff <@buffermet>"
}

func (mod *DnsProxy) Configure() error {
	var err error
	var address string
	var dnsPort int
	var doRedirect bool
	var nameserver string
	var netProtocol string
	var proxyPort int
	var scriptPath string

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, dnsPort = mod.IntParam("dns.port"); err != nil {
		return err
	} else if err, address = mod.StringParam("dns.proxy.address"); err != nil {
		return err
	} else if err, nameserver = mod.StringParam("dns.proxy.nameserver"); err != nil {
		return err
	} else if err, netProtocol = mod.StringParam("dns.proxy.networkprotocol"); err != nil {
		return err
	} else if err, proxyPort = mod.IntParam("dns.proxy.port"); err != nil {
		return err
	} else if err, doRedirect = mod.BoolParam("dns.proxy.redirect"); err != nil {
		return err
	} else if err, scriptPath = mod.StringParam("dns.proxy.script"); err != nil {
		return err
	}

	error := mod.proxy.Configure(address, dnsPort, doRedirect, nameserver, netProtocol, proxyPort, scriptPath)

	return error
}

func (mod *DnsProxy) Description() string {
	return "A full featured DNS proxy that can be used to manipulate DNS traffic."
}

func (mod *DnsProxy) Name() string {
	return "dns.proxy"
}

func NewDnsProxy(s *session.Session) *DnsProxy {
	mod := &DnsProxy{
		SessionModule: session.NewSessionModule("dns.proxy", s),
		proxy:         NewDNSProxy(s, "dns.proxy"),
	}

	mod.AddParam(session.NewIntParameter("dns.port",
		"53",
		"DNS port to redirect when the proxy is activated."))

	mod.AddParam(session.NewStringParameter("dns.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the DNS proxy to."))

	mod.AddParam(session.NewStringParameter("dns.proxy.nameserver",
		"1.1.1.1",
		session.IPv4Validator,
		"DNS resolver address."))

	mod.AddParam(session.NewStringParameter("dns.proxy.networkprotocol",
		"udp",
		"^(udp|tcp|tcp-tls)$",
		"Network protocol for the DNS proxy server to use. Accepted values: udp, tcp, tcp-tls"))

	mod.AddParam(session.NewIntParameter("dns.proxy.port",
		"8053",
		"Port to bind the DNS proxy to."))

	mod.AddParam(session.NewBoolParameter("dns.proxy.redirect",
		"true",
		"Enable or disable port redirection with iptables."))

	mod.AddParam(session.NewStringParameter("dns.proxy.script",
		"",
		"",
		"Path of a JS proxy script."))

	mod.AddHandler(session.NewModuleHandler("dns.proxy on", "",
		"Start the DNS proxy.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("dns.proxy off", "",
		"Stop the DNS proxy.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *DnsProxy) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.proxy.Start()
	})
}

func (mod *DnsProxy) Stop() error {
	return mod.SetRunning(false, func() {
		mod.proxy.Stop()
	})
}
