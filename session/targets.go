package session

import (
	"sync"

	"github.com/evilsocket/bettercap-ng/net"
)

type Targets struct {
	Session   *Session `json:"-"`
	Interface *net.Endpoint
	Gateway   *net.Endpoint
	Targets   map[string]*net.Endpoint
	TTL       map[string]uint
	lock      sync.Mutex
}

func NewTargets(s *Session, iface, gateway *net.Endpoint) *Targets {
	return &Targets{
		Session:   s,
		Interface: iface,
		Gateway:   gateway,
		Targets:   make(map[string]*net.Endpoint),
		TTL:       make(map[string]uint),
	}
}

func (tp *Targets) Lock() {
	tp.lock.Lock()
}

func (tp *Targets) Unlock() {
	tp.lock.Unlock()
}

func (tp *Targets) Remove(ip, mac string) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	if e, found := tp.Targets[mac]; found {
		tp.TTL[mac]--
		if tp.TTL[mac] == 0 {
			tp.Session.Events.Add("target.lost", e)
			delete(tp.Targets, mac)
			delete(tp.TTL, mac)
		}
		return
	}
}

func (tp *Targets) shouldIgnore(ip string) bool {
	return (ip == tp.Interface.IpAddress || ip == tp.Gateway.IpAddress)
}

func (tp *Targets) Has(ip string) bool {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	for _, e := range tp.Targets {
		if e.IpAddress == ip {
			return true
		}
	}

	return false
}

func (tp *Targets) AddIfNotExist(ip, mac string) *net.Endpoint {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	if tp.shouldIgnore(ip) {
		return nil
	}

	if t, found := tp.Targets[mac]; found {
		return t
	}

	e := net.NewEndpoint(ip, mac)
	e.ResolvedCallback = func(e *net.Endpoint) {
		tp.Session.Events.Add("target.resolved", e)
	}

	tp.Targets[mac] = e
	tp.TTL[mac] = 2

	tp.Session.Events.Add("target.new", e)

	return nil
}
