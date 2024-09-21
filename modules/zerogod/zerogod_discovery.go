package zerogod

import (
	"strings"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/tui"
)

type ServiceDiscoveryEvent struct {
	Service  zeroconf.ServiceEntry `json:"service"`
	Endpoint *network.Endpoint     `json:"endpoint"`
}

func (mod *ZeroGod) onServiceDiscovered(svc *zeroconf.ServiceEntry) {
	mod.Debug("%++v", *svc)

	if svc.Service == DNSSD_DISCOVERY_SERVICE && len(svc.AddrIPv4) == 0 && len(svc.AddrIPv6) == 0 {
		svcName := strings.Replace(svc.Instance, ".local", "", 1)
		if !mod.browser.HasResolverFor(svcName) {
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
			// update internal mapping
			mod.browser.AddServiceFor(address, svc)
			// update endpoint metadata
			mod.updateEndpointMeta(address, event.Endpoint, svc)
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

	if ch, err := mod.browser.StartBrowsing(service, "local.", mod); err != nil {
		return err
	} else {
		// start listening
		go func() {
			for entry := range ch {
				mod.onServiceDiscovered(entry)
			}
		}()
	}

	return nil
}
