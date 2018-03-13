package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"
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
	BroadcastHw   = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	IPv4Validator = regexp.MustCompile("^[0-9\\.]+/?\\d*$")
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
			if strings.IndexRune(address, '/') == -1 {
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
	name = core.Trim(name)
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
			if strings.Contains(ip, "127.0.0.1") == false && IPv4Validator.MatchString(ip) {
				return buildEndpointFromInterface(iface)
			}
		}
	}

	return nil, ErrNoIfaces
}
