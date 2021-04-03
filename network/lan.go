package network

import (
	"encoding/json"
	"net"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/data"
)

const LANDefaultttl = 10

type EndpointNewCallback func(e *Endpoint)
type EndpointLostCallback func(e *Endpoint)

type LAN struct {
	sync.Mutex
	hosts   map[string]*Endpoint
	iface   *Endpoint
	gateway *Endpoint
	ttl     map[string]uint
	aliases *data.UnsortedKV
	newCb   EndpointNewCallback
	lostCb  EndpointLostCallback
}

type lanJSON struct {
	Hosts []*Endpoint `json:"hosts"`
}

func NewLAN(iface, gateway *Endpoint, aliases *data.UnsortedKV, newcb EndpointNewCallback, lostcb EndpointLostCallback) *LAN {
	return &LAN{
		iface:   iface,
		gateway: gateway,
		hosts:   make(map[string]*Endpoint),
		ttl:     make(map[string]uint),
		aliases: aliases,
		newCb:   newcb,
		lostCb:  lostcb,
	}
}

func (l *LAN) MarshalJSON() ([]byte, error) {
	doc := lanJSON{
		Hosts: make([]*Endpoint, 0),
	}

	for _, h := range l.hosts {
		doc.Hosts = append(doc.Hosts, h)
	}

	return json.Marshal(doc)
}

func (lan *LAN) Get(mac string) (*Endpoint, bool) {
	lan.Lock()
	defer lan.Unlock()

	mac = NormalizeMac(mac)

	if mac == lan.iface.HwAddress {
		return lan.iface, true
	} else if mac == lan.gateway.HwAddress {
		return lan.gateway, true
	}

	if e, found := lan.hosts[mac]; found {
		return e, true
	}
	return nil, false
}

func (lan *LAN) GetByIp(ip string) *Endpoint {
	lan.Lock()
	defer lan.Unlock()

	if ip == lan.iface.IpAddress {
		return lan.iface
	} else if ip == lan.gateway.IpAddress {
		return lan.gateway
	}

	for _, e := range lan.hosts {
		if e.IpAddress == ip {
			return e
		}
	}

	return nil
}

func (lan *LAN) List() (list []*Endpoint) {
	lan.Lock()
	defer lan.Unlock()

	list = make([]*Endpoint, 0)
	for _, t := range lan.hosts {
		list = append(list, t)
	}
	return
}

func (lan *LAN) Aliases() *data.UnsortedKV {
	return lan.aliases
}

func (lan *LAN) WasMissed(mac string) bool {
	if mac == lan.iface.HwAddress || mac == lan.gateway.HwAddress {
		return false
	}

	lan.Lock()
	defer lan.Unlock()

	if ttl, found := lan.ttl[mac]; found {
		return ttl < LANDefaultttl
	}
	return true
}

func (lan *LAN) Remove(ip, mac string) {
	lan.Lock()
	defer lan.Unlock()

	if e, found := lan.hosts[mac]; found {
		lan.ttl[mac]--
		if lan.ttl[mac] == 0 {
			delete(lan.hosts, mac)
			delete(lan.ttl, mac)
			lan.lostCb(e)
		}
		return
	}
}

func (lan *LAN) shouldIgnore(ip, mac string) bool {
	// skip our own address
	if ip == lan.iface.IpAddress || mac == lan.iface.HwAddress {
		return true
	}
	// skip the gateway
	if ip == lan.gateway.IpAddress || mac == lan.gateway.HwAddress {
		return true
	}
	// skip broadcast addresses
	if strings.HasSuffix(ip, BroadcastSuffix) {
		return true
	}
	// skip broadcast macs
	if strings.ToLower(mac) == BroadcastMac {
		return true
	}
	// skip everything which is not in our subnet (multicast noise)
	addr := net.ParseIP(ip)
	return addr.To4() != nil && !lan.iface.Net.Contains(addr)
}

func (lan *LAN) Has(ip string) bool {
	lan.Lock()
	defer lan.Unlock()

	for _, e := range lan.hosts {
		if e.IpAddress == ip {
			return true
		}
	}

	return false
}

func (lan *LAN) EachHost(cb func(mac string, e *Endpoint)) {
	lan.Lock()
	defer lan.Unlock()

	for m, h := range lan.hosts {
		cb(m, h)
	}
}

func (lan *LAN) AddIfNew(ip, mac string) *Endpoint {
	lan.Lock()
	defer lan.Unlock()

	mac = NormalizeMac(mac)

	if lan.shouldIgnore(ip, mac) {
		return nil
	} else if t, found := lan.hosts[mac]; found {
		if lan.ttl[mac] < LANDefaultttl {
			lan.ttl[mac]++
		}
		return t
	}

	e := NewEndpointWithAlias(ip, mac, lan.aliases.GetOr(mac, ""))

	lan.hosts[mac] = e
	lan.ttl[mac] = LANDefaultttl

	lan.newCb(e)

	return nil
}

func (lan *LAN) GetAlias(mac string) string {
	return lan.aliases.GetOr(mac, "")
}

func (lan *LAN) Clear() {
	lan.Lock()
	defer lan.Unlock()
	lan.hosts = make(map[string]*Endpoint)
	lan.ttl = make(map[string]uint)
}
