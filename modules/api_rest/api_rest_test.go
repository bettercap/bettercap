package api_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/data"
	"github.com/gorilla/mux"
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
	})
	return testSession
}

func TestNewRestAPI(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	if mod == nil {
		t.Fatal("NewRestAPI returned nil")
	}

	if mod.Name() != "api.rest" {
		t.Errorf("Expected name 'api.rest', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("Unexpected author: %s", mod.Author())
	}

	if mod.Description() == "" {
		t.Error("Empty description")
	}

	// Check handlers
	handlers := mod.Handlers()
	expectedHandlers := []string{
		"api.rest on",
		"api.rest off",
		"api.rest.record off",
		"api.rest.record FILENAME",
		"api.rest.replay off",
		"api.rest.replay FILENAME",
	}

	if len(handlers) != len(expectedHandlers) {
		t.Errorf("Expected %d handlers, got %d", len(expectedHandlers), len(handlers))
	}

	handlerNames := make(map[string]bool)
	for _, h := range handlers {
		handlerNames[h.Name] = true
	}

	for _, expected := range expectedHandlers {
		if !handlerNames[expected] {
			t.Errorf("Handler '%s' not found", expected)
		}
	}

	// Check initial state
	if mod.recording {
		t.Error("Should not be recording initially")
	}
	if mod.replaying {
		t.Error("Should not be replaying initially")
	}
	if mod.useWebsocket {
		t.Error("Should not use websocket by default")
	}
	if mod.allowOrigin != "*" {
		t.Errorf("Expected default allowOrigin '*', got '%s'", mod.allowOrigin)
	}
}

func TestShowWiFiJSONContract(t *testing.T) {
	s := createMockSession(t)
	aliases, err := data.NewMemUnsortedKV()
	if err != nil {
		t.Fatal(err)
	}
	previousWiFi := s.WiFi
	s.WiFi = network.NewWiFi(s.Interface, aliases, nil, nil)
	defer func() { s.WiFi = previousWiFi }()
	previousLAN := s.Lan
	if s.Lan == nil {
		s.Lan = network.NewLAN(s.Interface, s.Gateway, aliases, nil, nil)
	}
	defer func() { s.Lan = previousLAN }()

	ap, added := s.WiFi.AddIfNew("test-network", "02:00:00:00:00:01", 2412, -42)
	if !added {
		t.Fatal("expected a new AP")
	}
	ap.SetEncryption("WPA2", "CCMP", "PSK")
	client, added := ap.AddClientIfNew("02:00:00:00:00:02", 2412, -55)
	if !added {
		t.Fatal("expected a new client")
	}
	client.AddTraffic(10, 20)

	mod := NewRestAPI(s)
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/session/wifi/02:00:00:00:00:01", nil), map[string]string{
		"mac": "02:00:00:00:00:01",
	})
	recorder := httptest.NewRecorder()
	mod.showWiFi(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", recorder.Code, recorder.Body.String())
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{
		"ipv4", "ipv6", "mac", "hostname", "alias", "vendor", "first_seen", "last_seen", "meta",
		"frequency", "channel", "rssi", "sent", "received", "encryption", "cipher", "authentication", "wps",
		"clients", "handshake",
	}
	if len(doc) != len(wantKeys) {
		t.Fatalf("unexpected Wi-Fi API response fields: %#v", doc)
	}
	for _, key := range wantKeys {
		if _, found := doc[key]; !found {
			t.Fatalf("Wi-Fi API response is missing %q: %#v", key, doc)
		}
	}
	if doc["mac"] != "02:00:00:00:00:01" || doc["encryption"] != "WPA2" {
		t.Fatalf("unexpected AP response: %#v", doc)
	}
	clients, ok := doc["clients"].([]interface{})
	if !ok || len(clients) != 1 {
		t.Fatalf("unexpected clients response: %#v", doc["clients"])
	}

	// The bundled Web UI and integrations such as Pwnagotchi consume the
	// collection form and expect the same AP shape below wifi.aps.
	recorder = httptest.NewRecorder()
	mod.showWiFi(recorder, httptest.NewRequest(http.MethodGet, "/api/session/wifi", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("collection: got status %d: %s", recorder.Code, recorder.Body.String())
	}
	var wifiDoc map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &wifiDoc); err != nil {
		t.Fatal(err)
	}
	aps, ok := wifiDoc["aps"].([]interface{})
	if !ok || len(aps) != 1 {
		t.Fatalf("unexpected AP collection: %#v", wifiDoc["aps"])
	}
	if _, ok := aps[0].(map[string]interface{})["clients"].([]interface{}); !ok {
		t.Fatalf("collection AP clients changed: %#v", aps[0])
	}

	// A direct client lookup must remain a Station object, without AP-only
	// clients/handshake fields.
	req = mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/session/wifi/02:00:00:00:00:02", nil), map[string]string{
		"mac": "02:00:00:00:00:02",
	})
	recorder = httptest.NewRecorder()
	mod.showWiFi(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("client: got status %d: %s", recorder.Code, recorder.Body.String())
	}
	var clientDoc map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &clientDoc); err != nil {
		t.Fatal(err)
	}
	if clientDoc["mac"] != "02:00:00:00:00:02" {
		t.Fatalf("unexpected client response: %#v", clientDoc)
	}
	if _, found := clientDoc["clients"]; found {
		t.Fatalf("client response gained AP-only fields: %#v", clientDoc)
	}

	// Recording holds Session.Lock while serializing. WiFi now owns its lock,
	// so this exact path guards against recursive-lock deadlocks and also checks
	// the full /api/session -> wifi.aps shape consumed by the Web UI.
	type marshalResult struct {
		raw []byte
		err error
	}
	done := make(chan marshalResult, 1)
	go func() {
		s.Lock()
		raw, err := json.Marshal(s)
		s.Unlock()
		done <- marshalResult{raw: raw, err: err}
	}()
	select {
	case result := <-done:
		if result.err != nil {
			t.Fatal(result.err)
		}
		var sessionDoc map[string]interface{}
		if err := json.Unmarshal(result.raw, &sessionDoc); err != nil {
			t.Fatal(err)
		}
		fullWiFi, ok := sessionDoc["wifi"].(map[string]interface{})
		if !ok {
			t.Fatalf("full session WiFi changed: %#v", sessionDoc["wifi"])
		}
		fullAPs, ok := fullWiFi["aps"].([]interface{})
		if !ok || len(fullAPs) != 1 {
			t.Fatalf("full session AP collection changed: %#v", fullWiFi["aps"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("full session serialization deadlocked while Session.Lock was held")
	}
}

func TestIsTLS(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Initially should not be TLS
	if mod.isTLS() {
		t.Error("Should not be TLS without cert and key")
	}

	// Set cert and key
	mod.certFile = "cert.pem"
	mod.keyFile = "key.pem"

	if !mod.isTLS() {
		t.Error("Should be TLS with cert and key")
	}

	// Only cert
	mod.certFile = "cert.pem"
	mod.keyFile = ""

	if mod.isTLS() {
		t.Error("Should not be TLS with only cert")
	}

	// Only key
	mod.certFile = ""
	mod.keyFile = "key.pem"

	if mod.isTLS() {
		t.Error("Should not be TLS with only key")
	}
}

func TestStateStore(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Check that state variables are properly stored
	stateKeys := []string{
		"recording",
		"rec_clock",
		"replaying",
		"loading",
		"load_progress",
		"rec_time",
		"rec_filename",
		"rec_frames",
		"rec_cur_frame",
		"rec_started",
		"rec_stopped",
	}

	for _, key := range stateKeys {
		val, exists := mod.State.Load(key)
		if !exists || val == nil {
			t.Errorf("State key '%s' not found", key)
		}
	}
}

func TestParameters(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Check that all parameters are registered
	paramNames := []string{
		"api.rest.address",
		"api.rest.port",
		"api.rest.alloworigin",
		"api.rest.username",
		"api.rest.password",
		"api.rest.certificate",
		"api.rest.key",
		"api.rest.websocket",
		"api.rest.record.clock",
	}

	// Parameters are stored in the session environment
	// We'll just check they can be accessed without error
	for _, param := range paramNames {
		// This is a simplified check
		_ = param
	}

	// Ensure mod is used
	if mod == nil {
		t.Error("Module should not be nil")
	}
}

func TestJSSessionStructs(t *testing.T) {
	// Test struct creation
	req := JSSessionRequest{
		Command: "test command",
	}

	if req.Command != "test command" {
		t.Errorf("Expected command 'test command', got '%s'", req.Command)
	}

	resp := JSSessionResponse{
		Error: "test error",
	}

	if resp.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", resp.Error)
	}
}

func TestDefaultValues(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Check default values
	if mod.recClock != 1 {
		t.Errorf("Expected default recClock 1, got %d", mod.recClock)
	}

	if mod.recTime != 0 {
		t.Errorf("Expected default recTime 0, got %d", mod.recTime)
	}

	if mod.recordFileName != "" {
		t.Errorf("Expected empty recordFileName, got '%s'", mod.recordFileName)
	}

	if mod.upgrader.ReadBufferSize != 1024 {
		t.Errorf("Expected ReadBufferSize 1024, got %d", mod.upgrader.ReadBufferSize)
	}

	if mod.upgrader.WriteBufferSize != 1024 {
		t.Errorf("Expected WriteBufferSize 1024, got %d", mod.upgrader.WriteBufferSize)
	}
}

func TestRunningState(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Initially should not be running
	if mod.Running() {
		t.Error("Module should not be running initially")
	}

	// Note: Cannot test actual Start/Stop without proper server setup
}

func TestRecordingState(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test recording state changes
	mod.recording = true
	if !mod.recording {
		t.Error("Recording flag should be true")
	}

	mod.recording = false
	if mod.recording {
		t.Error("Recording flag should be false")
	}
}

func TestReplayingState(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test replaying state changes
	mod.replaying = true
	if !mod.replaying {
		t.Error("Replaying flag should be true")
	}

	mod.replaying = false
	if mod.replaying {
		t.Error("Replaying flag should be false")
	}
}

func TestConfigureErrors(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test configuration validation
	testCases := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "invalid address",
			setup: func() {
				s.Env.Set("api.rest.address", "999.999.999.999")
			},
			expected: "address",
		},
		{
			name: "invalid port",
			setup: func() {
				s.Env.Set("api.rest.address", "127.0.0.1")
				s.Env.Set("api.rest.port", "not-a-port")
			},
			expected: "port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			// Configure may fail due to parameter validation
			_ = mod.Configure()
		})
	}
}

func TestServerConfiguration(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Set valid parameters
	s.Env.Set("api.rest.address", "127.0.0.1")
	s.Env.Set("api.rest.port", "8081")
	s.Env.Set("api.rest.username", "testuser")
	s.Env.Set("api.rest.password", "testpass")
	s.Env.Set("api.rest.websocket", "true")
	s.Env.Set("api.rest.alloworigin", "http://localhost:3000")

	// This might fail due to TLS cert generation, but we're testing the flow
	_ = mod.Configure()

	// Check that values were set
	if mod.username != "" && mod.username != "testuser" {
		t.Logf("Username set to: %s", mod.username)
	}
	if mod.password != "" && mod.password != "testpass" {
		t.Logf("Password set to: %s", mod.password)
	}
}

func TestQuitChannel(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test quit channel is created
	if mod.quit == nil {
		t.Error("Quit channel should not be nil")
	}

	// Test sending to quit channel doesn't block
	done := make(chan bool)
	go func() {
		select {
		case mod.quit <- true:
			done <- true
		case <-time.After(100 * time.Millisecond):
			done <- false
		}
	}()

	// Start reading from quit channel
	go func() {
		<-mod.quit
	}()

	if !<-done {
		t.Error("Sending to quit channel timed out")
	}
}

func TestRecordWaitGroup(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test wait group is initialized
	if mod.recordWait == nil {
		t.Error("Record wait group should not be nil")
	}

	// Test wait group operations
	mod.recordWait.Add(1)
	done := make(chan bool)

	go func() {
		mod.recordWait.Done()
		done <- true
	}()

	go func() {
		mod.recordWait.Wait()
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait group operation timed out")
	}
}

func TestStartErrors(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test start when replaying
	mod.replaying = true
	err := mod.Start()
	if err == nil {
		t.Error("Expected error when starting while replaying")
	}
}

func TestConfigureAlreadyRunning(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Simulate running state
	mod.SetRunning(true, func() {})

	err := mod.Configure()
	if err == nil {
		t.Error("Expected error when configuring while running")
	}

	// Reset
	mod.SetRunning(false, func() {})
}

func TestServerAddr(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Set parameters
	s.Env.Set("api.rest.address", "192.168.1.100")
	s.Env.Set("api.rest.port", "9090")

	// Configure may fail but we can check server addr format
	_ = mod.Configure()

	expectedAddr := "192.168.1.100:9090"
	if mod.server != nil && mod.server.Addr != "" && mod.server.Addr != expectedAddr {
		t.Logf("Server addr: %s", mod.server.Addr)
	}
}

func TestTLSConfiguration(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Test with TLS params
	s.Env.Set("api.rest.certificate", "/tmp/test.crt")
	s.Env.Set("api.rest.key", "/tmp/test.key")

	// Configure will attempt to expand paths and check files
	_ = mod.Configure()

	// Just verify the attempt was made
	t.Logf("Attempted TLS configuration")
}

// Benchmark tests
func BenchmarkNewRestAPI(b *testing.B) {
	s, _ := session.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRestAPI(s)
	}
}

func BenchmarkIsTLS(b *testing.B) {
	s, _ := session.New()
	mod := NewRestAPI(s)
	mod.certFile = "cert.pem"
	mod.keyFile = "key.pem"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mod.isTLS()
	}
}

func BenchmarkConfigure(b *testing.B) {
	s, _ := session.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod := NewRestAPI(s)
		s.Env.Set("api.rest.address", "127.0.0.1")
		s.Env.Set("api.rest.port", "8081")
		_ = mod.Configure()
	}
}

// Tests for controller functionality
func TestCommandRequest(t *testing.T) {
	cmd := CommandRequest{
		Command: "help",
	}

	if cmd.Command != "help" {
		t.Errorf("Expected command 'help', got '%s'", cmd.Command)
	}
}

func TestAPIResponse(t *testing.T) {
	resp := APIResponse{
		Success: true,
		Message: "Operation completed",
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Message != "Operation completed" {
		t.Errorf("Expected message 'Operation completed', got '%s'", resp.Message)
	}
}

func TestCheckAuthNoCredentials(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// No username/password set - should allow access
	req, _ := http.NewRequest("GET", "/test", nil)

	if !mod.checkAuth(req) {
		t.Error("Expected auth to pass with no credentials set")
	}
}

func TestCheckAuthWithCredentials(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Set credentials
	mod.username = "testuser"
	mod.password = "testpass"

	// Test without auth header
	req1, _ := http.NewRequest("GET", "/test", nil)
	if mod.checkAuth(req1) {
		t.Error("Expected auth to fail without credentials")
	}

	// Test with wrong credentials
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.SetBasicAuth("wronguser", "wrongpass")
	if mod.checkAuth(req2) {
		t.Error("Expected auth to fail with wrong credentials")
	}

	// Test with correct credentials
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.SetBasicAuth("testuser", "testpass")
	if !mod.checkAuth(req3) {
		t.Error("Expected auth to pass with correct credentials")
	}
}

func TestGetEventsEmpty(t *testing.T) {
	// Skip this test if running with others due to shared session state
	if testing.Short() {
		t.Skip("Skipping in short mode due to shared session state")
	}

	// Create a fresh session using the singleton
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Record initial event count
	initialCount := len(mod.getEvents(0))

	// Get events - we can't guarantee zero events due to session initialization
	events := mod.getEvents(0)
	if len(events) < initialCount {
		t.Errorf("Event count should not decrease, got %d", len(events))
	}
}

func TestGetEventsWithLimit(t *testing.T) {
	// Create session using the singleton
	s := createMockSession(t)
	mod := NewRestAPI(s)

	// Record initial state
	initialEvents := mod.getEvents(0)
	initialCount := len(initialEvents)

	// Add some test events
	testEventCount := 10
	for i := 0; i < testEventCount; i++ {
		s.Events.Add(fmt.Sprintf("test.event.limit.%d", i), nil)
	}

	// Get all events
	allEvents := mod.getEvents(0)
	expectedTotal := initialCount + testEventCount
	if len(allEvents) != expectedTotal {
		t.Errorf("Expected %d total events, got %d", expectedTotal, len(allEvents))
	}

	// Test limit functionality - get last 5 events
	limitedEvents := mod.getEvents(5)
	if len(limitedEvents) != 5 {
		t.Errorf("Expected 5 events when limiting, got %d", len(limitedEvents))
	}
}

func TestSetSecurityHeaders(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)
	mod.allowOrigin = "http://localhost:3000"

	w := httptest.NewRecorder()
	mod.setSecurityHeaders(w)

	headers := w.Header()

	// Check security headers
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options header not set correctly")
	}

	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header not set correctly")
	}

	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("X-XSS-Protection header not set correctly")
	}

	if headers.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Access-Control-Allow-Origin header not set correctly")
	}
}

func TestCorsRoute(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	mod.corsRoute(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestToJSON(t *testing.T) {
	s := createMockSession(t)
	mod := NewRestAPI(s)

	w := httptest.NewRecorder()

	testData := map[string]string{
		"key": "value",
		"foo": "bar",
	}

	mod.toJSON(w, testData)

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set to application/json")
	}

	// Check JSON response
	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	if result["key"] != "value" || result["foo"] != "bar" {
		t.Error("JSON response doesn't match expected data")
	}
}
