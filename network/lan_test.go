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

func TestNewLAN(t *testing.T) {
	iface, err := FindInterface("")
	if err != nil {
		t.Error("no iface found", err)
	}
	gateway, err := FindGateway(iface)
	if err != nil {
		t.Error("no gateway found", err)
	}
	exNewCallback := func(e *Endpoint) {}
	exLostCallback := func(e *Endpoint) {}
	lan := NewLAN(iface, gateway, exNewCallback, exLostCallback)
	if lan.iface != iface {
		t.Fatalf("expected '%v', got '%v'", iface, lan.iface)
	}
	if lan.gateway != gateway {
		t.Fatalf("expected '%v', got '%v'", gateway, lan.gateway)
	}
	if len(lan.hosts) != 0 {
		t.Fatalf("expected '%v', got '%v'", 0, len(lan.hosts))
	}
	if !(len(lan.aliases.data) >= 0) {
		t.Fatalf("expected '%v', got '%v'", 0, len(lan.aliases.data))
	}
}

func TestMarshalJSON(t *testing.T) {
	iface, err := FindInterface("")
	if err != nil {
		t.Error("no iface found", err)
	}
	gateway, err := FindGateway(iface)
	if err != nil {
		t.Error("no gateway found", err)
	}
	exNewCallback := func(e *Endpoint) {}
	exLostCallback := func(e *Endpoint) {}
	lan := NewLAN(iface, gateway, exNewCallback, exLostCallback)
	_, err = lan.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
}

func TestSetAliasFor(t *testing.T) {
	exampleAlias := "picat"
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	if !exampleLAN.SetAliasFor(exampleEndpoint.HwAddress, exampleAlias) {
		t.Error("unable to set alias for a given mac address")
	}
}
