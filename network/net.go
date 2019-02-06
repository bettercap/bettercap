package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/evilsocket/islazy/data"
	"github.com/evilsocket/islazy/str"

	"github.com/malfunkt/iprange"
)

var ErrNoIfaces = errors.New("No active interfaces found.")
var ErrNoGateway = errors.New("Could not detect gateway.")

const (
	MonitorModeAddress = "0.0.0.0"
	BroadcastSuffix    = ".255"
	BroadcastMac       = "ff:ff:ff:ff:ff:ff"
	IPv4MulticastStart = "01:00:5e:00:00:00"
	IPv4MulticastEnd   = "01:00:5e:7f:ff:ff"
)

var (
	BroadcastHw        = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	IPv4Validator      = regexp.MustCompile(`^[0-9\.]+/?\d*$`)
	IPv4RangeValidator = regexp.MustCompile(`^[0-9\.\-]+/?\d*$`)
	MACValidator       = regexp.MustCompile(`(?i)^[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}$`)
	// lulz this sounds like a hamburger
	macParser   = regexp.MustCompile(`(?i)([a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2}:[a-f0-9]{1,2})`)
	aliasParser = regexp.MustCompile(`(?i)([a-z_][a-z_0-9]+)`)
)

func IsZeroMac(mac net.HardwareAddr) bool {
	for _, b := range mac {
		if b != 0x00 {
			return false
		}
	}
	return true
}

func IsBroadcastMac(mac net.HardwareAddr) bool {
	for _, b := range mac {
		if b != 0xff {
			return false
		}
	}
	return true
}

func NormalizeMac(mac string) string {
	var parts []string
	if strings.ContainsRune(mac, '-') {
		parts = strings.Split(mac, "-")
	} else {
		parts = strings.Split(mac, ":")
	}

	for i, p := range parts {
		if len(p) < 2 {
			parts[i] = "0" + p
		}
	}
	return strings.ToLower(strings.Join(parts, ":"))
}

func ParseMACs(targets string) (macs []net.HardwareAddr, err error) {
	macs = make([]net.HardwareAddr, 0)
	if targets = str.Trim(targets); targets == "" {
		return
	}

	for _, mac := range macParser.FindAllString(targets, -1) {
		mac = NormalizeMac(mac)
		hw, err := net.ParseMAC(mac)
		if err != nil {
			return nil, fmt.Errorf("Error while parsing MAC '%s': %s", mac, err)
		}

		macs = append(macs, hw)
		targets = strings.Replace(targets, mac, "", -1)
	}

	return
}

func ParseTargets(targets string, aliasMap *data.UnsortedKV) (ips []net.IP, macs []net.HardwareAddr, err error) {
	ips = make([]net.IP, 0)
	macs = make([]net.HardwareAddr, 0)

	if targets = str.Trim(targets); targets == "" {
		return
	}

	// first isolate MACs and parse them
	for _, mac := range macParser.FindAllString(targets, -1) {
		mac = NormalizeMac(mac)
		hw, err := net.ParseMAC(mac)
		if err != nil {
			return nil, nil, fmt.Errorf("Error while parsing MAC '%s': %s", mac, err)
		}

		macs = append(macs, hw)
		targets = strings.Replace(targets, mac, "", -1)
	}
	targets = strings.Trim(targets, ", ")

	// check and resolve aliases
	for _, alias := range aliasParser.FindAllString(targets, -1) {
		if mac, found := aliasMap.Get(alias); found {
			mac = NormalizeMac(mac)
			hw, err := net.ParseMAC(mac)
			if err != nil {
				return nil, nil, fmt.Errorf("Error while parsing MAC '%s': %s", mac, err)
			}

			macs = append(macs, hw)
			targets = strings.Replace(targets, alias, "", -1)
		} else {
			return nil, nil, fmt.Errorf("Could not resolve alias %s", alias)
		}
	}
	targets = strings.Trim(targets, ", ")

	// parse what's left
	if targets != "" {
		list, err := iprange.ParseList(targets)
		if err != nil {
			return nil, nil, fmt.Errorf("Error while parsing address list '%s': %s.", targets, err)
		}

		ips = list.Expand()
	}

	return
}

func ParseEndpoints(targets string, lan *LAN) ([]*Endpoint, error) {
	ips, macs, err := ParseTargets(targets, lan.Aliases())
	if err != nil {
		return nil, err
	}

	tmp := make(map[string]*Endpoint)
	for _, ip := range ips {
		if e := lan.GetByIp(ip.String()); e != nil {
			tmp[e.HW.String()] = e
		}
	}

	for _, mac := range macs {
		if e, found := lan.Get(mac.String()); found {
			tmp[e.HW.String()] = e
		}
	}

	ret := make([]*Endpoint, 0)
	for _, e := range tmp {
		ret = append(ret, e)
	}
	return ret, nil
}

func buildEndpointFromInterface(iface net.Interface) (*Endpoint, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	ifName := getInterfaceName(iface)

	e := NewEndpointNoResolve(MonitorModeAddress, iface.HardwareAddr.String(), ifName, 0)

	e.Index = iface.Index

	for _, a := range addrs {
		address := a.String()
		if IPv4Validator.MatchString(address) {
			if !strings.ContainsRune(address, '/') {
				// plain ip
				e.SetIP(address)
			} else {
				// ip/bits
				e.SetNetwork(address)
			}
		} else {
			// ipv6/xxx
			e.SetIPv6(address)
		}
	}

	return e, nil
}

func matchByAddress(iface net.Interface, name string) bool {
	ifMac := iface.HardwareAddr.String()
	if NormalizeMac(ifMac) == NormalizeMac(name) {
		return true
	}

	addrs, err := iface.Addrs()
	if err == nil {
		for _, addr := range addrs {
			ip := addr.String()
			if ip == name || strings.HasPrefix(ip, name) {
				return true
			}
		}
	}

	return false
}

func findInterfaceByName(name string, ifaces []net.Interface) (*Endpoint, error) {
	for _, iface := range ifaces {
		ifName := getInterfaceName(iface)
		if ifName == name || matchByAddress(iface, name) {
			return buildEndpointFromInterface(iface)
		}
	}

	return nil, fmt.Errorf("No interface matching '%s' found.", name)
}

func FindInterface(name string) (*Endpoint, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	name = str.Trim(name)
	if name != "" {
		return findInterfaceByName(name, ifaces)
	}

	// user did not provide an interface name,
	// return the first one with a valid ipv4
	// address
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("WTF of the day: %s", err)
			continue
		}

		for _, address := range addrs {
			ip := address.String()
			if !strings.Contains(ip, "127.0.0.1") && IPv4Validator.MatchString(ip) {
				return buildEndpointFromInterface(iface)
			}
		}
	}

	return nil, ErrNoIfaces
}

func ActivateInterface(name string) error {
	if out, err := core.Exec("ifconfig", []string{name, "up"}); err != nil {
		return err
	} else if out != "" {
		return fmt.Errorf("unexpected output while activating interface %s: %s", name, out)
	}
	return nil
}
