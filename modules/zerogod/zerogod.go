package zerogod

import (
	"github.com/bettercap/bettercap/v2/session"
	"github.com/bettercap/bettercap/v2/tls"
)

type ZeroGod struct {
	session.SessionModule
	browser    *Browser
	advertiser *Advertiser
}

func NewZeroGod(s *session.Session) *ZeroGod {
	mod := &ZeroGod{
		SessionModule: session.NewSessionModule("zerogod", s),
		browser:       nil,
		advertiser:    nil,
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddHandler(session.NewModuleHandler("zerogod.discovery on", "",
		"Start DNS-SD / mDNS discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.discovery off", "",
		"Stop DNS-SD / mDNS discovery.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.show", "",
		"Show discovered services.",
		func(args []string) error {
			return mod.show("", false)
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.show-full", "",
		"Show discovered services and their DNS records.",
		func(args []string) error {
			return mod.show("", true)
		}))

	// TODO: add autocomplete
	mod.AddHandler(session.NewModuleHandler("zerogod.show ADDRESS", "zerogod.show (.+)",
		"Show discovered services given an ip address.",
		func(args []string) error {
			return mod.show(args[0], false)
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.show-full ADDRESS", "zerogod.show-full (.+)",
		"Show discovered services and DNS records given an ip address.",
		func(args []string) error {
			return mod.show(args[0], true)
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.save ADDRESS FILENAME", "zerogod.save (.+) (.+)",
		"Save the mDNS information of a given ADDRESS in the FILENAME yaml file.",
		func(args []string) error {
			return mod.save(args[0], args[1])
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.advertise FILENAME", "zerogod.advertise (.+)",
		"Start advertising the mDNS services from the FILENAME yaml file.",
		func(args []string) error {
			if args[0] == "off" {
				return mod.stopAdvertiser()
			}
			return mod.startAdvertiser(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("zerogod.advertise off", "",
		"Start a previously started advertiser.",
		func(args []string) error {
			return mod.stopAdvertiser()
		}))

	mod.AddParam(session.NewStringParameter("zerogod.advertise.certificate",
		"~/.bettercap-zerogod.cert.pem",
		"",
		"TLS certificate file (will be auto generated if filled but not existing) to use for advertised TCP services."))

	mod.AddParam(session.NewStringParameter("zerogod.advertise.key",
		"~/.bettercap-zerogod.key.pem",
		"",
		"TLS key file (will be auto generated if filled but not existing) to use for advertised TCP services."))

	tls.CertConfigToModule("zerogod.advertise", &mod.SessionModule, tls.DefaultLegitConfig)

	mod.AddParam(session.NewStringParameter("zerogod.ipp.save_path",
		"~/.bettercap/zerogod/documents/",
		"",
		"If an IPP acceptor is started, this setting defines where to save documents received for printing."))

	return mod
}

func (mod *ZeroGod) Name() string {
	return "zerogod"
}

func (mod *ZeroGod) Description() string {
	return "A DNS-SD / mDNS / Bonjour / Zeroconf module for discovery and spoofing."
}

func (mod *ZeroGod) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *ZeroGod) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if mod.browser != nil {
		mod.browser.Stop(false)
	}

	mod.browser = NewBrowser()

	return
}

func (mod *ZeroGod) Start() (err error) {
	if err = mod.Configure(); err != nil {
		return err
	}

	// start the root discovery
	if err = mod.startResolver(DNSSD_DISCOVERY_SERVICE); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("service discovery started")
		mod.browser.Wait()
		mod.Info("service discovery stopped")
	})
}

func (mod *ZeroGod) Stop() error {
	return mod.SetRunning(false, func() {
		mod.stopAdvertiser()
		if mod.browser != nil {
			mod.Debug("stopping discovery")

			mod.browser.Stop(true)

			mod.Debug("stopped")

			mod.browser = nil
		}
	})
}
