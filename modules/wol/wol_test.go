package wol

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"github.com/bettercap/bettercap/v2/session"
)

var (
	testSession *session.Session
	sessionOnce sync.Once
)

func createMockSession(t *testing.T) *session.Session {
	sessionOnce.Do(func() {
		var err error
		testSession, err = session.New()
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		// Initialize interface with mock data to avoid nil pointer
		// For now, we'll skip initializing these as they require more complex setup
		// The tests will handle the nil cases appropriately
	})
	return testSession
}

func TestNewWOL(t *testing.T) {
	s := createMockSession(t)
	mod := NewWOL(s)

	if mod == nil {
		t.Fatal("NewWOL returned nil")
	}

	if mod.Name() != "wol" {
		t.Errorf("Expected name 'wol', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("Unexpected author: %s", mod.Author())
	}

	if mod.Description() == "" {
		t.Error("Empty description")
	}

	// Check handlers
	handlers := []string{"wol.eth MAC", "wol.udp MAC"}
	for _, handlerName := range handlers {
		found := false
		for _, h := range mod.Handlers() {
			if h.Name == handlerName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Handler '%s' not found", handlerName)
		}
	}
}

func TestParseMAC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			want:    "ff:ff:ff:ff:ff:ff",
			wantErr: false,
		},
		{
			name:    "empty string arg",
			args:    []string{""},
			want:    "ff:ff:ff:ff:ff:ff",
			wantErr: false,
		},
		{
			name:    "valid MAC with colons",
			args:    []string{"aa:bb:cc:dd:ee:ff"},
			want:    "aa:bb:cc:dd:ee:ff",
			wantErr: false,
		},
		{
			name:    "valid MAC with dashes",
			args:    []string{"aa-bb-cc-dd-ee-ff"},
			want:    "aa-bb-cc-dd-ee-ff",
			wantErr: false,
		},
		{
			name:    "valid MAC uppercase",
			args:    []string{"AA:BB:CC:DD:EE:FF"},
			want:    "AA:BB:CC:DD:EE:FF",
			wantErr: false,
		},
		{
			name:    "valid MAC mixed case",
			args:    []string{"aA:bB:cC:dD:eE:fF"},
			want:    "aA:bB:cC:dD:eE:fF",
			wantErr: false,
		},
		{
			name:    "invalid MAC - too short",
			args:    []string{"aa:bb:cc:dd:ee"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid MAC - too long",
			args:    []string{"aa:bb:cc:dd:ee:ff:gg"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid MAC - bad characters",
			args:    []string{"aa:bb:cc:dd:ee:gg"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid MAC - no separators",
			args:    []string{"aabbccddeeff"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "MAC with spaces",
			args:    []string{" aa:bb:cc:dd:ee:ff "},
			want:    "aa:bb:cc:dd:ee:ff",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMAC(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMAC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPayload(t *testing.T) {
	tests := []struct {
		name string
		mac  string
	}{
		{
			name: "broadcast MAC",
			mac:  "ff:ff:ff:ff:ff:ff",
		},
		{
			name: "specific MAC",
			mac:  "aa:bb:cc:dd:ee:ff",
		},
		{
			name: "zeros MAC",
			mac:  "00:00:00:00:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := buildPayload(tt.mac)

			// Payload should be 102 bytes: 6 bytes sync + 16 * 6 bytes MAC
			if len(payload) != 102 {
				t.Errorf("buildPayload() length = %d, want 102", len(payload))
			}

			// First 6 bytes should be 0xff
			for i := 0; i < 6; i++ {
				if payload[i] != 0xff {
					t.Errorf("payload[%d] = %x, want 0xff", i, payload[i])
				}
			}

			// Parse the MAC for comparison
			parsedMAC, _ := net.ParseMAC(tt.mac)

			// Next 16 copies of the MAC
			for i := 0; i < 16; i++ {
				start := 6 + i*6
				end := start + 6
				if !bytes.Equal(payload[start:end], parsedMAC) {
					t.Errorf("MAC copy %d = %x, want %x", i, payload[start:end], parsedMAC)
				}
			}
		})
	}
}

func TestWOLConfigure(t *testing.T) {
	s := createMockSession(t)
	mod := NewWOL(s)

	if err := mod.Configure(); err != nil {
		t.Errorf("Configure() error = %v", err)
	}
}

func TestWOLStartStop(t *testing.T) {
	s := createMockSession(t)
	mod := NewWOL(s)

	if err := mod.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if err := mod.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestWOLHandlers(t *testing.T) {
	// Only test parseMAC validation since the actual handlers require a fully initialized session
	testCases := []struct {
		name    string
		args    []string
		wantMAC string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantMAC: "ff:ff:ff:ff:ff:ff",
			wantErr: false,
		},
		{
			name:    "valid MAC",
			args:    []string{"aa:bb:cc:dd:ee:ff"},
			wantMAC: "aa:bb:cc:dd:ee:ff",
			wantErr: false,
		},
		{
			name:    "invalid MAC",
			args:    []string{"invalid:mac"},
			wantMAC: "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mac, err := parseMAC(tc.args)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseMAC() error = %v, wantErr %v", err, tc.wantErr)
			}
			if mac != tc.wantMAC {
				t.Errorf("parseMAC() = %v, want %v", mac, tc.wantMAC)
			}
		})
	}
}

func TestWOLMethods(t *testing.T) {
	s := createMockSession(t)
	mod := NewWOL(s)

	// Test that the methods exist and can be called without panic
	// The actual execution will fail due to nil session interface/queue
	// but we're testing the module structure

	// Check that handlers were properly registered
	expectedHandlers := 2 // wol.eth and wol.udp
	if len(mod.Handlers()) != expectedHandlers {
		t.Errorf("Expected %d handlers, got %d", expectedHandlers, len(mod.Handlers()))
	}

	// Verify handler names
	handlerNames := make(map[string]bool)
	for _, h := range mod.Handlers() {
		handlerNames[h.Name] = true
	}

	if !handlerNames["wol.eth MAC"] {
		t.Error("wol.eth handler not found")
	}
	if !handlerNames["wol.udp MAC"] {
		t.Error("wol.udp handler not found")
	}
}

func TestReMAC(t *testing.T) {
	tests := []struct {
		mac   string
		valid bool
	}{
		{"aa:bb:cc:dd:ee:ff", true},
		{"AA:BB:CC:DD:EE:FF", true},
		{"aa-bb-cc-dd-ee-ff", true},
		{"AA-BB-CC-DD-EE-FF", true},
		{"aA:bB:cC:dD:eE:fF", true},
		{"00:00:00:00:00:00", true},
		{"ff:ff:ff:ff:ff:ff", true},
		{"aabbccddeeff", false},
		{"aa:bb:cc:dd:ee", false},
		{"aa:bb:cc:dd:ee:ff:gg", false},
		{"aa:bb:cc:dd:ee:gg", false},
		{"zz:zz:zz:zz:zz:zz", false},
		{"", false},
		{"not a mac", false},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			if got := reMAC.MatchString(tt.mac); got != tt.valid {
				t.Errorf("reMAC.MatchString(%q) = %v, want %v", tt.mac, got, tt.valid)
			}
		})
	}
}

// Test that the module sets running state correctly
func TestWOLRunningState(t *testing.T) {
	s := createMockSession(t)
	mod := NewWOL(s)

	// Initially should not be running
	if mod.Running() {
		t.Error("Module should not be running initially")
	}

	// Note: wolETH and wolUDP will fail due to nil session.Queue,
	// but they should still set the running state before failing
}

// Benchmark tests
func BenchmarkBuildPayload(b *testing.B) {
	mac := "aa:bb:cc:dd:ee:ff"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPayload(mac)
	}
}

func BenchmarkParseMAC(b *testing.B) {
	args := []string{"aa:bb:cc:dd:ee:ff"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseMAC(args)
	}
}

func BenchmarkReMAC(b *testing.B) {
	mac := "aa:bb:cc:dd:ee:ff"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reMAC.MatchString(mac)
	}
}
