package net_probe

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/packets"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/malfunkt/iprange"
)

// MockQueue implements a mock packet queue for testing
type MockQueue struct {
	sync.Mutex
	sentPackets [][]byte
	sendError   error
	active      bool
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		sentPackets: make([][]byte, 0),
		active:      true,
	}
}

func (m *MockQueue) Send(data []byte) error {
	m.Lock()
	defer m.Unlock()

	if m.sendError != nil {
		return m.sendError
	}

	// Store a copy of the packet
	packet := make([]byte, len(data))
	copy(packet, data)
	m.sentPackets = append(m.sentPackets, packet)
	return nil
}

func (m *MockQueue) GetSentPackets() [][]byte {
	m.Lock()
	defer m.Unlock()
	return m.sentPackets
}

func (m *MockQueue) ClearSentPackets() {
	m.Lock()
	defer m.Unlock()
	m.sentPackets = make([][]byte, 0)
}

func (m *MockQueue) Stop() {
	m.Lock()
	defer m.Unlock()
	m.active = false
}

// MockSession for testing
type MockSession struct {
	*session.Session
	runCommands []string
	skipIPs     map[string]bool
}

func (m *MockSession) Run(cmd string) error {
	m.runCommands = append(m.runCommands, cmd)

	// Handle module commands
	if cmd == "net.recon on" {
		// Find and start the net.recon module
		for _, mod := range m.Modules {
			if mod.Name() == "net.recon" {
				if !mod.Running() {
					return mod.Start()
				}
				return nil
			}
		}
	} else if cmd == "net.recon off" {
		// Find and stop the net.recon module
		for _, mod := range m.Modules {
			if mod.Name() == "net.recon" {
				if mod.Running() {
					return mod.Stop()
				}
				return nil
			}
		}
	} else if cmd == "zerogod.discovery on" || cmd == "zerogod.discovery off" {
		// Mock zerogod.discovery commands
		return nil
	}

	return nil
}

func (m *MockSession) Skip(ip net.IP) bool {
	if m.skipIPs == nil {
		return false
	}
	return m.skipIPs[ip.String()]
}

// MockNetRecon implements a minimal net.recon module for testing
type MockNetRecon struct {
	session.SessionModule
}

func NewMockNetRecon(s *session.Session) *MockNetRecon {
	mod := &MockNetRecon{
		SessionModule: session.NewSessionModule("net.recon", s),
	}

	// Add handlers so the module can be started/stopped via commands
	mod.AddHandler(session.NewModuleHandler("net.recon on", "",
		"Start net.recon",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("net.recon off", "",
		"Stop net.recon",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (m *MockNetRecon) Name() string {
	return "net.recon"
}

func (m *MockNetRecon) Description() string {
	return "Mock net.recon module"
}

func (m *MockNetRecon) Author() string {
	return "test"
}

func (m *MockNetRecon) Configure() error {
	return nil
}

func (m *MockNetRecon) Start() error {
	return m.SetRunning(true, nil)
}

func (m *MockNetRecon) Stop() error {
	return m.SetRunning(false, nil)
}

// Create a mock session for testing
func createMockSession() (*MockSession, *MockQueue) {
	// Create interface
	iface := &network.Endpoint{
		IpAddress: "192.168.1.100",
		HwAddress: "aa:bb:cc:dd:ee:ff",
		Hostname:  "eth0",
	}
	iface.SetIP("192.168.1.100")
	iface.SetBits(24)

	// Parse interface addresses
	ifaceIP := net.ParseIP("192.168.1.100")
	ifaceHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	iface.IP = ifaceIP
	iface.HW = ifaceHW

	// Create gateway
	gateway := &network.Endpoint{
		IpAddress: "192.168.1.1",
		HwAddress: "11:22:33:44:55:66",
	}

	// Create mock queue
	mockQueue := NewMockQueue()

	// Create environment
	env, _ := session.NewEnvironment("")

	// Create session
	sess := &session.Session{
		Interface: iface,
		Gateway:   gateway,
		StartedAt: time.Now(),
		Active:    true,
		Env:       env,
		Queue: &packets.Queue{
			Traffic: sync.Map{},
			Stats:   packets.Stats{},
		},
		Modules: make(session.ModuleList, 0),
	}

	// Initialize events
	sess.Events = session.NewEventPool(false, false)

	// Add mock net.recon module
	mockNetRecon := NewMockNetRecon(sess)
	sess.Modules = append(sess.Modules, mockNetRecon)

	// Create mock session wrapper
	mockSess := &MockSession{
		Session:     sess,
		runCommands: make([]string, 0),
		skipIPs:     make(map[string]bool),
	}

	return mockSess, mockQueue
}

func TestNewProber(t *testing.T) {
	mockSess, _ := createMockSession()

	mod := NewProber(mockSess.Session)

	if mod == nil {
		t.Fatal("NewProber returned nil")
	}

	if mod.Name() != "net.probe" {
		t.Errorf("expected module name 'net.probe', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("unexpected author: %s", mod.Author())
	}

	// Check parameters
	params := []string{"net.probe.nbns", "net.probe.mdns", "net.probe.upnp", "net.probe.wsd", "net.probe.throttle"}
	for _, param := range params {
		if !mod.Session.Env.Has(param) {
			t.Errorf("parameter %s not registered", param)
		}
	}
}

func TestProberConfigure(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]string
		expectErr bool
		expected  struct {
			throttle int
			nbns     bool
			mdns     bool
			upnp     bool
			wsd      bool
		}
	}{
		{
			name: "default configuration",
			params: map[string]string{
				"net.probe.throttle": "10",
				"net.probe.nbns":     "true",
				"net.probe.mdns":     "true",
				"net.probe.upnp":     "true",
				"net.probe.wsd":      "true",
			},
			expectErr: false,
			expected: struct {
				throttle int
				nbns     bool
				mdns     bool
				upnp     bool
				wsd      bool
			}{10, true, true, true, true},
		},
		{
			name: "disabled probes",
			params: map[string]string{
				"net.probe.throttle": "5",
				"net.probe.nbns":     "false",
				"net.probe.mdns":     "false",
				"net.probe.upnp":     "false",
				"net.probe.wsd":      "false",
			},
			expectErr: false,
			expected: struct {
				throttle int
				nbns     bool
				mdns     bool
				upnp     bool
				wsd      bool
			}{5, false, false, false, false},
		},
		{
			name: "invalid throttle",
			params: map[string]string{
				"net.probe.throttle": "invalid",
				"net.probe.nbns":     "true",
				"net.probe.mdns":     "true",
				"net.probe.upnp":     "true",
				"net.probe.wsd":      "true",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSess, _ := createMockSession()
			mod := NewProber(mockSess.Session)

			// Set parameters
			for k, v := range tt.params {
				mockSess.Env.Set(k, v)
			}

			err := mod.Configure()

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectErr {
				if mod.throttle != tt.expected.throttle {
					t.Errorf("expected throttle %d, got %d", tt.expected.throttle, mod.throttle)
				}
				if mod.probes.NBNS != tt.expected.nbns {
					t.Errorf("expected NBNS %v, got %v", tt.expected.nbns, mod.probes.NBNS)
				}
				if mod.probes.MDNS != tt.expected.mdns {
					t.Errorf("expected MDNS %v, got %v", tt.expected.mdns, mod.probes.MDNS)
				}
				if mod.probes.UPNP != tt.expected.upnp {
					t.Errorf("expected UPNP %v, got %v", tt.expected.upnp, mod.probes.UPNP)
				}
				if mod.probes.WSD != tt.expected.wsd {
					t.Errorf("expected WSD %v, got %v", tt.expected.wsd, mod.probes.WSD)
				}
			}
		})
	}
}

// MockProber wraps Prober to allow mocking probe methods
type MockProber struct {
	*Prober
	nbnsCount *int32
	upnpCount *int32
	wsdCount  *int32
	mockQueue *MockQueue
}

func (m *MockProber) sendProbeNBNS(from net.IP, from_hw net.HardwareAddr, to net.IP) {
	atomic.AddInt32(m.nbnsCount, 1)
	m.mockQueue.Send([]byte(fmt.Sprintf("NBNS probe to %s", to)))
}

func (m *MockProber) sendProbeUPNP(from net.IP, from_hw net.HardwareAddr) {
	atomic.AddInt32(m.upnpCount, 1)
	m.mockQueue.Send([]byte("UPNP probe"))
}

func (m *MockProber) sendProbeWSD(from net.IP, from_hw net.HardwareAddr) {
	atomic.AddInt32(m.wsdCount, 1)
	m.mockQueue.Send([]byte("WSD probe"))
}

func TestProberStartStop(t *testing.T) {
	mockSess, _ := createMockSession()
	mod := NewProber(mockSess.Session)

	// Configure with fast throttle for testing
	mockSess.Env.Set("net.probe.throttle", "1")
	mockSess.Env.Set("net.probe.nbns", "true")
	mockSess.Env.Set("net.probe.mdns", "true")
	mockSess.Env.Set("net.probe.upnp", "true")
	mockSess.Env.Set("net.probe.wsd", "true")

	// Start the prober
	err := mod.Start()
	if err != nil {
		t.Fatalf("Failed to start prober: %v", err)
	}

	if !mod.Running() {
		t.Error("Prober should be running after Start()")
	}

	// Give it a moment to initialize
	time.Sleep(50 * time.Millisecond)

	// Stop the prober
	err = mod.Stop()
	if err != nil {
		t.Fatalf("Failed to stop prober: %v", err)
	}

	if mod.Running() {
		t.Error("Prober should not be running after Stop()")
	}

	// Since we can't easily mock the probe methods, we'll verify the module's state
	// and trust that the actual probe sending is tested in integration tests
}

func TestProberMonitorMode(t *testing.T) {
	mockSess, _ := createMockSession()
	mod := NewProber(mockSess.Session)

	// Set interface to monitor mode
	mockSess.Interface.IpAddress = network.MonitorModeAddress

	// Start the prober
	err := mod.Start()
	if err != nil {
		t.Fatalf("Failed to start prober: %v", err)
	}

	// Give it time to potentially start probing
	time.Sleep(50 * time.Millisecond)

	// Stop the prober
	mod.Stop()

	// In monitor mode, the prober should exit early without doing any work
	// We can't easily verify no probes were sent without mocking network calls,
	// but we can verify the module starts and stops correctly
}

func TestProberHandlers(t *testing.T) {
	mockSess, _ := createMockSession()
	mod := NewProber(mockSess.Session)

	// Test handlers
	handlers := mod.Handlers()

	expectedHandlers := []string{"net.probe on", "net.probe off"}
	handlerMap := make(map[string]bool)

	for _, h := range handlers {
		handlerMap[h.Name] = true
	}

	for _, expected := range expectedHandlers {
		if !handlerMap[expected] {
			t.Errorf("Expected handler '%s' not found", expected)
		}
	}

	// Test handler execution
	for _, h := range handlers {
		if h.Name == "net.probe on" {
			// Should start the module
			err := h.Exec([]string{})
			if err != nil {
				t.Errorf("Handler 'net.probe on' failed: %v", err)
			}
			if !mod.Running() {
				t.Error("Module should be running after 'net.probe on'")
			}
			mod.Stop()
		} else if h.Name == "net.probe off" {
			// Start first, then stop
			mod.Start()
			err := h.Exec([]string{})
			if err != nil {
				t.Errorf("Handler 'net.probe off' failed: %v", err)
			}
			if mod.Running() {
				t.Error("Module should not be running after 'net.probe off'")
			}
		}
	}
}

func TestProberSelectiveProbes(t *testing.T) {
	tests := []struct {
		name          string
		enabledProbes map[string]bool
	}{
		{
			name: "only NBNS",
			enabledProbes: map[string]bool{
				"nbns": true,
				"mdns": false,
				"upnp": false,
				"wsd":  false,
			},
		},
		{
			name: "only UPNP and WSD",
			enabledProbes: map[string]bool{
				"nbns": false,
				"mdns": false,
				"upnp": true,
				"wsd":  true,
			},
		},
		{
			name: "all probes enabled",
			enabledProbes: map[string]bool{
				"nbns": true,
				"mdns": true,
				"upnp": true,
				"wsd":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSess, _ := createMockSession()
			mod := NewProber(mockSess.Session)

			// Configure probes
			mockSess.Env.Set("net.probe.throttle", "10")
			mockSess.Env.Set("net.probe.nbns", fmt.Sprintf("%v", tt.enabledProbes["nbns"]))
			mockSess.Env.Set("net.probe.mdns", fmt.Sprintf("%v", tt.enabledProbes["mdns"]))
			mockSess.Env.Set("net.probe.upnp", fmt.Sprintf("%v", tt.enabledProbes["upnp"]))
			mockSess.Env.Set("net.probe.wsd", fmt.Sprintf("%v", tt.enabledProbes["wsd"]))

			// Configure and verify the settings
			err := mod.Configure()
			if err != nil {
				t.Fatalf("Failed to configure: %v", err)
			}

			// Verify configuration
			if mod.probes.NBNS != tt.enabledProbes["nbns"] {
				t.Errorf("NBNS probe setting mismatch: expected %v, got %v",
					tt.enabledProbes["nbns"], mod.probes.NBNS)
			}
			if mod.probes.MDNS != tt.enabledProbes["mdns"] {
				t.Errorf("MDNS probe setting mismatch: expected %v, got %v",
					tt.enabledProbes["mdns"], mod.probes.MDNS)
			}
			if mod.probes.UPNP != tt.enabledProbes["upnp"] {
				t.Errorf("UPNP probe setting mismatch: expected %v, got %v",
					tt.enabledProbes["upnp"], mod.probes.UPNP)
			}
			if mod.probes.WSD != tt.enabledProbes["wsd"] {
				t.Errorf("WSD probe setting mismatch: expected %v, got %v",
					tt.enabledProbes["wsd"], mod.probes.WSD)
			}
		})
	}
}

func TestIPRangeExpansion(t *testing.T) {
	// Test that we correctly iterate through the subnet
	cidr := "192.168.1.0/30" // Small subnet for testing
	list, err := iprange.Parse(cidr)
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	addresses := list.Expand()

	// For /30, we should get 4 addresses
	expectedAddresses := []string{
		"192.168.1.0",
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}

	if len(addresses) != len(expectedAddresses) {
		t.Errorf("Expected %d addresses, got %d", len(expectedAddresses), len(addresses))
	}

	for i, addr := range addresses {
		if addr.String() != expectedAddresses[i] {
			t.Errorf("Expected address %s at position %d, got %s", expectedAddresses[i], i, addr.String())
		}
	}
}

// Benchmarks
func BenchmarkProberConfiguration(b *testing.B) {
	mockSess, _ := createMockSession()

	// Set up parameters
	mockSess.Env.Set("net.probe.throttle", "10")
	mockSess.Env.Set("net.probe.nbns", "true")
	mockSess.Env.Set("net.probe.mdns", "true")
	mockSess.Env.Set("net.probe.upnp", "true")
	mockSess.Env.Set("net.probe.wsd", "true")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mod := NewProber(mockSess.Session)
		mod.Configure()
	}
}

func BenchmarkIPRangeExpansion(b *testing.B) {
	cidr := "192.168.1.0/24"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list, _ := iprange.Parse(cidr)
		_ = list.Expand()
	}
}
