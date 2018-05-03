// +build !android

package network

import (
	"strings"

	"github.com/bettercap/bettercap/core"
)

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	output, err := core.Exec(IPv4RouteCmd, IPv4RouteCmdOpts)
	if err != nil {
		return nil, err
	}

	ifName := iface.Name()
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, ifName) {
			m := IPv4RouteParser.FindStringSubmatch(line)
			if len(m) == IPv4RouteTokens {
				return IPv4RouteIsGateway(ifName, m, func(gateway string) (*Endpoint, error) {
					if gateway == iface.IpAddress {
						return iface, nil
					} else {
						// we have the address, now we need its mac
						mac, err := ArpLookup(ifName, gateway, false)
						if err != nil {
							return nil, err
						}
						return NewEndpoint(gateway, mac), nil
					}
				})
			}
		}
	}

	return nil, ErrNoGateway
}
