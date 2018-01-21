package net

import (
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
)

var IPv4RouteParser = regexp.MustCompile("^([\\d\\.]+)\\s+([\\d\\.]+)\\s+([\\d\\.]+)\\s+([A-Z]+)\\s+\\d+\\s+\\d+\\s+\\d+\\s+(.+)$")
var IPv4RouteTokens = 6
var IPv4RouteParserMac = regexp.MustCompile("^([a-z]+)+\\s+(\\d+\\.+\\d+.\\d.+\\d)+\\s+([a-zA-z]+)+\\s+(\\d+)+\\s+(\\d+)+\\s+([a-zA-Z]+\\d+)$")
var IPv4RouteTokensMac = 7
var IPv4RouteGWFlags = "UG"
var IPv4RouteGWFlagsMac = "UGSc"

func FindInterface(name string) (*Endpoint, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		mac := iface.HardwareAddr.String()
		addrs, err := iface.Addrs()
		// is interface active?
		if err == nil && len(addrs) > 0 {
			if (name == "" && iface.Name != "lo" && iface.Name != "lo0") || iface.Name == name {
				var e *Endpoint = nil
				// For every address of the interface.
				for _, addr := range addrs {
					ip := addr.String()
					// Make sure this is an IPv4 address.
					if m, _ := regexp.MatchString("^[0-9\\.]+/?\\d*$", ip); m == true {
						if strings.IndexRune(ip, '/') == -1 {
							// plain ip
							e = NewEndpointNoResolve(ip, mac, iface.Name, 0)
						} else {
							// ip/bits
							parts := strings.Split(ip, "/")
							ip_part := parts[0]
							bits, err := strconv.Atoi(parts[1])
							if err == nil {
								e = NewEndpointNoResolve(ip_part, mac, iface.Name, uint32(bits))
							}
						}
					} else if e != nil {
						parts := strings.SplitN(ip, "/", 2)
						e.IPv6 = net.ParseIP(parts[0])
						if e.IPv6 != nil {
							e.Ip6Address = e.IPv6.String()
						}
					}
				}

				if e != nil {
					if len(e.HW) == 0 {
						return nil, fmt.Errorf("Could not detect interface hardware address.")
					}
					return e, nil
				}
			}
		}
	}

	if name == "" {
		return nil, fmt.Errorf("Could not find default network interface.")
	} else {
		return nil, fmt.Errorf("Could not find interface '%s'.", name)
	}
}

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	var routeParser = IPv4RouteParser
	var routeTokens = IPv4RouteTokens
	var flagsIndex = 4
	var ifnameIndex = 5
	var gwFlags = IPv4RouteGWFlags

	if runtime.GOOS == "darwin" {
		// "MacOS detected"
		routeParser = IPv4RouteParserMac
		routeTokens = IPv4RouteTokensMac
		flagsIndex = 3
		ifnameIndex = 6
		gwFlags = IPv4RouteGWFlagsMac
	}

	output, err := core.Exec("netstat", []string{"-n", "-r"})
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(output, "\n") {
		m := routeParser.FindStringSubmatch(line)
		if len(m) == routeTokens {
			// destination := m[1]
			// mask := m[3]
			flags := m[flagsIndex]
			ifname := m[ifnameIndex]
			if ifname == iface.Name() && flags == gwFlags {
				gateway := m[2]
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
			}
		}
	}

	return nil, fmt.Errorf("Could not detect the gateway.")
}
