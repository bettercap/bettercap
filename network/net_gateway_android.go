package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
)

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	output, err = core.Exec("getprop", []string{"net.dns1"})
	if err != nil {
		return nil, err
	}

	gw := core.Trim(output)
	if IPv4Validator.MatchString(gw) {
		// we have the address, now we need its mac
		mac, err := ArpLookup(iface.Name(), gw, false)
		if err != nil {
			return nil, err
		}
		return NewEndpoint(gateway, mac), nil
	}

	return nil, ErrNoGateway
}
