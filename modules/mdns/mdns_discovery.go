package mdns

import (
	"context"
	"fmt"
	"strings"

	"github.com/bettercap/bettercap/v2/modules/syn_scan"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"

	"github.com/grandcat/zeroconf"
)

type MDNSModule struct {
	session.SessionModule

	advertiser  *Advertiser
	rootContext context.Context
	rootCancel  context.CancelFunc
	resolvers   map[string]*zeroconf.Resolver
	mapping     map[string]map[string]*zeroconf.ServiceEntry
}

func NewMDNSModule(s *session.Session) *MDNSModule {
	mod := &MDNSModule{
		SessionModule: session.NewSessionModule("mdns", s),
		mapping:       make(map[string]map[string]*zeroconf.ServiceEntry),
		resolvers:     make(map[string]*zeroconf.Resolver),
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddHandler(session.NewModuleHandler("mdns.discovery on", "",
		"Start DNS-SD / mDNS discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.discovery off", "",
		"Stop DNS-SD / mDNS discovery.",
		func(args []string) error {
			return mod.Stop()
		}))

	// TODO: add autocomplete
	mod.AddHandler(session.NewModuleHandler("mdns.show", "",
		"Show discovered services.",
		func(args []string) error {
			return mod.show("", false)
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.show-full", "",
		"Show discovered services and their DNS records.",
		func(args []string) error {
			return mod.show("", true)
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.show ADDRESS", "mdns.show (.+)",
		"Show discovered services given an ip address.",
		func(args []string) error {
			return mod.show(args[0], false)
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.show-full ADDRESS", "mdns.show-full (.+)",
		"Show discovered services and DNS records given an ip address.",
		func(args []string) error {
			return mod.show(args[0], true)
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.save ADDRESS FILENAME", "mdns.save (.+) (.+)",
		"Save the mDNS information of a given ADDRESS in the FILENAME yaml file.",
		func(args []string) error {
			return mod.save(args[0], args[1])
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.advertise FILENAME", "mdns.advertise (.+)",
		"Start advertising the mDNS services from the FILENAME yaml file.",
		func(args []string) error {
			if args[0] == "off" {
				return mod.stopAdvertiser()
			}
			return mod.startAdvertiser(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.advertise off", "",
		"Start a previously started advertiser.",
		func(args []string) error {
			return mod.stopAdvertiser()
		}))

	return mod
}

func (mod *MDNSModule) Name() string {
	return "mdns"
}

func (mod *MDNSModule) Description() string {
	return "A DNS-SD / mDNS module for discovery and spoofing."
}

func (mod *MDNSModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *MDNSModule) Configure() (err error) {
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

func (mod *MDNSModule) updateEndpointMeta(address string, endpoint *network.Endpoint, svc *zeroconf.ServiceEntry) {
	mod.Debug("found endpoint %s for address %s", endpoint.HwAddress, address)

	// TODO: this is shit and needs to be refactored

	// update mdns metadata
	meta := make(map[string]string)

	svcType := svc.Service

	meta[fmt.Sprintf("mdns:%s:name", svcType)] = svc.ServiceName()
	meta[fmt.Sprintf("mdns:%s:hostname", svcType)] = svc.HostName

	// TODO: include all
	if len(svc.AddrIPv4) > 0 {
		meta[fmt.Sprintf("mdns:%s:ipv4", svcType)] = svc.AddrIPv4[0].String()
	}

	if len(svc.AddrIPv6) > 0 {
		meta[fmt.Sprintf("mdns:%s:ipv6", svcType)] = svc.AddrIPv6[0].String()
	}

	meta[fmt.Sprintf("mdns:%s:port", svcType)] = fmt.Sprintf("%d", svc.Port)

	for _, field := range svc.Text {
		field = str.Trim(field)
		if len(field) == 0 {
			continue
		}

		key := ""
		value := ""

		if strings.Contains(field, "=") {
			parts := strings.SplitN(field, "=", 2)
			key = parts[0]
			value = parts[1]
		} else {
			key = field
		}

		meta[fmt.Sprintf("mdns:%s:info:%s", svcType, key)] = value
	}

	mod.Debug("meta for %s: %v", address, meta)

	endpoint.OnMeta(meta)

	// update ports
	ports := endpoint.Meta.GetOr("ports", map[int]*syn_scan.OpenPort{}).(map[int]*syn_scan.OpenPort)
	if _, found := ports[svc.Port]; !found {
		ports[svc.Port] = &syn_scan.OpenPort{
			Proto:   "tcp",
			Port:    svc.Port,
			Service: network.GetServiceByPort(svc.Port, "tcp"),
		}
	}

	endpoint.Meta.Set("ports", ports)
}

func (mod *MDNSModule) onServiceDiscovered(svc *zeroconf.ServiceEntry) {
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

func (mod *MDNSModule) startResolver(service string) error {
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

func (mod *MDNSModule) Start() (err error) {
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

func (mod *MDNSModule) Stop() error {
	return mod.SetRunning(false, func() {
		if mod.rootCancel != nil {
			mod.Debug("stopping mDNS discovery")

			mod.rootCancel()
			<-mod.rootContext.Done()

			mod.Debug("stopped")

			mod.rootContext = nil
			mod.rootCancel = nil
		}
	})
}
