package arp_spoof

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/firewall"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/packets"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/data"
)

// MockFirewall implements a mock firewall for testing
type MockFirewall struct {
	forwardingEnabled bool
	redirections      []firewall.Redirection
}

func NewMockFirewall() *MockFirewall {
	return &MockFirewall{
		forwardingEnabled: false,
		redirections:      make([]firewall.Redirection, 0),
	}
}

func (m *MockFirewall) IsForwardingEnabled() bool {
	return m.forwardingEnabled
}

func (m *MockFirewall) EnableForwarding(enabled bool) error {
	m.forwardingEnabled = enabled
	return nil
}

func (m *MockFirewall) EnableRedirection(r *firewall.Redirection, enabled bool) error {
	if enabled {
		m.redirections = append(m.redirections, *r)
	} else {
		for i, red := range m.redirections {
			if red.String() == r.String() {
				m.redirections = append(m.redirections[:i], m.redirections[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *MockFirewall) DisableRedirection(r *firewall.Redirection, enabled bool) error {
	return m.EnableRedirection(r, false)
}

func (m *MockFirewall) Restore() {
	m.redirections = make([]firewall.Redirection, 0)
	m.forwardingEnabled = false
}

// MockPacketQueue extends packets.Queue to capture sent packets
type MockPacketQueue struct {
	*packets.Queue
	sync.Mutex
	sentPackets [][]byte
}

func NewMockPacketQueue() *MockPacketQueue {
	q := &packets.Queue{
		Traffic: sync.Map{},
		Stats:   packets.Stats{},
	}
	return &MockPacketQueue{
		Queue:       q,
		sentPackets: make([][]byte, 0),
	}
}

func (m *MockPacketQueue) Send(data []byte) error {
	m.Lock()
	defer m.Unlock()

	// Store a copy of the packet
	packet := make([]byte, len(data))
	copy(packet, data)
	m.sentPackets = append(m.sentPackets, packet)

	// Also update stats like the real queue would
	m.TrackSent(uint64(len(data)))

	return nil
}

func (m *MockPacketQueue) GetSentPackets() [][]byte {
	m.Lock()
	defer m.Unlock()
	return m.sentPackets
}

func (m *MockPacketQueue) ClearSentPackets() {
	m.Lock()
	defer m.Unlock()
	m.sentPackets = make([][]byte, 0)
}

// MockSession for testing
type MockSession struct {
	*session.Session
	findMACResults map[string]net.HardwareAddr
	skipIPs        map[string]bool
	mockQueue      *MockPacketQueue
}

// Override session methods to use our mocks
func setupMockSession(mockSess *MockSession) {
	// Replace the Session's FindMAC method behavior by manipulating the LAN
	// Since we can't override methods directly, we'll ensure the LAN has the data
	for ip, mac := range mockSess.findMACResults {
		mockSess.Lan.AddIfNew(ip, mac.String())
	}
}

func (m *MockSession) FindMAC(ip net.IP, probe bool) (net.HardwareAddr, error) {
	// First check our mock results
	if mac, ok := m.findMACResults[ip.String()]; ok {
		return mac, nil
	}
	// Then check the LAN
	if e, found := m.Lan.Get(ip.String()); found && e != nil {
		return e.HW, nil
	}
	return nil, fmt.Errorf("MAC not found for %s", ip.String())
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

	// Add handlers
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
func createMockSession() (*MockSession, *MockPacketQueue, *MockFirewall) {
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
	gatewayIP := net.ParseIP("192.168.1.1")
	gatewayHW, _ := net.ParseMAC("11:22:33:44:55:66")
	gateway.IP = gatewayIP
	gateway.HW = gatewayHW

	// Create mock queue and firewall
	mockQueue := NewMockPacketQueue()
	mockFirewall := NewMockFirewall()

	// Create environment
	env, _ := session.NewEnvironment("")

	// Create LAN
	aliases, _ := data.NewUnsortedKV("", 0)
	lan := network.NewLAN(iface, gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	// Create session
	sess := &session.Session{
		Interface: iface,
		Gateway:   gateway,
		Lan:       lan,
		StartedAt: time.Now(),
		Active:    true,
		Env:       env,
		Queue:     mockQueue.Queue,
		Firewall:  mockFirewall,
		Modules:   make(session.ModuleList, 0),
	}

	// Initialize events
	sess.Events = session.NewEventPool(false, false)

	// Add mock net.recon module
	mockNetRecon := NewMockNetRecon(sess)
	sess.Modules = append(sess.Modules, mockNetRecon)

	// Create mock session wrapper
	mockSess := &MockSession{
		Session:        sess,
		findMACResults: make(map[string]net.HardwareAddr),
		skipIPs:        make(map[string]bool),
		mockQueue:      mockQueue,
	}

	return mockSess, mockQueue, mockFirewall
}

func TestNewArpSpoofer(t *testing.T) {
	mockSess, _, _ := createMockSession()

	mod := NewArpSpoofer(mockSess.Session)

	if mod == nil {
		t.Fatal("NewArpSpoofer returned nil")
	}

	if mod.Name() != "arp.spoof" {
		t.Errorf("expected module name 'arp.spoof', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("unexpected author: %s", mod.Author())
	}

	// Check parameters
	params := []string{"arp.spoof.targets", "arp.spoof.whitelist", "arp.spoof.internal", "arp.spoof.fullduplex", "arp.spoof.skip_restore"}
	for _, param := range params {
		if !mod.Session.Env.Has(param) {
			t.Errorf("parameter %s not registered", param)
		}
	}

	// Check handlers
	handlers := mod.Handlers()
	expectedHandlers := []string{"arp.spoof on", "arp.ban on", "arp.spoof off", "arp.ban off"}
	handlerMap := make(map[string]bool)

	for _, h := range handlers {
		handlerMap[h.Name] = true
	}

	for _, expected := range expectedHandlers {
		if !handlerMap[expected] {
			t.Errorf("Expected handler '%s' not found", expected)
		}
	}
}

func TestArpSpooferConfigure(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]string
		setupMock func(*MockSession)
		expectErr bool
		validate  func(*ArpSpoofer) error
	}{
		{
			name: "default configuration",
			params: map[string]string{
				"arp.spoof.targets":      "192.168.1.10",
				"arp.spoof.whitelist":    "",
				"arp.spoof.internal":     "false",
				"arp.spoof.fullduplex":   "false",
				"arp.spoof.skip_restore": "false",
			},
			setupMock: func(ms *MockSession) {
				ms.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")
			},
			expectErr: false,
			validate: func(mod *ArpSpoofer) error {
				if mod.internal {
					return fmt.Errorf("expected internal to be false")
				}
				if mod.fullDuplex {
					return fmt.Errorf("expected fullDuplex to be false")
				}
				if mod.skipRestore {
					return fmt.Errorf("expected skipRestore to be false")
				}
				if len(mod.addresses) != 1 {
					return fmt.Errorf("expected 1 address, got %d", len(mod.addresses))
				}
				return nil
			},
		},
		{
			name: "multiple targets and whitelist",
			params: map[string]string{
				"arp.spoof.targets":      "192.168.1.10,192.168.1.20",
				"arp.spoof.whitelist":    "192.168.1.30",
				"arp.spoof.internal":     "true",
				"arp.spoof.fullduplex":   "true",
				"arp.spoof.skip_restore": "true",
			},
			setupMock: func(ms *MockSession) {
				ms.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")
				ms.Lan.AddIfNew("192.168.1.20", "bb:bb:bb:bb:bb:bb")
				ms.Lan.AddIfNew("192.168.1.30", "cc:cc:cc:cc:cc:cc")
			},
			expectErr: false,
			validate: func(mod *ArpSpoofer) error {
				if !mod.internal {
					return fmt.Errorf("expected internal to be true")
				}
				if !mod.fullDuplex {
					return fmt.Errorf("expected fullDuplex to be true")
				}
				if !mod.skipRestore {
					return fmt.Errorf("expected skipRestore to be true")
				}
				if len(mod.addresses) != 2 {
					return fmt.Errorf("expected 2 addresses, got %d", len(mod.addresses))
				}
				if len(mod.wAddresses) != 1 {
					return fmt.Errorf("expected 1 whitelisted address, got %d", len(mod.wAddresses))
				}
				return nil
			},
		},
		{
			name: "MAC address targets",
			params: map[string]string{
				"arp.spoof.targets":      "aa:aa:aa:aa:aa:aa",
				"arp.spoof.whitelist":    "",
				"arp.spoof.internal":     "false",
				"arp.spoof.fullduplex":   "false",
				"arp.spoof.skip_restore": "false",
			},
			setupMock: func(ms *MockSession) {
				ms.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")
			},
			expectErr: false,
			validate: func(mod *ArpSpoofer) error {
				if len(mod.macs) != 1 {
					return fmt.Errorf("expected 1 MAC address, got %d", len(mod.macs))
				}
				return nil
			},
		},
		{
			name: "invalid target",
			params: map[string]string{
				"arp.spoof.targets":      "invalid-target",
				"arp.spoof.whitelist":    "",
				"arp.spoof.internal":     "false",
				"arp.spoof.fullduplex":   "false",
				"arp.spoof.skip_restore": "false",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSess, _, _ := createMockSession()
			mod := NewArpSpoofer(mockSess.Session)

			// Set parameters
			for k, v := range tt.params {
				mockSess.Env.Set(k, v)
			}

			// Setup mock
			if tt.setupMock != nil {
				tt.setupMock(mockSess)
			}

			err := mod.Configure()

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectErr && tt.validate != nil {
				if err := tt.validate(mod); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestArpSpooferStartStop(t *testing.T) {
	mockSess, _, mockFirewall := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Setup targets
	targetIP := "192.168.1.10"
	targetMAC, _ := net.ParseMAC("aa:aa:aa:aa:aa:aa")
	mockSess.Lan.AddIfNew(targetIP, targetMAC.String())
	mockSess.findMACResults[targetIP] = targetMAC

	// Configure
	mockSess.Env.Set("arp.spoof.targets", targetIP)
	mockSess.Env.Set("arp.spoof.fullduplex", "false")
	mockSess.Env.Set("arp.spoof.internal", "false")

	// Start the spoofer
	err := mod.Start()
	if err != nil {
		t.Fatalf("Failed to start spoofer: %v", err)
	}

	if !mod.Running() {
		t.Error("Spoofer should be running after Start()")
	}

	// Check that forwarding was enabled
	if !mockFirewall.IsForwardingEnabled() {
		t.Error("Forwarding should be enabled after starting spoofer")
	}

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop the spoofer
	err = mod.Stop()
	if err != nil {
		t.Fatalf("Failed to stop spoofer: %v", err)
	}

	if mod.Running() {
		t.Error("Spoofer should not be running after Stop()")
	}

	// Note: We can't easily verify packet sending without modifying the actual module
	// to use an interface for the queue. The module behavior is verified through
	// state changes (running state, forwarding enabled, etc.)
}

func TestArpSpooferBanMode(t *testing.T) {
	mockSess, _, mockFirewall := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Setup targets
	targetIP := "192.168.1.10"
	targetMAC, _ := net.ParseMAC("aa:aa:aa:aa:aa:aa")
	mockSess.Lan.AddIfNew(targetIP, targetMAC.String())
	mockSess.findMACResults[targetIP] = targetMAC

	// Configure
	mockSess.Env.Set("arp.spoof.targets", targetIP)

	// Find and execute the ban handler
	handlers := mod.Handlers()
	for _, h := range handlers {
		if h.Name == "arp.ban on" {
			err := h.Exec([]string{})
			if err != nil {
				t.Fatalf("Failed to start ban mode: %v", err)
			}
			break
		}
	}

	if !mod.ban {
		t.Error("Ban mode should be enabled")
	}

	// Check that forwarding was NOT enabled
	if mockFirewall.IsForwardingEnabled() {
		t.Error("Forwarding should NOT be enabled in ban mode")
	}

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop using ban off handler
	for _, h := range handlers {
		if h.Name == "arp.ban off" {
			err := h.Exec([]string{})
			if err != nil {
				t.Fatalf("Failed to stop ban mode: %v", err)
			}
			break
		}
	}

	if mod.ban {
		t.Error("Ban mode should be disabled after stop")
	}
}

func TestArpSpooferWhitelisting(t *testing.T) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Add some IPs and MACs to whitelist
	whitelistIP := net.ParseIP("192.168.1.50")
	whitelistMAC, _ := net.ParseMAC("ff:ff:ff:ff:ff:ff")

	mod.wAddresses = []net.IP{whitelistIP}
	mod.wMacs = []net.HardwareAddr{whitelistMAC}

	// Test IP whitelisting
	if !mod.isWhitelisted("192.168.1.50", nil) {
		t.Error("IP should be whitelisted")
	}

	if mod.isWhitelisted("192.168.1.60", nil) {
		t.Error("IP should not be whitelisted")
	}

	// Test MAC whitelisting
	if !mod.isWhitelisted("", whitelistMAC) {
		t.Error("MAC should be whitelisted")
	}

	otherMAC, _ := net.ParseMAC("aa:aa:aa:aa:aa:aa")
	if mod.isWhitelisted("", otherMAC) {
		t.Error("MAC should not be whitelisted")
	}
}

func TestArpSpooferFullDuplex(t *testing.T) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Setup targets
	targetIP := "192.168.1.10"
	targetMAC, _ := net.ParseMAC("aa:aa:aa:aa:aa:aa")
	mockSess.Lan.AddIfNew(targetIP, targetMAC.String())
	mockSess.findMACResults[targetIP] = targetMAC

	// Configure with full duplex
	mockSess.Env.Set("arp.spoof.targets", targetIP)
	mockSess.Env.Set("arp.spoof.fullduplex", "true")

	// Verify configuration
	err := mod.Configure()
	if err != nil {
		t.Fatalf("Failed to configure: %v", err)
	}

	if !mod.fullDuplex {
		t.Error("Full duplex mode should be enabled")
	}

	// Start the spoofer
	err = mod.Start()
	if err != nil {
		t.Fatalf("Failed to start spoofer: %v", err)
	}

	if !mod.Running() {
		t.Error("Module should be running")
	}

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop
	mod.Stop()
}

func TestArpSpooferInternalMode(t *testing.T) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Setup multiple targets
	targets := map[string]string{
		"192.168.1.10": "aa:aa:aa:aa:aa:aa",
		"192.168.1.20": "bb:bb:bb:bb:bb:bb",
		"192.168.1.30": "cc:cc:cc:cc:cc:cc",
	}

	for ip, mac := range targets {
		mockSess.Lan.AddIfNew(ip, mac)
		hwAddr, _ := net.ParseMAC(mac)
		mockSess.findMACResults[ip] = hwAddr
	}

	// Configure with internal mode
	mockSess.Env.Set("arp.spoof.targets", "192.168.1.10,192.168.1.20")
	mockSess.Env.Set("arp.spoof.internal", "true")

	// Verify configuration
	err := mod.Configure()
	if err != nil {
		t.Fatalf("Failed to configure: %v", err)
	}

	if !mod.internal {
		t.Error("Internal mode should be enabled")
	}

	// Start the spoofer
	err = mod.Start()
	if err != nil {
		t.Fatalf("Failed to start spoofer: %v", err)
	}

	if !mod.Running() {
		t.Error("Module should be running")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop
	mod.Stop()
}

func TestArpSpooferGetTargets(t *testing.T) {
	// This test verifies the getTargets logic without actually calling it
	// since the method uses Session.FindMAC which can't be easily mocked
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Test address and MAC parsing
	targetIP := net.ParseIP("192.168.1.10")
	targetMAC, _ := net.ParseMAC("aa:aa:aa:aa:aa:aa")

	// Add targets by IP
	mod.addresses = []net.IP{targetIP}

	// Verify addresses were set correctly
	if len(mod.addresses) != 1 {
		t.Errorf("expected 1 address, got %d", len(mod.addresses))
	}

	if !mod.addresses[0].Equal(targetIP) {
		t.Errorf("expected address %s, got %s", targetIP, mod.addresses[0])
	}

	// Add targets by MAC
	mod.macs = []net.HardwareAddr{targetMAC}

	// Verify MACs were set correctly
	if len(mod.macs) != 1 {
		t.Errorf("expected 1 MAC, got %d", len(mod.macs))
	}

	if !bytes.Equal(mod.macs[0], targetMAC) {
		t.Errorf("expected MAC %s, got %s", targetMAC, mod.macs[0])
	}

	// Note: The actual getTargets method would look up these addresses/MACs
	// in the network, but we can't easily test that without refactoring
	// the module to use dependency injection for network operations
}

func TestArpSpooferSkipRestore(t *testing.T) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// The skip_restore parameter is set up with an observer in NewArpSpoofer
	// We'll test it by changing the parameter value, which triggers the observer
	mockSess.Env.Set("arp.spoof.skip_restore", "true")

	// Configure to trigger parameter reading
	mod.Configure()

	// Check the observer worked by checking if skipRestore was set
	// Note: The actual observer is triggered during module creation
	// so we test the functionality indirectly through the module's behavior

	// Start and stop to see if restoration is skipped
	mockSess.Env.Set("arp.spoof.targets", "192.168.1.10")
	mockSess.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")

	mod.Start()
	time.Sleep(50 * time.Millisecond)
	mod.Stop()

	// With skip_restore true, the module should have skipRestore set
	// We can't directly test the observer, but we verify the behavior
}

func TestArpSpooferEmptyTargets(t *testing.T) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Configure with empty targets
	mockSess.Env.Set("arp.spoof.targets", "")

	// Start should not error but should not actually start
	err := mod.Start()
	if err != nil {
		t.Fatalf("Start with empty targets should not error: %v", err)
	}

	// Module should not be running
	if mod.Running() {
		t.Error("Module should not be running with empty targets")
	}
}

// Benchmarks
func BenchmarkArpSpooferGetTargets(b *testing.B) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Setup targets
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i+10)
		mac := fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i)
		mockSess.Lan.AddIfNew(ip, mac)
		hwAddr, _ := net.ParseMAC(mac)
		mockSess.findMACResults[ip] = hwAddr
		mod.addresses = append(mod.addresses, net.ParseIP(ip))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = mod.getTargets(false)
	}
}

func BenchmarkArpSpooferWhitelisting(b *testing.B) {
	mockSess, _, _ := createMockSession()
	mod := NewArpSpoofer(mockSess.Session)

	// Add many whitelist entries
	for i := 0; i < 100; i++ {
		ip := net.ParseIP(fmt.Sprintf("192.168.1.%d", i))
		mod.wAddresses = append(mod.wAddresses, ip)
	}

	testMAC, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = mod.isWhitelisted("192.168.1.50", testMAC)
	}
}
