package net_recon

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/modules/utils"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/packets"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/data"
)

// Mock ArpUpdate function
var mockArpUpdateFunc func(string) (network.ArpTable, error)

// Override the network.ArpUpdate function for testing
func mockArpUpdate(iface string) (network.ArpTable, error) {
	if mockArpUpdateFunc != nil {
		return mockArpUpdateFunc(iface)
	}
	return make(network.ArpTable), nil
}

// MockLAN implements a mock version of the LAN interface
type MockLAN struct {
	sync.RWMutex
	hosts        map[string]*network.Endpoint
	wasMissed    map[string]bool
	addedHosts   []string
	removedHosts []string
}

func NewMockLAN() *MockLAN {
	return &MockLAN{
		hosts:        make(map[string]*network.Endpoint),
		wasMissed:    make(map[string]bool),
		addedHosts:   []string{},
		removedHosts: []string{},
	}
}

func (m *MockLAN) AddIfNew(ip, mac string) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.hosts[mac]; !exists {
		m.hosts[mac] = &network.Endpoint{
			IpAddress: ip,
			HwAddress: mac,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
		m.addedHosts = append(m.addedHosts, mac)
	}
}

func (m *MockLAN) Remove(ip, mac string) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.hosts[mac]; exists {
		delete(m.hosts, mac)
		m.removedHosts = append(m.removedHosts, mac)
	}
}

func (m *MockLAN) Clear() {
	m.Lock()
	defer m.Unlock()

	m.hosts = make(map[string]*network.Endpoint)
	m.wasMissed = make(map[string]bool)
	m.addedHosts = []string{}
	m.removedHosts = []string{}
}

func (m *MockLAN) EachHost(cb func(mac string, e *network.Endpoint)) {
	m.RLock()
	defer m.RUnlock()

	for mac, host := range m.hosts {
		cb(mac, host)
	}
}

func (m *MockLAN) List() []*network.Endpoint {
	m.RLock()
	defer m.RUnlock()

	list := make([]*network.Endpoint, 0, len(m.hosts))
	for _, host := range m.hosts {
		list = append(list, host)
	}
	return list
}

func (m *MockLAN) WasMissed(mac string) bool {
	m.RLock()
	defer m.RUnlock()

	return m.wasMissed[mac]
}

func (m *MockLAN) Get(mac string) *network.Endpoint {
	m.RLock()
	defer m.RUnlock()

	return m.hosts[mac]
}

// Create a mock session for testing
func createMockSession() *session.Session {
	iface := &network.Endpoint{
		IpAddress: "192.168.1.100",
		HwAddress: "aa:bb:cc:dd:ee:ff",
		Hostname:  "eth0",
	}
	iface.SetIP("192.168.1.100")
	iface.SetBits(24)

	gateway := &network.Endpoint{
		IpAddress: "192.168.1.1",
		HwAddress: "11:22:33:44:55:66",
	}

	// Create environment
	env, _ := session.NewEnvironment("")

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

	// Initialize the Events field with a mock EventPool
	sess.Events = session.NewEventPool(false, false)

	return sess
}

func TestNewDiscovery(t *testing.T) {
	sess := createMockSession()
	mod := NewDiscovery(sess)

	if mod == nil {
		t.Fatal("NewDiscovery returned nil")
	}

	if mod.Name() != "net.recon" {
		t.Errorf("expected module name 'net.recon', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("unexpected author: %s", mod.Author())
	}

	if mod.selector == nil {
		t.Error("selector should be initialized")
	}
}

func TestRunDiff(t *testing.T) {
	// Test the basic diff functionality with a simpler approach
	tests := []struct {
		name            string
		initialHosts    map[string]string // IP -> MAC
		arpTable        network.ArpTable
		expectedAdded   []string
		expectedRemoved []string
	}{
		{
			name: "no changes",
			initialHosts: map[string]string{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
				"192.168.1.20": "bb:bb:bb:bb:bb:bb",
			},
			arpTable: network.ArpTable{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
				"192.168.1.20": "bb:bb:bb:bb:bb:bb",
			},
			expectedAdded:   []string{},
			expectedRemoved: []string{},
		},
		{
			name: "new host discovered",
			initialHosts: map[string]string{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
			},
			arpTable: network.ArpTable{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
				"192.168.1.20": "bb:bb:bb:bb:bb:bb",
			},
			expectedAdded:   []string{"bb:bb:bb:bb:bb:bb"},
			expectedRemoved: []string{},
		},
		{
			name: "host disappeared",
			initialHosts: map[string]string{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
				"192.168.1.20": "bb:bb:bb:bb:bb:bb",
			},
			arpTable: network.ArpTable{
				"192.168.1.10": "aa:aa:aa:aa:aa:aa",
			},
			expectedAdded:   []string{},
			expectedRemoved: []string{"bb:bb:bb:bb:bb:bb"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := createMockSession()

			// Track callbacks
			addedHosts := []string{}
			removedHosts := []string{}

			newCb := func(e *network.Endpoint) {
				addedHosts = append(addedHosts, e.HwAddress)
			}

			lostCb := func(e *network.Endpoint) {
				removedHosts = append(removedHosts, e.HwAddress)
			}

			aliases, _ := data.NewUnsortedKV("", 0)
			sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, newCb, lostCb)

			mod := &Discovery{
				SessionModule: session.NewSessionModule("net.recon", sess),
			}

			// Add initial hosts
			for ip, mac := range tt.initialHosts {
				sess.Lan.AddIfNew(ip, mac)
			}

			// Reset tracking
			addedHosts = []string{}
			removedHosts = []string{}

			// Add interface and gateway to ARP table to avoid them being removed
			finalArpTable := make(network.ArpTable)
			for k, v := range tt.arpTable {
				finalArpTable[k] = v
			}
			finalArpTable[sess.Interface.IpAddress] = sess.Interface.HwAddress
			finalArpTable[sess.Gateway.IpAddress] = sess.Gateway.HwAddress

			// Run the diff multiple times to trigger actual removal (TTL countdown)
			for i := 0; i < network.LANDefaultttl+1; i++ {
				mod.runDiff(finalArpTable)
			}

			// Check results
			if len(addedHosts) != len(tt.expectedAdded) {
				t.Errorf("expected %d added hosts, got %d. Added: %v", len(tt.expectedAdded), len(addedHosts), addedHosts)
			}

			if len(removedHosts) != len(tt.expectedRemoved) {
				t.Errorf("expected %d removed hosts, got %d. Removed: %v", len(tt.expectedRemoved), len(removedHosts), removedHosts)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	sess := createMockSession()
	mod := NewDiscovery(sess)

	err := mod.Configure()
	if err != nil {
		t.Errorf("Configure() returned error: %v", err)
	}
}

func TestStartStop(t *testing.T) {
	sess := createMockSession()
	aliases, _ := data.NewUnsortedKV("", 0)
	sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	mod := NewDiscovery(sess)

	// Test starting the module
	err := mod.Start()
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	if !mod.Running() {
		t.Error("module should be running after Start()")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Test stopping the module
	err = mod.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if mod.Running() {
		t.Error("module should not be running after Stop()")
	}
}

func TestShowMethods(t *testing.T) {
	// Skip this test as it requires a full session with readline
	t.Skip("Skipping TestShowMethods as it requires readline initialization")
}

func TestDoSelection(t *testing.T) {
	sess := createMockSession()
	aliases, _ := data.NewUnsortedKV("", 0)
	sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	// Add test endpoints
	sess.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")
	sess.Lan.AddIfNew("192.168.1.20", "bb:bb:bb:bb:bb:bb")
	sess.Lan.AddIfNew("192.168.1.30", "cc:cc:cc:cc:cc:cc")

	// Get endpoints and set additional properties
	if e, found := sess.Lan.Get("aa:aa:aa:aa:aa:aa"); found {
		e.Hostname = "host1"
		e.Vendor = "Vendor1"
	}

	if e, found := sess.Lan.Get("bb:bb:bb:bb:bb:bb"); found {
		e.Alias = "mydevice"
		e.Vendor = "Vendor2"
	}

	mod := NewDiscovery(sess)
	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "net.show",
		[]string{"ip", "mac", "seen", "sent", "rcvd"}, "ip asc")

	tests := []struct {
		name          string
		arg           string
		expectedCount int
		expectedIPs   []string
	}{
		{
			name:          "select all",
			arg:           "",
			expectedCount: 3,
		},
		{
			name:          "select by IP",
			arg:           "192.168.1.10",
			expectedCount: 1,
			expectedIPs:   []string{"192.168.1.10"},
		},
		{
			name:          "select by MAC",
			arg:           "aa:aa:aa:aa:aa:aa",
			expectedCount: 1,
			expectedIPs:   []string{"192.168.1.10"},
		},
		{
			name:          "select multiple by comma",
			arg:           "192.168.1.10,192.168.1.20",
			expectedCount: 2,
			expectedIPs:   []string{"192.168.1.10", "192.168.1.20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, targets := mod.doSelection(tt.arg)
			if err != nil {
				t.Errorf("doSelection returned error: %v", err)
			}

			if len(targets) != tt.expectedCount {
				t.Errorf("expected %d targets, got %d", tt.expectedCount, len(targets))
			}

			if tt.expectedIPs != nil {
				for _, expectedIP := range tt.expectedIPs {
					found := false
					for _, target := range targets {
						if target.IpAddress == expectedIP {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected to find IP %s in targets", expectedIP)
					}
				}
			}
		})
	}
}

func TestHandlers(t *testing.T) {
	sess := createMockSession()
	aliases, _ := data.NewUnsortedKV("", 0)
	sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	mod := NewDiscovery(sess)

	handlers := []struct {
		name     string
		handler  string
		args     []string
		setup    func()
		validate func() error
	}{
		{
			name:    "net.clear",
			handler: "net.clear",
			args:    []string{},
			setup: func() {
				sess.Lan.AddIfNew("192.168.1.10", "aa:aa:aa:aa:aa:aa")
			},
			validate: func() error {
				// Check if hosts were cleared
				hosts := sess.Lan.List()
				if len(hosts) != 0 {
					return fmt.Errorf("expected empty hosts after clear, got %d", len(hosts))
				}
				return nil
			},
		},
	}

	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			// Find and execute the handler
			found := false
			for _, h := range mod.Handlers() {
				if h.Name == tt.handler {
					found = true
					err := h.Exec(tt.args)
					if err != nil {
						t.Errorf("handler %s returned error: %v", tt.handler, err)
					}
					break
				}
			}

			if !found {
				t.Errorf("handler %s not found", tt.handler)
			}

			if tt.validate != nil {
				if err := tt.validate(); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGetRow(t *testing.T) {
	sess := createMockSession()
	aliases, _ := data.NewUnsortedKV("", 0)
	sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	mod := NewDiscovery(sess)

	// Test endpoint with metadata
	endpoint := &network.Endpoint{
		IpAddress: "192.168.1.10",
		HwAddress: "aa:aa:aa:aa:aa:aa",
		Hostname:  "testhost",
		Vendor:    "Test Vendor",
		FirstSeen: time.Now().Add(-time.Hour),
		LastSeen:  time.Now(),
		Meta:      network.NewMeta(),
	}
	endpoint.Meta.Set("key1", "value1")
	endpoint.Meta.Set("key2", "value2")

	// Test without meta
	rows := mod.getRow(endpoint, false)
	if len(rows) != 1 {
		t.Errorf("expected 1 row without meta, got %d", len(rows))
	}
	if len(rows[0]) != 7 {
		t.Errorf("expected 7 columns, got %d", len(rows[0]))
	}

	// Test with meta
	rows = mod.getRow(endpoint, true)
	if len(rows) != 2 { // One main row + one meta row per metadata entry
		t.Errorf("expected 2 rows with meta, got %d", len(rows))
	}

	// Test interface endpoint
	ifaceEndpoint := sess.Interface
	rows = mod.getRow(ifaceEndpoint, false)
	if len(rows) != 1 {
		t.Errorf("expected 1 row for interface, got %d", len(rows))
	}

	// Test gateway endpoint
	gatewayEndpoint := sess.Gateway
	rows = mod.getRow(gatewayEndpoint, false)
	if len(rows) != 1 {
		t.Errorf("expected 1 row for gateway, got %d", len(rows))
	}
}

func TestDoFilter(t *testing.T) {
	sess := createMockSession()
	mod := NewDiscovery(sess)
	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "net.show",
		[]string{"ip", "mac", "seen", "sent", "rcvd"}, "ip asc")

	// Test that doFilter behavior matches the actual implementation
	// When Expression is nil, it returns true (no filtering)
	// When Expression is set, it matches against any of the fields

	tests := []struct {
		name        string
		filter      string
		endpoint    *network.Endpoint
		shouldMatch bool
	}{
		{
			name:   "no filter",
			filter: "",
			endpoint: &network.Endpoint{
				IpAddress: "192.168.1.10",
				Meta:      network.NewMeta(),
			},
			shouldMatch: true,
		},
		{
			name:   "ip filter match",
			filter: "192.168",
			endpoint: &network.Endpoint{
				IpAddress: "192.168.1.10",
				Meta:      network.NewMeta(),
			},
			shouldMatch: true,
		},
		{
			name:   "mac filter match",
			filter: "aa:bb",
			endpoint: &network.Endpoint{
				IpAddress: "192.168.1.10",
				HwAddress: "aa:bb:cc:dd:ee:ff",
				Meta:      network.NewMeta(),
			},
			shouldMatch: true,
		},
		{
			name:   "hostname filter match",
			filter: "myhost",
			endpoint: &network.Endpoint{
				IpAddress: "192.168.1.10",
				Hostname:  "myhost.local",
				Meta:      network.NewMeta(),
			},
			shouldMatch: true,
		},
		{
			name:   "no match - testing unique string",
			filter: "xyz123nomatch",
			endpoint: &network.Endpoint{
				IpAddress:  "192.168.1.10",
				Ip6Address: "",
				HwAddress:  "aa:bb:cc:dd:ee:ff",
				Hostname:   "host.local",
				Alias:      "",
				Vendor:     "",
				Meta:       network.NewMeta(),
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset selector for each test
			// Set the parameter value that Update() will read
			sess.Env.Set("net.show.filter", tt.filter)
			mod.selector.Expression = nil

			// Update will read from the parameter
			err := mod.selector.Update()
			if err != nil {
				t.Fatalf("selector.Update() failed: %v", err)
			}

			result := mod.doFilter(tt.endpoint)
			if result != tt.shouldMatch {
				if mod.selector.Expression != nil {
					t.Errorf("expected doFilter to return %v, got %v. Regex: %s", tt.shouldMatch, result, mod.selector.Expression.String())
				} else {
					t.Errorf("expected doFilter to return %v, got %v. Expression is nil", tt.shouldMatch, result)
				}
			}
		})
	}
}

// Benchmark the runDiff method
func BenchmarkRunDiff(b *testing.B) {
	sess := createMockSession()
	aliases, _ := data.NewUnsortedKV("", 0)
	sess.Lan = network.NewLAN(sess.Interface, sess.Gateway, aliases, func(e *network.Endpoint) {}, func(e *network.Endpoint) {})

	mod := &Discovery{
		SessionModule: session.NewSessionModule("net.recon", sess),
	}

	// Create a large ARP table
	arpTable := make(network.ArpTable)
	for i := 0; i < 100; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		mac := fmt.Sprintf("aa:bb:cc:dd:%02x:%02x", i/256, i%256)
		arpTable[ip] = mac

		// Add half to the existing LAN
		if i < 50 {
			sess.Lan.AddIfNew(ip, mac)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.runDiff(arpTable)
	}
}
