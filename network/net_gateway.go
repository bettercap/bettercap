// +build !android

package network

import (
	"github.com/bettercap/bettercap/routing"
)

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	gateway, err := routing.Gateway(routing.IPv4, iface.Name())
	if err != nil {
		return nil, err
	}

	if gateway == iface.IpAddress {
		Debug("gateway is the interface")
		return iface, nil
	} else {
		// we have the address, now we need its mac
		mac, err := ArpLookup(iface.Name(), gateway, false)
		if err != nil {
			return nil, err
		}
		Debug("gateway is %s[%s]", gateway, mac)
		return NewEndpoint(gateway, mac), nil
	}

	Debug("FindGateway(%s): nothing found :/", iface.Name())
	return nil, ErrNoGateway
}
