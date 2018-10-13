// +build !android

package network

import (
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/evilsocket/islazy/str"
)

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	Debug("FindGateway(%s) [cmd=%v opts=%v parser=%v]", iface.Name(), IPv4RouteCmd, IPv4RouteCmdOpts, IPv4RouteParser)

	output, err := core.Exec(IPv4RouteCmd, IPv4RouteCmdOpts)
	if err != nil {
		Debug("FindGateway(%s): core.Exec failed with %s", err)
		return nil, err
	}

	Debug("FindGateway(%s) output:\n%s", iface.Name(), output)

	ifName := iface.Name()
	for _, line := range strings.Split(output, "\n") {
		if line = str.Trim(line); strings.Contains(line, ifName) {
			m := IPv4RouteParser.FindStringSubmatch(line)
			if len(m) == IPv4RouteTokens {
				Debug("FindGateway(%s) line '%s' matched with %v", iface.Name(), line, m)
				return IPv4RouteIsGateway(ifName, m, func(gateway string) (*Endpoint, error) {
					if gateway == iface.IpAddress {
						Debug("gateway is the interface")
						return iface, nil
					} else {
						// we have the address, now we need its mac
						mac, err := ArpLookup(ifName, gateway, false)
						if err != nil {
							return nil, err
						}
						Debug("gateway is %s[%s]", gateway, mac)
						return NewEndpoint(gateway, mac), nil
					}
				})
			}
		}
	}

	Debug("FindGateway(%s): nothing found :/", iface.Name())
	return nil, ErrNoGateway
}
