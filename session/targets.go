package session

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
)

const TargetsDefaultTTL = 2
const TargetsAliasesFile = "~/bettercap.aliases"

type Targets struct {
	sync.Mutex

	Session   *Session `json:"-"`
	Interface *net.Endpoint
	Gateway   *net.Endpoint
	Targets   map[string]*net.Endpoint
	TTL       map[string]uint
	Aliases   map[string]string

	aliasesFileName string
}

func NewTargets(s *Session, iface, gateway *net.Endpoint) *Targets {
	t := &Targets{
		Session:   s,
		Interface: iface,
		Gateway:   gateway,
		Targets:   make(map[string]*net.Endpoint),
		TTL:       make(map[string]uint),
		Aliases:   make(map[string]string),
	}

	t.aliasesFileName, _ = core.ExpandPath(TargetsAliasesFile)
	if core.Exists(t.aliasesFileName) {
		if err := t.loadAliases(); err != nil {
			s.Events.Log(core.ERROR, "%s", err)
		}
	}

	return t
}

func (tp *Targets) List() (list []*net.Endpoint) {
	tp.Lock()
	defer tp.Unlock()

	list = make([]*net.Endpoint, 0)
	for _, t := range tp.Targets {
		list = append(list, t)
	}
	return
}

func (tp *Targets) loadAliases() error {
	tp.Session.Events.Log(core.INFO, "Loading aliases from %s ...", tp.aliasesFileName)
	file, err := os.Open(tp.aliasesFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		mac := strings.Trim(parts[0], "\r\n\t ")
		alias := strings.Trim(parts[1], "\r\n\t ")
		tp.Session.Events.Log(core.DEBUG, " aliases[%s] = '%s'", mac, alias)
		tp.Aliases[mac] = alias
	}

	return nil
}

func (tp *Targets) saveAliases() {
	data := ""
	for mac, alias := range tp.Aliases {
		data += fmt.Sprintf("%s %s\n", mac, alias)
	}
	ioutil.WriteFile(tp.aliasesFileName, []byte(data), 0644)
}

func (tp *Targets) SetAliasFor(mac, alias string) bool {
	tp.Lock()
	defer tp.Unlock()

	if t, found := tp.Targets[mac]; found == true {
		tp.Aliases[mac] = alias
		t.Alias = alias
		tp.saveAliases()
		return true
	}

	return false
}

func (tp *Targets) WasMissed(mac string) bool {
	if mac == tp.Session.Interface.HwAddress || mac == tp.Session.Gateway.HwAddress {
		return false
	}

	tp.Lock()
	defer tp.Unlock()

	if ttl, found := tp.TTL[mac]; found == true {
		return ttl < TargetsDefaultTTL
	}
	return true
}

func (tp *Targets) Remove(ip, mac string) {
	tp.Lock()
	defer tp.Unlock()

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
	tp.Lock()
	defer tp.Unlock()

	for _, e := range tp.Targets {
		if e.IpAddress == ip {
			return true
		}
	}

	return false
}

func (tp *Targets) AddIfNotExist(ip, mac string) *net.Endpoint {
	tp.Lock()
	defer tp.Unlock()

	if tp.shouldIgnore(ip) {
		return nil
	}

	mac = net.NormalizeMac(mac)
	if t, found := tp.Targets[mac]; found {
		if tp.TTL[mac] < TargetsDefaultTTL {
			tp.TTL[mac]++
		}
		return t
	}

	e := net.NewEndpoint(ip, mac)
	e.ResolvedCallback = func(e *net.Endpoint) {
		tp.Session.Events.Add("target.resolved", e)
	}

	if alias, found := tp.Aliases[mac]; found {
		e.Alias = alias
	}

	tp.Targets[mac] = e
	tp.TTL[mac] = TargetsDefaultTTL

	tp.Session.Events.Add("target.new", e)

	return nil
}
