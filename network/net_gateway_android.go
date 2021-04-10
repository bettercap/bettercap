package network

import (
	"github.com/bettercap/bettercap/core"

	"github.com/evilsocket/islazy/str"
)

// Hi, i'm Android and my mum said I'm special.
func FindGateway(iface *Endpoint) (*Endpoint, error) {
	output, err := core.Exec("getprop", []string{"net.dns1"})
	if err != nil {
		return nil, err
	}

	gw := str.Trim(output)
	if IPv4Validator.MatchString(gw) {
		// we have the address, now we need its mac
		mac, err := ArpLookup(iface.Name(), gw, false)
		if err != nil {
			return nil, err
		}
		return NewEndpoint(gw, mac), nil
	}

	return nil, ErrNoGateway
}