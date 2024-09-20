package zerogod

import (
	"context"
	"strings"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/bettercap/bettercap/v2/tls"
	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/tui"
)

type ZeroGod struct {
	session.SessionModule

	advertiser  *Advertiser
	rootContext context.Context
	rootCancel  context.CancelFunc
	resolvers   map[string]*zeroconf.Resolver
	mapping     map[string]map[string]*zeroconf.ServiceEntry
}

func NewZeroGod(s *session.Session) *ZeroGod {
	mod := &ZeroGod{
		SessionModule: session.NewSessionModule("zerogod", s),
		mapping:       make(map[string]map[string]*zeroconf.ServiceEntry),
		resolvers:     make(map[string]*zeroconf.Resolver),
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

	if mod.rootContext != nil {
		mod.rootCancel()
	}

	mod.mapping = make(map[string]map[string]*zeroconf.ServiceEntry)
	mod.resolvers = make(map[string]*zeroconf.Resolver)
	mod.rootContext, mod.rootCancel = context.WithCancel(context.Background())

	return
}

type ServiceDiscoveryEvent struct {
	Service  zeroconf.ServiceEntry `json:"service"`
	Endpoint *network.Endpoint     `json:"endpoint"`
}

func (mod *ZeroGod) onServiceDiscovered(svc *zeroconf.ServiceEntry) {
	mod.Debug("%++v", *svc)

	if svc.Service == "_services._dns-sd._udp" && len(svc.AddrIPv4) == 0 && len(svc.AddrIPv6) == 0 {
		svcName := strings.Replace(svc.Instance, ".local", "", 1)
		if _, found := mod.resolvers[svcName]; !found {
			mod.Debug("discovered service %s", tui.Green(svcName))
			if err := mod.startResolver(svcName); err != nil {
				mod.Error("%v", err)
			}
		}
		return
	}

	mod.Debug("discovered instance %s (%s) [%v / %v]:%d",
		tui.Green(svc.ServiceInstanceName()),
		tui.Dim(svc.HostName),
		svc.AddrIPv4,
		svc.AddrIPv6,
		svc.Port)

	event := ServiceDiscoveryEvent{
		Service:  *svc,
		Endpoint: nil,
	}

	addresses := append(svc.AddrIPv4, svc.AddrIPv6...)

	for _, ip := range addresses {
		address := ip.String()
		if event.Endpoint = mod.Session.Lan.GetByIp(address); event.Endpoint != nil {
			// update endpoint metadata
			mod.updateEndpointMeta(address, event.Endpoint, svc)

			// update internal module mapping
			if ipServices, found := mod.mapping[address]; found {
				ipServices[svc.ServiceInstanceName()] = svc
			} else {
				mod.mapping[address] = map[string]*zeroconf.ServiceEntry{
					svc.ServiceInstanceName(): svc,
				}
			}
			break
		}
	}

	if event.Endpoint == nil {
		// TODO: this is probably an IPv6 only record, try to somehow check which known IPv4 it is
		mod.Debug("got mdns entry for unknown ip: %++v", *svc)
	}

	session.I.Events.Add("mdns.service", event)
	session.I.Refresh()
}

func (mod *ZeroGod) startResolver(service string) error {
	mod.Debug("starting resolver for service %s", tui.Yellow(service))

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	// start listening
	channel := make(chan *zeroconf.ServiceEntry)
	go func() {
		for entry := range channel {
			mod.onServiceDiscovered(entry)
		}
	}()

	// start browsing
	go func() {
		err = resolver.Browse(mod.rootContext, service, "local.", channel)
		if err != nil {
			mod.Error("%v", err)
		}
		mod.Debug("resolver for service %s stopped", tui.Yellow(service))
	}()

	mod.resolvers[service] = resolver

	return nil
}

func (mod *ZeroGod) Start() (err error) {
	if err = mod.Configure(); err != nil {
		return err
	}

	// start the root discovery
	if err = mod.startResolver("_services._dns-sd._udp"); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("service discovery started")

		<-mod.rootContext.Done()

		mod.Info("service discovery stopped")
	})
}

func (mod *ZeroGod) Stop() error {
	return mod.SetRunning(false, func() {
		mod.stopAdvertiser()

		if mod.rootCancel != nil {
			mod.Debug("stopping discovery")

			mod.rootCancel()
			<-mod.rootContext.Done()

			mod.Debug("stopped")

			mod.rootContext = nil
			mod.rootCancel = nil
		}
	})
}
