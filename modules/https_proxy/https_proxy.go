package https_proxy

import (
	"github.com/bettercap/bettercap/v2/modules/http_proxy"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/bettercap/bettercap/v2/tls"
	"github.com/bettercap/bettercap/v2/tls/emudhra"

	"github.com/evilsocket/islazy/str"
)

type HttpsProxy struct {
	session.SessionModule
	proxy *http_proxy.HTTPProxy
}

func NewHttpsProxy(s *session.Session) *HttpsProxy {
	mod := &HttpsProxy{
		SessionModule: session.NewSessionModule("https.proxy", s),
		proxy:         http_proxy.NewHTTPProxy(s, "https.proxy"),
	}

	mod.AddParam(session.NewIntParameter("https.port",
		"443",
		"HTTPS port to redirect when the proxy is activated."))

	mod.AddParam(session.NewStringParameter("https.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the HTTPS proxy to."))

	mod.AddParam(session.NewIntParameter("https.proxy.port",
		"8083",
		"Port to bind the HTTPS proxy to."))

	mod.AddParam(session.NewBoolParameter("https.proxy.redirect",
		"true",
		"Enable or disable port redirection with iptables."))

	mod.AddParam(session.NewBoolParameter("https.proxy.sslstrip",
		"false",
		"Enable or disable SSL stripping."))

	mod.AddParam(session.NewStringParameter("https.proxy.injectjs",
		"",
		"",
		"URL, path or javascript code to inject into every HTML page."))

	mod.AddParam(session.NewStringParameter("https.proxy.certificate",
		"~/.bettercap-ca.cert.pem",
		"",
		"Proxy CA certificate for 'auto' mode, installed on the client; ignored in 'emudhra' mode."))

	mod.AddParam(session.NewStringParameter("https.proxy.key",
		"~/.bettercap-ca.key.pem",
		"",
		"Proxy CA private key for 'auto' mode; ignored in 'emudhra' mode."))

	mod.AddParam(session.NewIntParameter("https.proxy.leaf.bits",
		"2048",
		"Number of bits of the RSA private key of the spoofed per-host certificates. Values below 2048 are rejected by clients enforcing OpenSSL security level >= 2."))

	tls.CertConfigToModule("https.proxy", &mod.SessionModule, tls.DefaultSpoofConfig)
	tls.CAConfigToModule("https.proxy", &mod.SessionModule)

	mod.AddParam(session.NewStringParameter("https.proxy.script",
		"",
		"",
		"Path of a proxy JS script."))

	mod.AddParam(session.NewStringParameter("https.proxy.blacklist", "", "",
		"Comma separated list of hostnames to skip while proxying (wildcard expressions can be used)."))

	mod.AddParam(session.NewStringParameter("https.proxy.whitelist", "", "",
		"Comma separated list of hostnames to proxy if the blacklist is used (wildcard expressions can be used)."))

	mod.AddHandler(session.NewModuleHandler("https.proxy on", "",
		"Start HTTPS proxy.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("https.proxy.ca", "",
		"List the compromised eMudhra intermediate CAs available in 'emudhra' mode.",
		func(args []string) error {
			if err := emudhra.Err(); err != nil {
				return err
			}
			for _, ca := range emudhra.List() {
				mod.Info("%-20s chain=%-6s key=%-3d sha256=%s",
					ca.Name, ca.Chain, ca.KeyValue(), ca.SHA256)
			}
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("https.proxy off", "",
		"Stop HTTPS proxy.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.InitState("stripper")

	return mod
}

func (mod *HttpsProxy) Name() string {
	return "https.proxy"
}

func (mod *HttpsProxy) Description() string {
	return "A full featured HTTPS proxy that can be used to inject malicious contents into webpages, all HTTPS traffic will be redirected to it."
}

func (mod *HttpsProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *HttpsProxy) Configure() error {
	var err error
	var address string
	var proxyPort int
	var httpPort int
	var doRedirect bool
	var scriptPath string
	var stripSSL bool
	var jsToInject string
	var whitelist string
	var blacklist string
	var leafBits int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, address = mod.StringParam("https.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = mod.IntParam("https.proxy.port"); err != nil {
		return err
	} else if err, httpPort = mod.IntParam("https.port"); err != nil {
		return err
	} else if err, doRedirect = mod.BoolParam("https.proxy.redirect"); err != nil {
		return err
	} else if err, stripSSL = mod.BoolParam("https.proxy.sslstrip"); err != nil {
		return err
	} else if err, scriptPath = mod.StringParam("https.proxy.script"); err != nil {
		return err
	} else if err, jsToInject = mod.StringParam("https.proxy.injectjs"); err != nil {
		return err
	} else if err, blacklist = mod.StringParam("https.proxy.blacklist"); err != nil {
		return err
	} else if err, whitelist = mod.StringParam("https.proxy.whitelist"); err != nil {
		return err
	} else if err, leafBits = mod.IntParam("https.proxy.leaf.bits"); err != nil {
		return err
	}

	mod.proxy.Blacklist = str.Comma(blacklist)
	mod.proxy.Whitelist = str.Comma(whitelist)

	caCfg, err := tls.CAConfigFromModule("https.proxy", mod.SessionModule)
	if err != nil {
		return err
	}
	caCfg.IsCA = true

	if caCfg.IsEmudhra() {
		ca, err := caCfg.Resolve()
		if err != nil {
			return err
		}
		if _, err := ca.Recover(); err != nil {
			return err
		}
		mod.Warning("using revoked eMudhra CA %q (chain %s, recovered private key %d)", ca.Name, ca.Chain, ca.KeyValue())
		mod.Warning("only clients with stale trust stores and no hard-fail CRL/OCSP or CT checks can be intercepted")
		if ca.Chain == emudhra.ChainEmSign {
			mod.Info("if the client only trusts AAA Certificate Services, try 'set https.proxy.ca.chain aaa'")
		}
		if err, m := mod.Session.Module("http.proxy"); err == nil && !m.Running() {
			mod.Info("start 'http.proxy' too so CRL/OCSP lookups are answered locally")
		}
	} else if caCfg.FilesExist() {
		mod.Info("loading proxy certification authority TLS key from %s", caCfg.KeyFile)
		mod.Info("loading proxy certification authority TLS certificate from %s", caCfg.CertFile)
	} else {
		mod.Debug("%+v", caCfg.Config)
		mod.Info("generating proxy certification authority TLS key to %s", caCfg.KeyFile)
		mod.Info("generating proxy certification authority TLS certificate to %s", caCfg.CertFile)
	}

	ca, err := caCfg.Load()
	if err != nil {
		return err
	}

	error := mod.proxy.ConfigureTLS(address, proxyPort, httpPort, doRedirect, scriptPath, ca, jsToInject,
		stripSSL, leafBits)

	// save stripper to share it with other http(s) proxies
	mod.State.Store("stripper", mod.proxy.Stripper)

	return error
}

func (mod *HttpsProxy) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.proxy.Start()
	})
}

func (mod *HttpsProxy) Stop() error {
	mod.State.Store("stripper", nil)
	return mod.SetRunning(false, func() {
		mod.proxy.Stop()
	})
}
