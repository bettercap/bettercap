package zerogod

import (
	"context"
	"sort"
	"sync"

	"github.com/bettercap/bettercap/v2/modules/zerogod/zeroconf"
	"github.com/evilsocket/islazy/tui"
)

const DNSSD_DISCOVERY_SERVICE = "_services._dns-sd._udp"

type AddressServices struct {
	Address  string
	Services []*zeroconf.ServiceEntry
}

type Browser struct {
	sync.RWMutex

	resolvers    map[string]*zeroconf.Resolver
	servicesByIP map[string]map[string]*zeroconf.ServiceEntry
	context      context.Context
	cancel       context.CancelFunc
}

func NewBrowser() *Browser {
	servicesByIP := make(map[string]map[string]*zeroconf.ServiceEntry)
	resolvers := make(map[string]*zeroconf.Resolver)
	context, cancel := context.WithCancel(context.Background())
	return &Browser{
		resolvers:    resolvers,
		servicesByIP: servicesByIP,
		context:      context,
		cancel:       cancel,
	}
}

func (b *Browser) Wait() {
	<-b.context.Done()
}

func (b *Browser) Stop(wait bool) {
	b.cancel()
	if wait {
		b.Wait()
	}
}

func (b *Browser) HasResolverFor(service string) bool {
	b.RLock()
	defer b.RUnlock()
	_, found := b.resolvers[service]
	return found
}

func (b *Browser) AddServiceFor(ip string, svc *zeroconf.ServiceEntry) {
	b.Lock()
	defer b.Unlock()

	if ipServices, found := b.servicesByIP[ip]; found {
		ipServices[svc.ServiceInstanceName()] = svc
	} else {
		b.servicesByIP[ip] = map[string]*zeroconf.ServiceEntry{
			svc.ServiceInstanceName(): svc,
		}
	}
}

func (b *Browser) GetServicesFor(ip string) map[string]*zeroconf.ServiceEntry {
	b.RLock()
	defer b.RUnlock()

	if ipServices, found := b.servicesByIP[ip]; found {
		return ipServices
	}
	return nil
}

func (b *Browser) StartBrowsing(service string, domain string, mod *ZeroGod) (chan *zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	b.Lock()
	defer b.Unlock()

	b.resolvers[service] = resolver
	ch := make(chan *zeroconf.ServiceEntry)

	// start browsing
	go func() {
		if err := resolver.Browse(b.context, service, domain, ch); err != nil {
			mod.Error("%v", err)
		}
		mod.Debug("resolver for service %s stopped", tui.Yellow(service))
	}()

	return ch, nil
}

func (b *Browser) ServicesByAddress(filter string) []AddressServices {
	b.RLock()
	defer b.RUnlock()

	// convert to list for sorting
	entries := make([]AddressServices, 0)

	for ip, services := range b.servicesByIP {
		if filter == "" || ip == filter {
			// collect and sort services by name
			svcList := make([]*zeroconf.ServiceEntry, 0)
			for _, svc := range services {
				svcList = append(svcList, svc)
			}

			sort.Slice(svcList, func(i, j int) bool {
				return svcList[i].ServiceInstanceName() < svcList[j].ServiceInstanceName()
			})

			entries = append(entries, AddressServices{
				Address:  ip,
				Services: svcList,
			})
		}
	}

	// sort entries by ip
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Address < entries[j].Address
	})

	return entries
}
