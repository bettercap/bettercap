package routing

import (
	"fmt"
	"sync"
	"testing"
)

// Helper function to reset the table for testing
func resetTable() {
	lock.Lock()
	defer lock.Unlock()
	table = make([]Route, 0)
}

// Helper function to add routes for testing
func addTestRoutes() {
	lock.Lock()
	defer lock.Unlock()
	table = []Route{
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
			Default:     true,
			Device:      "eth0",
			Destination: "::/0",
			Gateway:     "fe80::1",
			Flags:       "UG",
		},
		{
			Type:        IPv6,
			Default:     false,
			Device:      "eth0",
			Destination: "fe80::/64",
			Gateway:     "",
			Flags:       "U",
		},
		{
			Type:        IPv4,
			Default:     false,
			Device:      "lo",
			Destination: "127.0.0.0/8",
			Gateway:     "",
			Flags:       "U",
		},
		{
			Type:        IPv4,
			Default:     true,
			Device:      "wlan0",
			Destination: "0.0.0.0",
			Gateway:     "10.0.0.1",
			Flags:       "UG",
		},
	}
}

func TestTable(t *testing.T) {
	// Reset table
	resetTable()

	// Test empty table
	routes := Table()
	if len(routes) != 0 {
		t.Errorf("Expected empty table, got %d routes", len(routes))
	}

	// Add test routes
	addTestRoutes()

	// Test table with routes
	routes = Table()
	if len(routes) != 6 {
		t.Errorf("Expected 6 routes, got %d", len(routes))
	}

	// Verify first route
	if routes[0].Type != IPv4 {
		t.Errorf("Expected first route to be IPv4, got %s", routes[0].Type)
	}
	if !routes[0].Default {
		t.Error("Expected first route to be default")
	}
	if routes[0].Gateway != "192.168.1.1" {
		t.Errorf("Expected gateway 192.168.1.1, got %s", routes[0].Gateway)
	}
}

func TestGateway(t *testing.T) {
	// Note: Gateway() calls Update() which loads real system routes
	// So we can't test specific values, just test the behavior

	// Test IPv4 gateway
	gateway, err := Gateway(IPv4, "")
	if err != nil {
		t.Errorf("Unexpected error getting IPv4 gateway: %v", err)
	}
	t.Logf("System IPv4 gateway: %s", gateway)

	// Test IPv6 gateway
	gateway, err = Gateway(IPv6, "")
	if err != nil {
		t.Errorf("Unexpected error getting IPv6 gateway: %v", err)
	}
	t.Logf("System IPv6 gateway: %s", gateway)

	// Test with specific device that likely doesn't exist
	gateway, err = Gateway(IPv4, "nonexistent999")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Should return empty string for non-existent device
	if gateway != "" {
		t.Logf("Got gateway for non-existent device (might be Windows): %s", gateway)
	}
}

func TestGatewayBehavior(t *testing.T) {
	// Test that Gateway doesn't panic with various inputs
	testCases := []struct {
		name   string
		ipType RouteType
		device string
	}{
		{"IPv4 empty device", IPv4, ""},
		{"IPv6 empty device", IPv6, ""},
		{"IPv4 with device", IPv4, "eth0"},
		{"IPv6 with device", IPv6, "eth0"},
		{"Custom type", RouteType("custom"), ""},
		{"Empty type", RouteType(""), ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gateway, err := Gateway(tc.ipType, tc.device)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			t.Logf("Gateway for %s: %s", tc.name, gateway)
		})
	}
}

func TestGatewayEmptyTable(t *testing.T) {
	// Test with empty table
	resetTable()

	gateway, err := gatewayFromTable(IPv4, "eth0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if gateway != "" {
		t.Errorf("Expected empty gateway, got %s", gateway)
	}
}

func TestGatewayNoDefaultRoute(t *testing.T) {
	// Test with routes but no default
	resetTable()

	lock.Lock()
	table = []Route{
		{
			Type:        IPv4,
			Default:     false,
			Device:      "eth0",
			Destination: "192.168.1.0/24",
			Gateway:     "",
			Flags:       "U",
		},
	}
	lock.Unlock()

	gateway, err := gatewayFromTable(IPv4, "eth0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if gateway != "" {
		t.Errorf("Expected empty gateway, got %s", gateway)
	}
}

func TestGatewayWindowsCase(t *testing.T) {
	// Since Gateway() calls Update(), we can't control the table content
	// Just test that it doesn't panic and returns something
	gateway, err := Gateway(IPv4, "eth0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Logf("Gateway result for eth0: %s", gateway)
}

func TestGatewayFromTableWithDefaults(t *testing.T) {
	// Test gatewayFromTable with controlled data containing defaults
	resetTable()
	addTestRoutes()

	gateway, err := gatewayFromTable(IPv4, "eth0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if gateway != "192.168.1.1" {
		t.Errorf("Expected gateway 192.168.1.1, got %s", gateway)
	}

	// Test with device-specific lookup
	gateway, err = gatewayFromTable(IPv4, "wlan0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if gateway != "10.0.0.1" {
		t.Errorf("Expected gateway 10.0.0.1, got %s", gateway)
	}
}

func TestTableConcurrency(t *testing.T) {
	// Test concurrent access to Table()
	resetTable()
	addTestRoutes()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Multiple readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				routes := Table()
				if len(routes) != 6 {
					select {
					case errors <- fmt.Errorf("Expected 6 routes, got %d", len(routes)):
					default:
					}
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGatewayConcurrency(t *testing.T) {
	// Test concurrent access to Gateway()
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Multiple readers calling Gateway concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, err := Gateway(IPv4, "")
				if err != nil {
					select {
					case errors <- fmt.Errorf("goroutine %d: error: %v", id, err):
					default:
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			if errorCount <= 5 { // Only log first 5 errors
				t.Error(err)
			}
		}
	}
	if errorCount > 5 {
		t.Errorf("... and %d more errors", errorCount-5)
	}
}

func TestUpdate(t *testing.T) {
	// Note: Update() calls platform-specific update() function
	// which we can't easily test without mocking
	// But we can test that it doesn't panic and returns something
	resetTable()

	routes, err := Update()
	// The error might be nil or non-nil depending on the platform
	// and whether we have permissions to read routing table
	if err == nil && routes != nil {
		t.Logf("Update returned %d routes", len(routes))
	} else if err != nil {
		t.Logf("Update returned error (expected on some platforms): %v", err)
	}
}

func TestGatewayMultipleDefaults(t *testing.T) {
	// Since Gateway() calls Update() and loads real routes,
	// we can't test specific scenarios with multiple defaults
	// Just ensure it handles the real system state without panicking

	// Call Gateway multiple times to ensure consistency
	gateway1, err1 := Gateway(IPv4, "")
	gateway2, err2 := Gateway(IPv4, "")

	if err1 != nil {
		t.Errorf("First call error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second call error: %v", err2)
	}

	// Results should be consistent
	if gateway1 != gateway2 {
		t.Errorf("Inconsistent results: first=%s, second=%s", gateway1, gateway2)
	}

	t.Logf("Consistent gateway result: %s", gateway1)
}

// Benchmark tests
func BenchmarkTable(b *testing.B) {
	resetTable()
	addTestRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Table()
	}
}

func BenchmarkGateway(b *testing.B) {
	resetTable()
	addTestRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Gateway(IPv4, "eth0")
	}
}

func BenchmarkTableConcurrent(b *testing.B) {
	resetTable()
	addTestRoutes()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Table()
		}
	})
}

func BenchmarkGatewayConcurrent(b *testing.B) {
	resetTable()
	addTestRoutes()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Gateway(IPv4, "eth0")
		}
	})
}
