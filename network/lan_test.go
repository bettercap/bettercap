package network

import (
	"testing"

	"github.com/evilsocket/islazy/data"
)

func buildExampleLAN() *LAN {
	iface, _ := FindInterface("")
	gateway, _ := FindGateway(iface)
	exNewCallback := func(e *Endpoint) {}
	exLostCallback := func(e *Endpoint) {}
	aliases := &data.UnsortedKV{}
	return NewLAN(iface, gateway, aliases, exNewCallback, exLostCallback)
}

func buildExampleEndpoint() *Endpoint {
	iface, _ := FindInterface("")
	return iface
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
	aliases := &data.UnsortedKV{}
	lan := NewLAN(iface, gateway, aliases, exNewCallback, exLostCallback)
	if lan.iface != iface {
		t.Fatalf("expected '%v', got '%v'", iface, lan.iface)
	}
	if lan.gateway != gateway {
		t.Fatalf("expected '%v', got '%v'", gateway, lan.gateway)
	}
	if len(lan.hosts) != 0 {
		t.Fatalf("expected '%v', got '%v'", 0, len(lan.hosts))
	}
	// FIXME: update this to current code base
	// if !(len(lan.aliases.data) >= 0) {
	// 	t.Fatalf("expected '%v', got '%v'", 0, len(lan.aliases.data))
	// }
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
	aliases := &data.UnsortedKV{}
	lan := NewLAN(iface, gateway, aliases, exNewCallback, exLostCallback)
	_, err = lan.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
}

// FIXME: update this to current code base
// func TestSetAliasFor(t *testing.T) {
// 	exampleAlias := "picat"
// 	exampleLAN := buildExampleLAN()
// 	exampleEndpoint := buildExampleEndpoint()
// 	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
// 	if !exampleLAN.SetAliasFor(exampleEndpoint.HwAddress, exampleAlias) {
// 		t.Error("unable to set alias for a given mac address")
// 	}
// }

func TestGet(t *testing.T) {
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	foundEndpoint, foundBool := exampleLAN.Get(exampleEndpoint.HwAddress)
	if foundEndpoint.String() != exampleEndpoint.String() {
		t.Fatalf("expected '%v', got '%v'", foundEndpoint, exampleEndpoint)
	}
	if !foundBool {
		t.Error("unable to get known endpoint via mac address from LAN struct")
	}
}

func TestList(t *testing.T) {
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	foundList := exampleLAN.List()
	if len(foundList) != 1 {
		t.Fatalf("expected '%d', got '%d'", 1, len(foundList))
	}
	exp := 1
	got := len(exampleLAN.List())
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

// FIXME: update this to current code base
// func TestAliases(t *testing.T) {
// 	exampleAlias := "picat"
// 	exampleLAN := buildExampleLAN()
// 	exampleEndpoint := buildExampleEndpoint()
// 	exampleLAN.hosts["pi:ca:tw:as:he:re"] = exampleEndpoint
// 	exp := exampleAlias
// 	got := exampleLAN.Aliases().Get("pi:ca:tw:as:he:re")
// 	if got != exp {
// 		t.Fatalf("expected '%v', got '%v'", exp, got)
// 	}
// }

func TestWasMissed(t *testing.T) {
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	exp := false
	got := exampleLAN.WasMissed(exampleEndpoint.HwAddress)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

// TODO Add TestRemove after removing unnecessary ip argument
// func TestRemove(t *testing.T) {
// }

func TestHas(t *testing.T) {
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	if !exampleLAN.Has(exampleEndpoint.IpAddress) {
		t.Error("unable find a known IP address in LAN struct")
	}
}

func TestEachHost(t *testing.T) {
	exampleBuffer := []string{}
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
	exampleCB := func(mac string, e *Endpoint) {
		exampleBuffer = append(exampleBuffer, exampleEndpoint.HwAddress)
	}
	exampleLAN.EachHost(exampleCB)
	exp := 1
	got := len(exampleBuffer)
	if got != exp {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func TestGetByIp(t *testing.T) {
	exampleLAN := buildExampleLAN()
	exampleEndpoint := buildExampleEndpoint()
	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint

	exp := exampleEndpoint
	got := exampleLAN.GetByIp(exampleEndpoint.IpAddress)
	if got.String() != exp.String() {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}

func TestAddIfNew(t *testing.T) {
	exampleLAN := buildExampleLAN()
	iface, _ := FindInterface("")
	// won't add our own IP address
	if exampleLAN.AddIfNew(iface.IpAddress, iface.HwAddress) != nil {
		t.Error("added address that should've been ignored ( your own )")
	}
}

// FIXME: update this to current code base
// func TestGetAlias(t *testing.T) {
// 	exampleAlias := "picat"
// 	exampleLAN := buildExampleLAN()
// 	exampleEndpoint := buildExampleEndpoint()
// 	exampleLAN.hosts[exampleEndpoint.HwAddress] = exampleEndpoint
// 	exp := exampleAlias
// 	got := exampleLAN.GetAlias(exampleEndpoint.HwAddress)
// 	if got != exp {
// 		t.Fatalf("expected '%v', got '%v'", exp, got)
// 	}
// }

func TestShouldIgnore(t *testing.T) {
	exampleLAN := buildExampleLAN()
	iface, _ := FindInterface("")
	gateway, _ := FindGateway(iface)
	exp := true
	got := exampleLAN.shouldIgnore(iface.IpAddress, iface.HwAddress)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
	got = exampleLAN.shouldIgnore(gateway.IpAddress, gateway.HwAddress)
	if got != exp {
		t.Fatalf("expected '%v', got '%v'", exp, got)
	}
}
