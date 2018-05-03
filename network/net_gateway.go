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
						return nil, err
					}
					return NewEndpoint(gateway, mac), nil
				}
			})
		}
	}

	return nil, ErrNoGateway
}
