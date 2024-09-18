package mdns

import (
	"fmt"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/modules/syn_scan"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

type MDNSModule struct {
	session.SessionModule

	advertiser   *Advertiser
	discoChannel chan *ServiceEntry
	mapping      map[string]map[string]*ServiceEntry
}

func NewMDNSModule(s *session.Session) *MDNSModule {
	mod := &MDNSModule{
		SessionModule: session.NewSessionModule("mdns", s),
		discoChannel:  make(chan *ServiceEntry),
		mapping:       make(map[string]map[string]*ServiceEntry),
		advertiser:    nil,
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

	if mod.discoChannel != nil {
		close(mod.discoChannel)
	}
	mod.discoChannel = make(chan *ServiceEntry)
	mod.mapping = make(map[string]map[string]*ServiceEntry)

	return
}

type ServiceDiscoveryEvent struct {
	Service  ServiceEntry      `json:"service"`
	Endpoint *network.Endpoint `json:"endpoint"`
}

func (mod *MDNSModule) updateEndpointMeta(address string, endpoint *network.Endpoint, svc *ServiceEntry) {
	mod.Debug("found endpoint %s for address %s", endpoint.HwAddress, address)

	// update mdns metadata
	meta := make(map[string]string)

	svcType := strings.SplitN(svc.Name, ".", 2)[1]

	meta[fmt.Sprintf("mdns:%s:name", svcType)] = svc.Name
	meta[fmt.Sprintf("mdns:%s:hostname", svcType)] = svc.Host

	if svc.AddrV4 != nil {
		meta[fmt.Sprintf("mdns:%s:ipv4", svcType)] = svc.AddrV4.String()
	}

	if svc.AddrV6 != nil {
		meta[fmt.Sprintf("mdns:%s:ipv6", svcType)] = svc.AddrV6.String()
	}

	meta[fmt.Sprintf("mdns:%s:port", svcType)] = fmt.Sprintf("%d", svc.Port)

	for _, field := range svc.InfoFields {
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

func (mod *MDNSModule) onServiceDiscovered(svc *ServiceEntry) {
	mod.Debug("discovered service %s (%s) [%v / %v]:%d", tui.Green(svc.Name), tui.Dim(svc.Host), svc.AddrV4, svc.AddrV6, svc.Port)

	event := ServiceDiscoveryEvent{
		Service:  *svc,
		Endpoint: nil,
	}

	addresses := []string{}
	if svc.AddrV4 != nil {
		addresses = append(addresses, svc.AddrV4.String())
	}
	if svc.AddrV6 != nil {
		addresses = append(addresses, svc.AddrV6.String())
	}

	for _, address := range addresses {
		if event.Endpoint = mod.Session.Lan.GetByIp(address); event.Endpoint != nil {
			// update endpoint metadata
			mod.updateEndpointMeta(address, event.Endpoint, svc)

			// update internal module mapping
			if ipServices, found := mod.mapping[address]; found {
				ipServices[svc.Name] = svc
			} else {
				mod.mapping[address] = map[string]*ServiceEntry{
					svc.Name: svc,
				}
			}
			break
		} else {
			mod.Warning("got mdns entry for unknown ip %s", svc.AddrV4)
		}
	}

	session.I.Events.Add("mdns.service", event)
	session.I.Refresh()
}

func (mod *MDNSModule) Start() (err error) {
	if err = mod.Configure(); err != nil {
		return err
	}

	// start the discovery
	service := "_services._dns-sd._udp"
	params := DefaultParams(service)

	params.Module = mod
	params.Service = service
	params.Domain = "local"
	params.Entries = mod.discoChannel
	params.DisableIPv6 = true // https://github.com/hashicorp/mdns/issues/35
	params.Timeout = time.Duration(10) * time.Minute

	go func() {
		mod.Info("starting query routine ...")
		if err := Query(params); err != nil {
			mod.Error("service discovery query: %v", err)
		}
		mod.Info("stopping query routine ...")
	}()

	return mod.SetRunning(true, func() {
		mod.Info("mDNS service discovery started")

		for entry := range mod.discoChannel {
			mod.onServiceDiscovered(entry)
		}

		mod.Info("mDNS service discovery stopped")
	})
}

func (mod *MDNSModule) Stop() error {
	return mod.SetRunning(false, func() {
		if mod.discoChannel != nil {
			mod.Info("closing mDNS discovery channel")
			close(mod.discoChannel)
			mod.discoChannel = nil
		}
	})
}
