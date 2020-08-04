package network

import (
	"bufio"
	"fmt"
	"github.com/bettercap/bettercap/core"
	"strings"
)

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	Debug("FindGateway(%s) [cmd=%v opts=%v parser=%v]", iface.Name(), IPv4RouteCmd, IPv4RouteCmdOpts, IPv4RouteParser)

	output, err := core.ExecInEnglish(IPv4RouteCmd, append(IPv4RouteCmdOpts, fmt.Sprintf("%d", iface.Index)))
	if err != nil {
		Debug("FindGateway(%s): core.Exec failed with %s", err)
		return nil, err
	}

	Debug("FindGateway(%s) output:\n%s", iface.Name(), output)

	ifName := iface.Name()
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		keyPair := strings.Split(scanner.Text(), ":")
		if len(keyPair) != 2 {
			continue
		}
		key, value := strings.TrimSpace(keyPair[0]), strings.TrimSpace(keyPair[1])
		if key == "Default Gateway" {
			if value == iface.IpAddress {
				return iface, nil
			} else {
				// we have the address, now we need its mac
				mac, err := ArpLookup(ifName, value, false)
				if err != nil {
					return nil, err
				}
				Debug("gateway is %s[%s]", value, mac)
				return NewEndpoint(value, mac), nil
			}
		}
	}

	Debug("FindGateway(%s): nothing found :/", iface.Name())
	return nil, ErrNoGateway
}
