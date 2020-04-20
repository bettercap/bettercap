package http_proxy

import (
	"net"
	"sync"
)

type Host struct {
	Hostname string
	Address  net.IP
	Resolved sync.WaitGroup
}

func NewHost(name string) *Host {
	h := &Host{
		Hostname: name,
		Address:  nil,
		Resolved: sync.WaitGroup{},
	}

	h.Resolved.Add(1)
	go func(ph *Host) {
		defer ph.Resolved.Done()
		if addrs, err := net.LookupIP(ph.Hostname); err == nil && len(addrs) > 0 {
			ph.Address = make(net.IP, len(addrs[0]))
			copy(ph.Address, addrs[0])
		} else {
			ph.Address = nil
		}
	}(h)

	return h
}

type HostTracker struct {
	sync.RWMutex
	uhosts map[string]*Host
	shosts map[string]*Host
}

func NewHostTracker() *HostTracker {
	return &HostTracker{
		uhosts: make(map[string]*Host),
		shosts: make(map[string]*Host),
	}
}

func (t *HostTracker) Track(host, stripped string) {
	t.Lock()
	defer t.Unlock()
	t.uhosts[stripped] = NewHost(host)
	t.shosts[host] = NewHost(stripped)
}

func (t *HostTracker) Unstrip(stripped string) *Host {
	t.RLock()
	defer t.RUnlock()
	if host, found := t.uhosts[stripped]; found {
		return host
	}
	return nil
}


func (t *HostTracker) Strip(unstripped string) *Host {
	t.RLock()
	defer t.RUnlock()
	if host, found := t.shosts[unstripped]; found {
		return host
	}
	return nil
}
