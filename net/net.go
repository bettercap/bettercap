package net

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
)

const MonitorModeAddress = "0.0.0.0"

func FindInterface(name string) (*Endpoint, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		mac := iface.HardwareAddr.String()
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}
		nAddrs := len(addrs)

		/*
		 * If no interface has been specified, return the first active
		 * one with at least an ip address, otherwise just the match
		 * whatever it has, in order to also consider monitor interfaces
		 * if passed explicitly.
		 */
		doCheck := false
		if name == "" && iface.Name != "lo" && iface.Name != "lo0" && nAddrs > 0 {
			doCheck = true
		} else if iface.Name == name {
			doCheck = true
		}

		if doCheck {
			var e *Endpoint = nil
			// interface is in monitor mode (or it's just down and the user is dumb)
			if nAddrs == 0 {
				e = NewEndpointNoResolve(MonitorModeAddress, mac, iface.Name, 0)
			} else {
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
			}

			if e != nil {
				if len(e.HW) == 0 {
					return nil, fmt.Errorf("Could not detect interface hardware address.")
				}
				return e, nil
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
