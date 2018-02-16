package modules

import (
	"sync"
	"time"

	bnet "github.com/evilsocket/bettercap-ng/network"
	session "github.com/evilsocket/bettercap-ng/session"
)

const TargetsDefaultTTL = 30

type WlanEndpoint struct {
	Endpoint *bnet.Endpoint
	Essid    string
	IsAP     bool
	Channel  int
}

func NewWlanEndpoint(essid, mac string, isAp bool, channel int) *WlanEndpoint {
	e := bnet.NewEndpointNoResolve("0.0.0.0", mac, "", 0)

	we := &WlanEndpoint{
		Endpoint: e,
		Essid:    essid,
		IsAP:     isAp,
		Channel:  channel,
	}

	return we
}

type WlanTargets struct {
	sync.Mutex

	Session   *session.Session `json:"-"`
	Interface *bnet.Endpoint
	Targets   map[string]*WlanEndpoint
	TTL       map[string]uint
	Aliases   map[string]string
}

func NewWlanTargets(s *session.Session, iface *bnet.Endpoint) *WlanTargets {
	t := &WlanTargets{
		Session:   s,
		Interface: iface,
		Targets:   make(map[string]*WlanEndpoint),
		TTL:       make(map[string]uint),
		Aliases:   s.Targets.Aliases,
	}

	return t
}

func (tp *WlanTargets) List() (list []*WlanEndpoint) {
	tp.Lock()
	defer tp.Unlock()

	list = make([]*WlanEndpoint, 0)
	for _, t := range tp.Targets {
		list = append(list, t)
	}
	return
}

func (tp *WlanTargets) WasMissed(mac string) bool {
	if mac == tp.Session.Interface.HwAddress {
		return false
	}

	tp.Lock()
	defer tp.Unlock()

	if ttl, found := tp.TTL[mac]; found == true {
		return ttl < TargetsDefaultTTL
	}
	return true
}

func (tp *WlanTargets) Remove(mac string) {
	tp.Lock()
	defer tp.Unlock()

	if e, found := tp.Targets[mac]; found {
		tp.TTL[mac]--
		if tp.TTL[mac] == 0 {
			tp.Session.Events.Add("endpoint.lost", e.Endpoint)
			delete(tp.Targets, mac)
			delete(tp.TTL, mac)
		}
		return
	}
}

func (tp *WlanTargets) AddIfNew(ssid, mac string, isAp bool, channel int) *WlanEndpoint {
	tp.Lock()
	defer tp.Unlock()

	mac = bnet.NormalizeMac(mac)
	if t, found := tp.Targets[mac]; found {
		if tp.TTL[mac] < TargetsDefaultTTL {
			tp.TTL[mac]++
		}

		tp.Targets[mac].Endpoint.LastSeen = time.Now()

		return t
	}

	e := NewWlanEndpoint(ssid, mac, isAp, channel)

	if alias, found := tp.Aliases[mac]; found {
		e.Endpoint.Alias = alias
	}

	tp.Targets[mac] = e
	tp.TTL[mac] = TargetsDefaultTTL

	tp.Session.Events.Add("endpoint.new", e.Endpoint)

	return nil
}

func (tp *WlanTargets) ClearAll() error {
	tp.Targets = make(map[string]*WlanEndpoint)
	tp.TTL = make(map[string]uint)
	tp.Aliases = make(map[string]string)

	return nil
}
