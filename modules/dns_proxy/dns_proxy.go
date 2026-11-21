package dns_proxy

import (
	"crypto/tls"

	"github.com/bettercap/bettercap/v2/session"
	btls "github.com/bettercap/bettercap/v2/tls"

	"github.com/evilsocket/islazy/str"
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
	var whitelist string
	var blacklist string
	var cert *tls.Certificate

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, dnsPort = mod.IntParam("dns.port"); err != nil {
		return err
	} else if err, address = mod.StringParam("dns.proxy.address"); err != nil {
		return err
	} else if err, nameserver = mod.StringParam("dns.proxy.nameserver"); err != nil {
		return err
	} else if err, proxyPort = mod.IntParam("dns.proxy.port"); err != nil {
		return err
	} else if err, netProtocol = mod.StringParam("dns.proxy.protocol"); err != nil {
		return err
	} else if err, doRedirect = mod.BoolParam("dns.proxy.redirect"); err != nil {
		return err
	} else if err, scriptPath = mod.StringParam("dns.proxy.script"); err != nil {
		return err
	} else if err, blacklist = mod.StringParam("dns.proxy.blacklist"); err != nil {
		return err
	} else if err, whitelist = mod.StringParam("dns.proxy.whitelist"); err != nil {
		return err
	}

	mod.proxy.Blacklist = str.Comma(blacklist)
	mod.proxy.Whitelist = str.Comma(whitelist)

	if netProtocol == "tcp-tls" {
		caCfg, err := btls.CAConfigFromModule("dns.proxy", mod.SessionModule)
		if err != nil {
			return err
		}
		caCfg.IsCA = true

		if caCfg.IsEmudhra() {
			ca, err := caCfg.Resolve()
			if err != nil {
				return err
			}
			mod.Warning("serving DNS over TLS with revoked eMudhra CA %s", ca)
		} else if exists, err := caCfg.FilesExist(); err != nil {
			return err
		} else if exists {
			mod.Info("loading proxy certification authority TLS key from %s", caCfg.KeyFile)
			mod.Info("loading proxy certification authority TLS certificate from %s", caCfg.CertFile)
		} else {
			mod.Debug("%+v", caCfg.Config)
			mod.Info("generating proxy certification authority TLS key to %s", caCfg.KeyFile)
			mod.Info("generating proxy certification authority TLS certificate to %s", caCfg.CertFile)
		}

		// DoT clients validate the resolver hostname, which defaults to the
		// common name of the configured certificate (cloudflare-dns.com).
		if cert, err = caCfg.LoadForHost(caCfg.Config.CommonName, caCfg.Config.Bits); err != nil {
			return err
		}
	}

	err = mod.proxy.Configure(address, dnsPort, doRedirect, nameserver, netProtocol,
		proxyPort, scriptPath, cert)

	return err
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

	mod.AddParam(session.NewStringParameter("dns.proxy.blacklist", "", "",
		"Comma separated list of client IPs to skip while proxying (wildcard allowed)."))

	mod.AddParam(session.NewStringParameter("dns.proxy.whitelist", "", "",
		"Comma separated list of client IPs to proxy if the blacklist is used."))

	mod.AddParam(session.NewStringParameter("dns.proxy.nameserver",
		"1.1.1.1",
		session.IPv4Validator,
		"DNS resolver address."))

	mod.AddParam(session.NewIntParameter("dns.proxy.port",
		"8053",
		"Port to bind the DNS proxy to."))

	mod.AddParam(session.NewStringParameter("dns.proxy.protocol",
		"udp",
		"^(udp|tcp|tcp-tls)$",
		"Network protocol for the DNS proxy server to use. Accepted values: udp, tcp, tcp-tls"))

	mod.AddParam(session.NewBoolParameter("dns.proxy.redirect",
		"true",
		"Enable or disable port redirection with iptables."))

	mod.AddParam(session.NewStringParameter("dns.proxy.certificate",
		"~/.bettercap-ca.cert.pem",
		"",
		"DoT certificate for 'auto' mode (generated if missing); ignored in 'emudhra' mode."))

	mod.AddParam(session.NewStringParameter("dns.proxy.key",
		"~/.bettercap-ca.key.pem",
		"",
		"DoT private key for 'auto' mode (generated if missing); ignored in 'emudhra' mode."))

	btls.CertConfigToModule("dns.proxy", &mod.SessionModule, btls.DefaultCloudflareDNSConfig)
	btls.CAConfigToModule("dns.proxy", &mod.SessionModule)

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
