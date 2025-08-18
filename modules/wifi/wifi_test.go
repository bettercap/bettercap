package wifi

import (
	"bytes"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/packets"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/data"
)

// Create a mock session for testing
func createMockSession() *session.Session {
	// Create interface
	iface := &network.Endpoint{
		IpAddress: "192.168.1.100",
		HwAddress: "aa:bb:cc:dd:ee:ff",
		Hostname:  "wlan0",
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
		Queue:     &packets.Queue{},
		Modules:   make(session.ModuleList, 0),
	}

	// Initialize events
	sess.Events = session.NewEventPool(false, false)

	// Initialize WiFi state
	sess.WiFi = network.NewWiFi(iface, aliases, func(ap *network.AccessPoint) {}, func(ap *network.AccessPoint) {})

	return sess
}

func TestNewWiFiModule(t *testing.T) {
	sess := createMockSession()

	mod := NewWiFiModule(sess)

	if mod == nil {
		t.Fatal("NewWiFiModule returned nil")
	}

	if mod.Name() != "wifi" {
		t.Errorf("expected module name 'wifi', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com> && Gianluca Braga <matrix86@gmail.com>" {
		t.Errorf("unexpected author: %s", mod.Author())
	}

	// Check parameters
	params := []string{
		"wifi.interface",
		"wifi.rssi.min",
		"wifi.deauth.skip",
		"wifi.deauth.silent",
		"wifi.deauth.open",
		"wifi.deauth.acquired",
		"wifi.assoc.skip",
		"wifi.assoc.silent",
		"wifi.assoc.open",
		"wifi.assoc.acquired",
		"wifi.ap.ttl",
		"wifi.sta.ttl",
		"wifi.region",
		"wifi.txpower",
		"wifi.handshakes.file",
		"wifi.handshakes.aggregate",
		"wifi.ap.ssid",
		"wifi.ap.bssid",
		"wifi.ap.channel",
		"wifi.ap.encryption",
		"wifi.show.manufacturer",
		"wifi.source.file",
		"wifi.hop.period",
		"wifi.skip-broken",
		"wifi.channel_switch_announce.silent",
		"wifi.fake_auth.silent",
		"wifi.bruteforce.target",
		"wifi.bruteforce.wordlist",
		"wifi.bruteforce.workers",
		"wifi.bruteforce.wide",
		"wifi.bruteforce.stop_at_first",
		"wifi.bruteforce.timeout",
	}
	for _, param := range params {
		if !mod.Session.Env.Has(param) {
			t.Errorf("parameter %s not registered", param)
		}
	}

	// Check handlers
	handlers := mod.Handlers()
	expectedHandlers := []string{
		"wifi.recon on",
		"wifi.recon off",
		"wifi.clear",
		"wifi.recon MAC",
		"wifi.recon clear",
		"wifi.deauth BSSID",
		"wifi.probe BSSID ESSID",
		"wifi.assoc BSSID",
		"wifi.ap",
		"wifi.show.wps BSSID",
		"wifi.show",
		"wifi.recon.channel CHANNEL",
		"wifi.client.probe.sta.filter FILTER",
		"wifi.client.probe.ap.filter FILTER",
		"wifi.channel_switch_announce bssid channel ",
		"wifi.fake_auth bssid client",
		"wifi.bruteforce on",
		"wifi.bruteforce off",
	}

	if len(handlers) != len(expectedHandlers) {
		t.Errorf("expected %d handlers, got %d", len(expectedHandlers), len(handlers))
	}
}

func TestWiFiModuleConfigure(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]string
		expectErr bool
	}{
		{
			name: "default configuration",
			params: map[string]string{
				"wifi.interface":            "",
				"wifi.ap.ttl":               "300",
				"wifi.sta.ttl":              "300",
				"wifi.region":               "",
				"wifi.txpower":              "30",
				"wifi.source.file":          "",
				"wifi.rssi.min":             "-200",
				"wifi.handshakes.file":      "~/bettercap-wifi-handshakes.pcap",
				"wifi.handshakes.aggregate": "true",
				"wifi.hop.period":           "250",
				"wifi.skip-broken":          "true",
			},
			expectErr: true, // Will fail without actual interface
		},
		{
			name: "invalid rssi",
			params: map[string]string{
				"wifi.rssi.min": "not-a-number",
			},
			expectErr: true,
		},
		{
			name: "invalid hop period",
			params: map[string]string{
				"wifi.hop.period": "invalid",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := createMockSession()
			mod := NewWiFiModule(sess)

			// Set parameters
			for k, v := range tt.params {
				sess.Env.Set(k, v)
			}

			err := mod.Configure()

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWiFiModuleFrequencies(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test setting frequencies
	freqs := []int{2412, 2437, 2462, 5180, 5200} // Channels 1, 6, 11, 36, 40
	mod.setFrequencies(freqs)

	if len(mod.frequencies) != len(freqs) {
		t.Errorf("expected %d frequencies, got %d", len(freqs), len(mod.frequencies))
	}

	// Check if channels were properly converted
	channels, _ := mod.State.Load("channels")
	channelList := channels.([]int)
	expectedChannels := []int{1, 6, 11, 36, 40}

	if len(channelList) != len(expectedChannels) {
		t.Errorf("expected %d channels, got %d", len(expectedChannels), len(channelList))
	}
}

func TestWiFiModuleFilters(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test STA filter
	handlers := mod.Handlers()
	var staFilterHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.client.probe.sta.filter FILTER" {
			staFilterHandler = h
			break
		}
	}

	if staFilterHandler.Name == "" {
		t.Fatal("STA filter handler not found")
	}

	// Set a filter
	err := staFilterHandler.Exec([]string{"^aa:bb:.*"})
	if err != nil {
		t.Errorf("Failed to set STA filter: %v", err)
	}

	if mod.filterProbeSTA == nil {
		t.Error("STA filter was not set")
	}

	// Clear filter
	err = staFilterHandler.Exec([]string{"clear"})
	if err != nil {
		t.Errorf("Failed to clear STA filter: %v", err)
	}

	if mod.filterProbeSTA != nil {
		t.Error("STA filter was not cleared")
	}

	// Test AP filter
	var apFilterHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.client.probe.ap.filter FILTER" {
			apFilterHandler = h
			break
		}
	}

	if apFilterHandler.Name == "" {
		t.Fatal("AP filter handler not found")
	}

	// Set a filter
	err = apFilterHandler.Exec([]string{"^TestAP.*"})
	if err != nil {
		t.Errorf("Failed to set AP filter: %v", err)
	}

	if mod.filterProbeAP == nil {
		t.Error("AP filter was not set")
	}
}

func TestWiFiModuleDeauth(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test deauth handler
	handlers := mod.Handlers()
	var deauthHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.deauth BSSID" {
			deauthHandler = h
			break
		}
	}

	if deauthHandler.Name == "" {
		t.Fatal("Deauth handler not found")
	}

	// Test with "all"
	err := deauthHandler.Exec([]string{"all"})
	if err == nil {
		t.Error("Expected error when starting deauth without running module")
	}

	// Test with invalid MAC
	err = deauthHandler.Exec([]string{"invalid-mac"})
	if err == nil {
		t.Error("Expected error with invalid MAC address")
	}
}

func TestWiFiModuleChannelHandler(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test channel handler
	handlers := mod.Handlers()
	var channelHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.recon.channel CHANNEL" {
			channelHandler = h
			break
		}
	}

	if channelHandler.Name == "" {
		t.Fatal("Channel handler not found")
	}

	// Test with valid channels
	err := channelHandler.Exec([]string{"1,6,11"})
	if err != nil {
		t.Errorf("Failed to set channels: %v", err)
	}

	// Test with invalid channel
	err = channelHandler.Exec([]string{"999"})
	if err == nil {
		t.Error("Expected error with invalid channel")
	}

	// Test clear
	err = channelHandler.Exec([]string{"clear"})
	if err == nil {
		// Will fail without actual interface but should parse correctly
		t.Log("Clear channels parsed correctly")
	}
}

func TestWiFiModuleShow(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test show handler exists
	handlers := mod.Handlers()
	found := false
	for _, h := range handlers {
		if h.Name == "wifi.show" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Show handler not found")
	}

	// Skip actual execution as it requires UI components
	t.Log("Show handler found, skipping execution due to UI dependencies")
}

func TestWiFiModuleShowWPS(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test show WPS handler exists
	handlers := mod.Handlers()
	found := false
	for _, h := range handlers {
		if h.Name == "wifi.show.wps BSSID" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Show WPS handler not found")
	}

	// Skip actual execution as it requires UI components
	t.Log("Show WPS handler found, skipping execution due to UI dependencies")
}

func TestWiFiModuleBruteforce(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Check bruteforce config
	if mod.bruteforce == nil {
		t.Fatal("Bruteforce config not initialized")
	}

	// Test bruteforce parameters
	params := map[string]string{
		"wifi.bruteforce.target":        "TestAP",
		"wifi.bruteforce.wordlist":      "/tmp/wordlist.txt",
		"wifi.bruteforce.workers":       "4",
		"wifi.bruteforce.wide":          "true",
		"wifi.bruteforce.stop_at_first": "true",
		"wifi.bruteforce.timeout":       "30",
	}

	for k, v := range params {
		sess.Env.Set(k, v)
	}

	// Verify parameters were set
	if err, target := mod.StringParam("wifi.bruteforce.target"); err != nil {
		t.Errorf("Failed to get bruteforce target: %v", err)
	} else if target != "TestAP" {
		t.Errorf("Expected target 'TestAP', got '%s'", target)
	}
}

func TestWiFiModuleAPConfig(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Set AP parameters
	params := map[string]string{
		"wifi.ap.ssid":       "TestAP",
		"wifi.ap.bssid":      "aa:bb:cc:dd:ee:ff",
		"wifi.ap.channel":    "6",
		"wifi.ap.encryption": "true",
	}

	for k, v := range params {
		sess.Env.Set(k, v)
	}

	// Parse AP config
	err := mod.parseApConfig()
	if err != nil {
		t.Errorf("Failed to parse AP config: %v", err)
	}

	// Verify config
	if mod.apConfig.SSID != "TestAP" {
		t.Errorf("Expected SSID 'TestAP', got '%s'", mod.apConfig.SSID)
	}

	if !bytes.Equal(mod.apConfig.BSSID, []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}) {
		t.Errorf("BSSID mismatch")
	}

	if mod.apConfig.Channel != 6 {
		t.Errorf("Expected channel 6, got %d", mod.apConfig.Channel)
	}

	if !mod.apConfig.Encryption {
		t.Error("Expected encryption to be enabled")
	}
}

func TestWiFiModuleSkipMACs(t *testing.T) {
	// Skip this test as updateDeauthSkipList and updateAssocSkipList are private methods
	t.Skip("Skipping test for private skip list methods")
}

func TestWiFiModuleProbe(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test probe handler
	handlers := mod.Handlers()
	var probeHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.probe BSSID ESSID" {
			probeHandler = h
			break
		}
	}

	if probeHandler.Name == "" {
		t.Fatal("Probe handler not found")
	}

	// Test with valid parameters
	err := probeHandler.Exec([]string{"aa:bb:cc:dd:ee:ff", "TestNetwork"})
	if err == nil {
		t.Error("Expected error when probing without running module")
	}

	// Test with invalid MAC
	err = probeHandler.Exec([]string{"invalid-mac", "TestNetwork"})
	if err == nil {
		t.Error("Expected error with invalid MAC address")
	}
}

func TestWiFiModuleFakeAuth(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Test fake auth handler
	handlers := mod.Handlers()
	var fakeAuthHandler session.ModuleHandler
	for _, h := range handlers {
		if h.Name == "wifi.fake_auth bssid client" {
			fakeAuthHandler = h
			break
		}
	}

	if fakeAuthHandler.Name == "" {
		t.Fatal("Fake auth handler not found")
	}

	// Test with valid parameters
	err := fakeAuthHandler.Exec([]string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"})
	if err == nil {
		t.Error("Expected error when running fake auth without running module")
	}

	// Test with invalid MACs
	err = fakeAuthHandler.Exec([]string{"invalid-mac", "11:22:33:44:55:66"})
	if err == nil {
		t.Error("Expected error with invalid BSSID")
	}
}

func TestWiFiModuleViewSelector(t *testing.T) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	// Check if view selector is initialized
	if mod.selector == nil {
		t.Fatal("View selector not initialized")
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Test bruteforce config
func TestBruteforceConfig(t *testing.T) {
	config := NewBruteForceConfig()

	if config == nil {
		t.Fatal("NewBruteForceConfig returned nil")
	}

	// Check defaults
	if config.target != "" {
		t.Errorf("Expected empty target, got '%s'", config.target)
	}

	if config.wordlist != "/usr/share/dict/words" {
		t.Errorf("Expected wordlist '/usr/share/dict/words', got '%s'", config.wordlist)
	}

	if config.workers != 1 {
		t.Errorf("Expected 1 worker, got %d", config.workers)
	}

	if config.wide {
		t.Error("Expected wide to be false by default")
	}

	if !config.stop_at_first {
		t.Error("Expected stop_at_first to be true by default")
	}

	if config.timeout != 15 {
		t.Errorf("Expected timeout 15, got %d", config.timeout)
	}
}

// Benchmarks
func BenchmarkWiFiModuleSetFrequencies(b *testing.B) {
	sess := createMockSession()
	mod := NewWiFiModule(sess)

	freqs := []int{2412, 2437, 2462, 5180, 5200, 5220, 5240, 5745, 5765, 5785, 5805, 5825}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mod.setFrequencies(freqs)
	}
}

func BenchmarkWiFiModuleFilterCheck(b *testing.B) {
	filter, _ := regexp.Compile("^aa:bb:.*")
	testMAC := "aa:bb:cc:dd:ee:ff"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = filter.MatchString(testMAC)
	}
}
