package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/evilsocket/bettercap-ng/core"
)

const LANDefaultTTL = 10
const LANAliasesFile = "~/bettercap.aliases"

type EndpointNewCallback func(e *Endpoint)
type EndpointLostCallback func(e *Endpoint)

type LAN struct {
	sync.Mutex

	Interface *Endpoint
	Gateway   *Endpoint
	Hosts     map[string]*Endpoint
	TTL       map[string]uint
	Aliases   map[string]string

	newCb           EndpointNewCallback
	lostCb          EndpointLostCallback
	aliasesFileName string
}

func NewLAN(iface, gateway *Endpoint, newcb EndpointNewCallback, lostcb EndpointLostCallback) *LAN {
	lan := &LAN{
		Interface: iface,
		Gateway:   gateway,
		Hosts:     make(map[string]*Endpoint),
		TTL:       make(map[string]uint),
		Aliases:   make(map[string]string),
		newCb:     newcb,
		lostCb:    lostcb,
	}

	lan.aliasesFileName, _ = core.ExpandPath(LANAliasesFile)
	if core.Exists(lan.aliasesFileName) {
		if err := lan.loadAliases(); err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return lan
}

func (lan *LAN) List() (list []*Endpoint) {
	lan.Lock()
	defer lan.Unlock()

	list = make([]*Endpoint, 0)
	for _, t := range lan.Hosts {
		list = append(list, t)
	}
	return
}

func (lan *LAN) loadAliases() error {
	file, err := os.Open(lan.aliasesFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		mac := core.Trim(parts[0])
		alias := core.Trim(parts[1])
		lan.Aliases[mac] = alias
	}

	return nil
}

func (lan *LAN) saveAliases() {
	data := ""
	for mac, alias := range lan.Aliases {
		data += fmt.Sprintf("%s %s\n", mac, alias)
	}
	ioutil.WriteFile(lan.aliasesFileName, []byte(data), 0644)
}

func (lan *LAN) SetAliasFor(mac, alias string) bool {
	lan.Lock()
	defer lan.Unlock()

	if t, found := lan.Hosts[mac]; found == true {
		if alias != "" {
			lan.Aliases[mac] = alias
		} else {
			delete(lan.Aliases, mac)
		}

		t.Alias = alias
		lan.saveAliases()
		return true
	}

	return false
}

func (lan *LAN) WasMissed(mac string) bool {
	if mac == lan.Interface.HwAddress || mac == lan.Gateway.HwAddress {
		return false
	}

	lan.Lock()
	defer lan.Unlock()

	if ttl, found := lan.TTL[mac]; found == true {
		return ttl < LANDefaultTTL
	}
	return true
}

func (lan *LAN) Remove(ip, mac string) {
	lan.Lock()
	defer lan.Unlock()

	if e, found := lan.Hosts[mac]; found {
		lan.TTL[mac]--
		if lan.TTL[mac] == 0 {
			delete(lan.Hosts, mac)
			delete(lan.TTL, mac)

			lan.lostCb(e)
		}
		return
	}
}

func (lan *LAN) shouldIgnore(ip, mac string) bool {
	// skip our own address
	if ip == lan.Interface.IpAddress {
		return true
	}
	// skip the gateway
	if ip == lan.Gateway.IpAddress {
		return true
	}
	// skip broadcast addresses
	if strings.HasSuffix(ip, ".255") {
		return true
	}
	// skip broadcast macs
	if strings.ToLower(mac) == "ff:ff:ff:ff:ff:ff" {
		return true
	}
	// skip everything which is not in our subnet (multicast noise)
	addr := net.ParseIP(ip)
	return lan.Interface.Net.Contains(addr) == false
}

func (lan *LAN) Has(ip string) bool {
	lan.Lock()
	defer lan.Unlock()

	for _, e := range lan.Hosts {
		if e.IpAddress == ip {
			return true
		}
	}

	return false
}

func (lan *LAN) Get(ip string) *Endpoint {
	lan.Lock()
	defer lan.Unlock()

	for _, e := range lan.Hosts {
		if e.IpAddress == ip {
			return e
		}
	}

	return nil
}

func (lan *LAN) AddIfNew(ip, mac string) *Endpoint {
	lan.Lock()
	defer lan.Unlock()

	if lan.shouldIgnore(ip, mac) {
		return nil
	}

	mac = NormalizeMac(mac)
	if t, found := lan.Hosts[mac]; found {
		if lan.TTL[mac] < LANDefaultTTL {
			lan.TTL[mac]++
		}
		return t
	}

	e := NewEndpoint(ip, mac)
	if alias, found := lan.Aliases[mac]; found {
		e.Alias = alias
	}

	lan.Hosts[mac] = e
	lan.TTL[mac] = LANDefaultTTL

	lan.newCb(e)

	return nil
}
