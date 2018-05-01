package network

import (
	"net"
	"testing"
)

func buildExampleLAN() *LAN {
	iface, _ := FindInterface("")
	gateway, _ := FindGateway(iface)
	exNewCallback := func(e *Endpoint) {}
	exLostCallback := func(e *Endpoint) {}
	return NewLAN(iface, gateway, exNewCallback, exLostCallback)
}

func buildExampleEndpoint() *Endpoint {
	ifaces, _ := net.Interfaces()
	var exampleIface net.Interface
	for _, iface := range ifaces {
		if iface.HardwareAddr != nil {
			exampleIface = iface
			break
		}
	}
	foundEndpoint, _ := FindInterface(exampleIface.Name)
	return foundEndpoint
}
