package routing

import (
	"testing"
)

func TestRouteType(t *testing.T) {
	// Test the RouteType constants
	if IPv4 != RouteType("IPv4") {
		t.Errorf("IPv4 constant has wrong value: %s", IPv4)
	}
	if IPv6 != RouteType("IPv6") {
		t.Errorf("IPv6 constant has wrong value: %s", IPv6)
	}
}

func TestRouteStruct(t *testing.T) {
	tests := []struct {
		name  string
		route Route
	}{
		{
			name: "IPv4 default route",
			route: Route{
				Type:        IPv4,
				Default:     true,
				Device:      "eth0",
				Destination: "0.0.0.0",
				Gateway:     "192.168.1.1",
				Flags:       "UG",
			},
		},
		{
			name: "IPv4 network route",
			route: Route{
				Type:        IPv4,
				Default:     false,
				Device:      "eth0",
				Destination: "192.168.1.0/24",
				Gateway:     "",
				Flags:       "U",
			},
		},
		{
			name: "IPv6 default route",
			route: Route{
				Type:        IPv6,
				Default:     true,
				Device:      "eth0",
				Destination: "::/0",
				Gateway:     "fe80::1",
				Flags:       "UG",
			},
		},
		{
			name: "IPv6 link-local route",
			route: Route{
				Type:        IPv6,
				Default:     false,
				Device:      "eth0",
				Destination: "fe80::/64",
				Gateway:     "",
				Flags:       "U",
			},
		},
		{
			name: "localhost route",
			route: Route{
				Type:        IPv4,
				Default:     false,
				Device:      "lo",
				Destination: "127.0.0.0/8",
				Gateway:     "",
				Flags:       "U",
			},
		},
		{
			name: "VPN route",
			route: Route{
				Type:        IPv4,
				Default:     false,
				Device:      "tun0",
				Destination: "10.8.0.0/24",
				Gateway:     "",
				Flags:       "U",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that all fields are accessible
			_ = tt.route.Type
			_ = tt.route.Default
			_ = tt.route.Device
			_ = tt.route.Destination
			_ = tt.route.Gateway
			_ = tt.route.Flags

			// Verify the route has the expected type
			if tt.route.Type != IPv4 && tt.route.Type != IPv6 {
				t.Errorf("route has invalid type: %s", tt.route.Type)
			}
		})
	}
}

func TestRouteDefaultFlag(t *testing.T) {
	// Test routes with different default flag settings
	defaultRoute := Route{
		Type:        IPv4,
		Default:     true,
		Device:      "eth0",
		Destination: "0.0.0.0",
		Gateway:     "192.168.1.1",
		Flags:       "UG",
	}

	normalRoute := Route{
		Type:        IPv4,
		Default:     false,
		Device:      "eth0",
		Destination: "192.168.1.0/24",
		Gateway:     "",
		Flags:       "U",
	}

	if !defaultRoute.Default {
		t.Error("default route should have Default=true")
	}

	if normalRoute.Default {
		t.Error("normal route should have Default=false")
	}
}

func TestRouteTypeString(t *testing.T) {
	// Test that RouteType can be converted to string
	ipv4Str := string(IPv4)
	ipv6Str := string(IPv6)

	if ipv4Str != "IPv4" {
		t.Errorf("IPv4 string conversion failed: got %s", ipv4Str)
	}

	if ipv6Str != "IPv6" {
		t.Errorf("IPv6 string conversion failed: got %s", ipv6Str)
	}
}

func TestRouteTypeComparison(t *testing.T) {
	// Test RouteType comparisons
	var rt1 RouteType = IPv4
	var rt2 RouteType = IPv4
	var rt3 RouteType = IPv6

	if rt1 != rt2 {
		t.Error("identical RouteType values should be equal")
	}

	if rt1 == rt3 {
		t.Error("different RouteType values should not be equal")
	}
}

func TestRouteTypeCustomValues(t *testing.T) {
	// Test that custom RouteType values can be created
	customType := RouteType("Custom")

	if customType == IPv4 || customType == IPv6 {
		t.Error("custom RouteType should not equal predefined constants")
	}

	if string(customType) != "Custom" {
		t.Errorf("custom RouteType string conversion failed: got %s", customType)
	}
}

func TestRouteWithEmptyFields(t *testing.T) {
	// Test route with empty fields
	emptyRoute := Route{}

	if emptyRoute.Type != "" {
		t.Errorf("empty route Type should be empty string, got %s", emptyRoute.Type)
	}

	if emptyRoute.Default != false {
		t.Error("empty route Default should be false")
	}

	if emptyRoute.Device != "" {
		t.Errorf("empty route Device should be empty string, got %s", emptyRoute.Device)
	}

	if emptyRoute.Destination != "" {
		t.Errorf("empty route Destination should be empty string, got %s", emptyRoute.Destination)
	}

	if emptyRoute.Gateway != "" {
		t.Errorf("empty route Gateway should be empty string, got %s", emptyRoute.Gateway)
	}

	if emptyRoute.Flags != "" {
		t.Errorf("empty route Flags should be empty string, got %s", emptyRoute.Flags)
	}
}

func TestRouteFieldAssignment(t *testing.T) {
	// Test that route fields can be assigned individually
	r := Route{}

	r.Type = IPv6
	r.Default = true
	r.Device = "wlan0"
	r.Destination = "2001:db8::/32"
	r.Gateway = "fe80::1"
	r.Flags = "UGH"

	if r.Type != IPv6 {
		t.Errorf("Type assignment failed: got %s", r.Type)
	}

	if !r.Default {
		t.Error("Default assignment failed")
	}

	if r.Device != "wlan0" {
		t.Errorf("Device assignment failed: got %s", r.Device)
	}

	if r.Destination != "2001:db8::/32" {
		t.Errorf("Destination assignment failed: got %s", r.Destination)
	}

	if r.Gateway != "fe80::1" {
		t.Errorf("Gateway assignment failed: got %s", r.Gateway)
	}

	if r.Flags != "UGH" {
		t.Errorf("Flags assignment failed: got %s", r.Flags)
	}
}

func TestRouteArrayOperations(t *testing.T) {
	// Test operations on arrays of routes
	routes := []Route{
		{
			Type:        IPv4,
			Default:     true,
			Device:      "eth0",
			Destination: "0.0.0.0",
			Gateway:     "192.168.1.1",
			Flags:       "UG",
		},
		{
			Type:        IPv4,
			Default:     false,
			Device:      "eth0",
			Destination: "192.168.1.0/24",
			Gateway:     "",
			Flags:       "U",
		},
		{
			Type:        IPv6,
			Default:     false,
			Device:      "eth0",
			Destination: "fe80::/64",
			Gateway:     "",
			Flags:       "U",
		},
	}

	// Test array length
	if len(routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(routes))
	}

	// Count IPv4 vs IPv6 routes
	ipv4Count := 0
	ipv6Count := 0
	defaultCount := 0

	for _, r := range routes {
		switch r.Type {
		case IPv4:
			ipv4Count++
		case IPv6:
			ipv6Count++
		}

		if r.Default {
			defaultCount++
		}
	}

	if ipv4Count != 2 {
		t.Errorf("expected 2 IPv4 routes, got %d", ipv4Count)
	}

	if ipv6Count != 1 {
		t.Errorf("expected 1 IPv6 route, got %d", ipv6Count)
	}

	if defaultCount != 1 {
		t.Errorf("expected 1 default route, got %d", defaultCount)
	}
}

func BenchmarkRouteCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Route{
			Type:        IPv4,
			Default:     true,
			Device:      "eth0",
			Destination: "0.0.0.0",
			Gateway:     "192.168.1.1",
			Flags:       "UG",
		}
	}
}

func BenchmarkRouteTypeComparison(b *testing.B) {
	rt1 := IPv4
	rt2 := IPv6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rt1 == rt2
	}
}

func BenchmarkRouteArrayIteration(b *testing.B) {
	routes := make([]Route, 100)
	for i := range routes {
		if i%2 == 0 {
			routes[i].Type = IPv4
		} else {
			routes[i].Type = IPv6
		}
		routes[i].Device = "eth0"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for _, r := range routes {
			if r.Type == IPv4 {
				count++
			}
		}
		_ = count
	}
}
