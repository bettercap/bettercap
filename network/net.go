package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
)

var ErrNoIfaces = errors.New("No active interfaces found.")

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
	return strings.Join(parts, ":")
}

func buildEndpointFromInterface(iface net.Interface) (*Endpoint, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	e := NewEndpointNoResolve(MonitorModeAddress, iface.HardwareAddr.String(), "", 0)

	e.Hostname = iface.Name
	e.Index = iface.Index

	for _, addr := range addrs {
		ip := addr.String()

		if IPv4Validator.MatchString(ip) {
			if strings.IndexRune(ip, '/') == -1 {
				// plain ip
				e.SetIP(ip)
			} else {
				// ip/bits
				parts := strings.Split(ip, "/")
				ip_part := parts[0]
				bits, _ := strconv.Atoi(parts[1])

				e.SetIP(ip_part)
				e.SubnetBits = uint32(bits)
			}
		} else {
			parts := strings.SplitN(ip, "/", 2)
			e.IPv6 = net.ParseIP(parts[0])
			if e.IPv6 != nil {
				e.Ip6Address = e.IPv6.String()
			}
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
		if iface.Name == name || matchByAddress(iface, name) {
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

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	output, err := core.Exec(IPv4RouteCmd, IPv4RouteCmdOpts)
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(output, "\n") {
		m := IPv4RouteParser.FindStringSubmatch(line)
		if len(m) == IPv4RouteTokens {
			return IPv4RouteIsGateway(iface.Name(), m, func(gateway string) (*Endpoint, error) {
				if gateway == iface.IpAddress {
					return iface, nil
				} else {
					// we have the address, now we need its mac
					mac, err := ArpLookup(iface.Name(), gateway, false)
					if err != nil {
						fmt.Printf("%s\n", err)
					}
					return NewEndpoint(gateway, mac), nil
				}
			})
		}
	}

	return nil, fmt.Errorf("Could not detect the gateway.")
}
