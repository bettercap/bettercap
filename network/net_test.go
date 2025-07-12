package network

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/evilsocket/islazy/data"
)

func TestIsZeroMac(t *testing.T) {
	tests := []struct {
		name     string
		mac      string
		expected bool
	}{
		{"zero mac", "00:00:00:00:00:00", true},
		{"non-zero mac", "00:00:00:00:00:01", false},
		{"broadcast mac", "ff:ff:ff:ff:ff:ff", false},
		{"random mac", "aa:bb:cc:dd:ee:ff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mac, _ := net.ParseMAC(tt.mac)
			if got := IsZeroMac(mac); got != tt.expected {
				t.Errorf("IsZeroMac() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsBroadcastMac(t *testing.T) {
	tests := []struct {
		name     string
		mac      string
		expected bool
	}{
		{"broadcast mac", "ff:ff:ff:ff:ff:ff", true},
		{"zero mac", "00:00:00:00:00:00", false},
		{"partial broadcast", "ff:ff:ff:ff:ff:00", false},
		{"random mac", "aa:bb:cc:dd:ee:ff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mac, _ := net.ParseMAC(tt.mac)
			if got := IsBroadcastMac(mac); got != tt.expected {
				t.Errorf("IsBroadcastMac() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNormalizeMac(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase with colons", "AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff"},
		{"uppercase with dashes", "AA-BB-CC-DD-EE-FF", "aa:bb:cc:dd:ee:ff"},
		{"lowercase with colons", "aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff"},
		{"mixed case with dashes", "aA-bB-cC-dD-eE-fF", "aa:bb:cc:dd:ee:ff"},
		{"short segments", "a:b:c:d:e:f", "0a:0b:0c:0d:0e:0f"},
		{"mixed short and full", "aa:b:cc:d:ee:f", "aa:0b:cc:0d:ee:0f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeMac(tt.input); got != tt.expected {
				t.Errorf("NormalizeMac(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseMACs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		expectError bool
	}{
		{
			name:     "single MAC",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: []string{"aa:bb:cc:dd:ee:ff"},
		},
		{
			name:     "multiple MACs comma separated",
			input:    "aa:bb:cc:dd:ee:ff, 11:22:33:44:55:66",
			expected: []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"},
		},
		{
			name:     "MACs with dashes",
			input:    "AA-BB-CC-DD-EE-FF",
			expected: []string{"aa:bb:cc:dd:ee:ff"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "mixed formats",
			input:    "aa:bb:cc:dd:ee:ff, AA-BB-CC-DD-EE-00",
			expected: []string{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macs, err := ParseMACs(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("ParseMACs() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if len(macs) != len(tt.expected) {
				t.Errorf("ParseMACs() returned %d MACs, want %d", len(macs), len(tt.expected))
				return
			}
			for i, mac := range macs {
				if mac.String() != tt.expected[i] {
					t.Errorf("ParseMACs()[%d] = %v, want %v", i, mac.String(), tt.expected[i])
				}
			}
		})
	}
}

func TestParseTargets(t *testing.T) {
	aliasMap, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}

	aliasMap.Set("aa:bb:cc:dd:ee:ff", "test_alias")
	aliasMap.Set("11:22:33:44:55:66", "home_laptop")

	cases := []struct {
		name             string
		inputTargets     string
		inputAliases     *data.UnsortedKV
		expectedIPCount  int
		expectedMACCount int
		expectError      bool
	}{
		{
			name:             "empty target string",
			inputTargets:     "",
			inputAliases:     &data.UnsortedKV{},
			expectedIPCount:  0,
			expectedMACCount: 0,
			expectError:      false,
		},
		{
			name:             "MACs and IPs",
			inputTargets:     "192.168.1.2, 192.168.1.3, aa:bb:cc:dd:ee:ff, 11:22:33:44:55:66",
			inputAliases:     &data.UnsortedKV{},
			expectedIPCount:  2,
			expectedMACCount: 2,
			expectError:      false,
		},
		{
			name:             "aliases",
			inputTargets:     "test_alias, home_laptop",
			inputAliases:     aliasMap,
			expectedIPCount:  0,
			expectedMACCount: 2,
			expectError:      false,
		},
		{
			name:             "mixed aliases and MACs",
			inputTargets:     "test_alias, 99:88:77:66:55:44",
			inputAliases:     aliasMap,
			expectedIPCount:  0,
			expectedMACCount: 2,
			expectError:      false,
		},
		{
			name:             "IP range",
			inputTargets:     "192.168.1.1-3",
			inputAliases:     &data.UnsortedKV{},
			expectedIPCount:  3,
			expectedMACCount: 0,
			expectError:      false,
		},
		{
			name:             "CIDR notation",
			inputTargets:     "192.168.1.0/30",
			inputAliases:     &data.UnsortedKV{},
			expectedIPCount:  4,
			expectedMACCount: 0,
			expectError:      false,
		},
		{
			name:             "unknown alias",
			inputTargets:     "unknown_alias",
			inputAliases:     aliasMap,
			expectedIPCount:  0,
			expectedMACCount: 0,
			expectError:      true,
		},
		{
			name:             "invalid IP",
			inputTargets:     "invalid.ip.address",
			inputAliases:     &data.UnsortedKV{},
			expectedIPCount:  0,
			expectedMACCount: 0,
			expectError:      true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ips, macs, err := ParseTargets(test.inputTargets, test.inputAliases)
			if (err != nil) != test.expectError {
				t.Errorf("ParseTargets() error = %v, expectError %v", err, test.expectError)
			}
			if test.expectError {
				return
			}
			if len(ips) != test.expectedIPCount {
				t.Errorf("Wrong number of IPs. Got %d, want %d", len(ips), test.expectedIPCount)
			}
			if len(macs) != test.expectedMACCount {
				t.Errorf("Wrong number of MACs. Got %d, want %d", len(macs), test.expectedMACCount)
			}
		})
	}
}

func TestParseEndpoints(t *testing.T) {
	// Create a mock LAN with some endpoints
	iface := NewEndpoint("192.168.1.100", "aa:bb:cc:dd:ee:ff")
	gateway := NewEndpoint("192.168.1.1", "11:22:33:44:55:66")
	aliases, _ := data.NewMemUnsortedKV()

	// Need to provide non-nil callbacks
	newCb := func(e *Endpoint) {}
	lostCb := func(e *Endpoint) {}
	lan := NewLAN(iface, gateway, aliases, newCb, lostCb)

	// Add test endpoints
	lan.AddIfNew("192.168.1.10", "10:20:30:40:50:60")
	lan.AddIfNew("192.168.1.20", "20:30:40:50:60:70")

	// Set up an alias
	aliases.Set("10:20:30:40:50:60", "test_device")

	tests := []struct {
		name          string
		targets       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "single IP",
			targets:       "192.168.1.10",
			expectedCount: 1,
		},
		{
			name:          "single MAC",
			targets:       "10:20:30:40:50:60",
			expectedCount: 1,
		},
		{
			name:          "alias",
			targets:       "test_device",
			expectedCount: 1,
		},
		{
			name:          "multiple targets",
			targets:       "192.168.1.10, 20:30:40:50:60:70",
			expectedCount: 2,
		},
		{
			name:          "unknown IP",
			targets:       "192.168.1.99",
			expectedCount: 0,
		},
		{
			name:        "invalid target",
			targets:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints, err := ParseEndpoints(tt.targets, lan)
			if (err != nil) != tt.expectError {
				t.Errorf("ParseEndpoints() error = %v, expectError %v", err, tt.expectError)
			}
			if !tt.expectError && len(endpoints) != tt.expectedCount {
				t.Errorf("ParseEndpoints() returned %d endpoints, want %d", len(endpoints), tt.expectedCount)
			}
		})
	}
}

func TestBuildEndpointFromInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skip("Unable to get network interfaces")
	}
	if len(ifaces) == 0 {
		t.Skip("No network interfaces available")
	}

	// Find a suitable interface for testing
	var testIface *net.Interface
	for _, iface := range ifaces {
		if iface.HardwareAddr != nil && len(iface.HardwareAddr) > 0 {
			testIface = &iface
			break
		}
	}

	if testIface == nil {
		t.Skip("No suitable network interface found for testing")
	}

	endpoint, err := buildEndpointFromInterface(*testIface)
	if err != nil {
		t.Fatalf("buildEndpointFromInterface() error = %v", err)
	}

	if endpoint == nil {
		t.Fatal("buildEndpointFromInterface() returned nil endpoint")
	}

	// Verify basic properties
	if endpoint.Index != testIface.Index {
		t.Errorf("endpoint.Index = %d, want %d", endpoint.Index, testIface.Index)
	}

	if endpoint.HwAddress != testIface.HardwareAddr.String() {
		t.Errorf("endpoint.HwAddress = %s, want %s", endpoint.HwAddress, testIface.HardwareAddr.String())
	}
}

func TestMatchByAddress(t *testing.T) {
	// Create a mock interface for testing
	mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	iface := net.Interface{
		Name:         "eth0",
		HardwareAddr: mac,
	}

	tests := []struct {
		name     string
		search   string
		expected bool
	}{
		{"exact MAC match", "aa:bb:cc:dd:ee:ff", true},
		{"MAC with different case", "AA:BB:CC:DD:EE:FF", true},
		{"MAC with dashes", "aa-bb-cc-dd-ee-ff", true},
		{"different MAC", "11:22:33:44:55:66", false},
		{"partial MAC", "aa:bb:cc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchByAddress(iface, tt.search); got != tt.expected {
				t.Errorf("matchByAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindInterfaceByName(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skip("Unable to get network interfaces")
	}
	if len(ifaces) == 0 {
		t.Skip("No network interfaces available")
	}

	// Test with first available interface
	testIface := ifaces[0]

	// Test finding by name
	endpoint, err := findInterfaceByName(testIface.Name, ifaces)
	if err != nil {
		t.Errorf("findInterfaceByName() error = %v", err)
	}
	if endpoint != nil && endpoint.Name() != testIface.Name {
		t.Errorf("findInterfaceByName() returned wrong interface")
	}

	// Test with non-existent interface
	_, err = findInterfaceByName("nonexistent999", ifaces)
	if err == nil {
		t.Error("findInterfaceByName() should return error for non-existent interface")
	}
}

func TestFindInterface(t *testing.T) {
	// Test with empty name (should return first suitable interface)
	endpoint, err := FindInterface("")
	if err != nil && err != ErrNoIfaces {
		t.Errorf("FindInterface() unexpected error = %v", err)
	}

	// Test with specific interface name
	ifaces, err := net.Interfaces()
	if err == nil && len(ifaces) > 0 {
		endpoint, err = FindInterface(ifaces[0].Name)
		if err != nil {
			t.Errorf("FindInterface() error = %v", err)
		}
		if endpoint != nil && endpoint.Name() != ifaces[0].Name {
			t.Errorf("FindInterface() returned wrong interface")
		}
	}

	// Test with non-existent interface
	_, err = FindInterface("nonexistent999")
	if err == nil {
		t.Error("FindInterface() should return error for non-existent interface")
	}
}

func TestColorRSSI(t *testing.T) {
	tests := []struct {
		name string
		rssi int
	}{
		{"excellent signal", -30},
		{"very good signal", -67},
		{"good signal", -70},
		{"fair signal", -80},
		{"poor signal", -90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorRSSI(tt.rssi)
			// Just ensure it returns a non-empty string
			if result == "" {
				t.Error("ColorRSSI() returned empty string")
			}
			// Check it contains the dBm value
			expected := fmt.Sprintf("%d dBm", tt.rssi)
			if !strings.Contains(result, expected) {
				t.Errorf("ColorRSSI() result doesn't contain expected value %s", expected)
			}
		})
	}
}

func TestSetWiFiRegion(t *testing.T) {
	// This test will likely fail without proper permissions
	// Just ensure the function doesn't panic
	err := SetWiFiRegion("US")
	// We don't check the error as it requires root/iw binary
	_ = err
}

func TestActivateInterface(t *testing.T) {
	// This test will likely fail without proper permissions
	// Just ensure the function doesn't panic
	err := ActivateInterface("nonexistent")
	// We expect an error for non-existent interface
	if err == nil {
		t.Error("ActivateInterface() should return error for non-existent interface")
	}
}

func TestSetInterfaceTxPower(t *testing.T) {
	// This test will likely fail without proper permissions
	// Just ensure the function doesn't panic
	err := SetInterfaceTxPower("nonexistent", 20)
	// We don't check the error as it requires root/iw binary
	_ = err
}

func TestGatewayProvidedByUser(t *testing.T) {
	iface := NewEndpoint("192.168.1.100", "aa:bb:cc:dd:ee:ff")

	tests := []struct {
		name        string
		gateway     string
		expectError bool
	}{
		{
			name:        "valid IPv4",
			gateway:     "192.168.1.1",
			expectError: false, // Will error without actual ARP
		},
		{
			name:        "invalid IPv4",
			gateway:     "999.999.999.999",
			expectError: true,
		},
		{
			name:        "not an IP",
			gateway:     "not-an-ip",
			expectError: true,
		},
		{
			name:        "IPv6",
			gateway:     "fe80::1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GatewayProvidedByUser(iface, tt.gateway)
			// We always expect an error in tests as we can't do actual ARP lookup
			if err == nil {
				t.Error("GatewayProvidedByUser() expected error in test environment")
			}
		})
	}
}

// Benchmarks
func BenchmarkNormalizeMac(b *testing.B) {
	macs := []string{
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"a:b:c:d:e:f",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeMac(macs[i%len(macs)])
	}
}

func BenchmarkParseMACs(b *testing.B) {
	input := "aa:bb:cc:dd:ee:ff, 11:22:33:44:55:66, AA-BB-CC-DD-EE-FF"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMACs(input)
	}
}

func BenchmarkParseTargets(b *testing.B) {
	aliases, _ := data.NewMemUnsortedKV()
	aliases.Set("aa:bb:cc:dd:ee:ff", "test_alias")

	targets := "192.168.1.1-10, aa:bb:cc:dd:ee:ff, test_alias"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseTargets(targets, aliases)
	}
}
