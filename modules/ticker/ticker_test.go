package ticker

import (
	"sync"
	"testing"
	"time"

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
	})
	return testSession
}

func TestNewTicker(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	if mod == nil {
		t.Fatal("NewTicker returned nil")
	}

	if mod.Name() != "ticker" {
		t.Errorf("Expected name 'ticker', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("Unexpected author: %s", mod.Author())
	}

	if mod.Description() == "" {
		t.Error("Empty description")
	}

	// Check parameters exist
	if err, _ := mod.StringParam("ticker.commands"); err != nil {
		t.Error("ticker.commands parameter not found")
	}

	if err, _ := mod.IntParam("ticker.period"); err != nil {
		t.Error("ticker.period parameter not found")
	}

	// Check handlers - only check the main ones since create/destroy have regex patterns
	handlers := []string{"ticker on", "ticker off"}
	for _, handler := range handlers {
		found := false
		for _, h := range mod.Handlers() {
			if h.Name == handler {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Handler '%s' not found", handler)
		}
	}

	// Check that we have handlers for create and destroy (they have regex patterns)
	hasCreate := false
	hasDestroy := false
	for _, h := range mod.Handlers() {
		if h.Name == "ticker.create <name> <period> <commands>" {
			hasCreate = true
		} else if h.Name == "ticker.destroy <name>" {
			hasDestroy = true
		}
	}
	if !hasCreate {
		t.Error("ticker.create handler not found")
	}
	if !hasDestroy {
		t.Error("ticker.destroy handler not found")
	}
}

func TestTickerConfigure(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Test configure before start
	if err := mod.Configure(); err != nil {
		t.Errorf("Configure failed: %v", err)
	}

	// Check main params were set
	if mod.main.Period == 0 {
		t.Error("Period not set")
	}

	if len(mod.main.Commands) == 0 {
		t.Error("Commands not set")
	}

	if !mod.main.Running {
		t.Error("Running flag not set")
	}
}

func TestTickerStartStop(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Set a short period for testing using session environment
	mod.Session.Env.Set("ticker.period", "1")
	mod.Session.Env.Set("ticker.commands", "help")

	// Start ticker
	if err := mod.Start(); err != nil {
		t.Fatalf("Failed to start ticker: %v", err)
	}

	if !mod.Running() {
		t.Error("Ticker should be running")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop ticker
	if err := mod.Stop(); err != nil {
		t.Fatalf("Failed to stop ticker: %v", err)
	}

	if mod.Running() {
		t.Error("Ticker should not be running")
	}

	if mod.main.Running {
		t.Error("Main ticker should not be running")
	}
}

func TestTickerAlreadyStarted(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Start ticker
	if err := mod.Start(); err != nil {
		t.Fatalf("Failed to start ticker: %v", err)
	}

	// Try to configure while running
	if err := mod.Configure(); err == nil {
		t.Error("Configure should fail when already running")
	}

	// Stop ticker
	mod.Stop()
}

func TestTickerNamedOperations(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Create named ticker
	name := "test_ticker"
	if err := mod.createNamed(name, 1, "help"); err != nil {
		t.Fatalf("Failed to create named ticker: %v", err)
	}

	// Check it was created
	if _, found := mod.named[name]; !found {
		t.Error("Named ticker not found in map")
	}

	// Try to create duplicate
	if err := mod.createNamed(name, 1, "help"); err == nil {
		t.Error("Should not allow duplicate named ticker")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Destroy named ticker
	if err := mod.destroyNamed(name); err != nil {
		t.Fatalf("Failed to destroy named ticker: %v", err)
	}

	// Check it was removed
	if _, found := mod.named[name]; found {
		t.Error("Named ticker still in map after destroy")
	}

	// Try to destroy non-existent
	if err := mod.destroyNamed("nonexistent"); err == nil {
		t.Error("Should fail when destroying non-existent ticker")
	}
}

func TestTickerHandlers(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	tests := []struct {
		name    string
		handler string
		regex   string
		args    []string
		wantErr bool
	}{
		{
			name:    "ticker on",
			handler: "ticker on",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "ticker off",
			handler: "ticker off",
			args:    []string{},
			wantErr: true, // ticker off will fail if not running
		},
		{
			name:    "ticker.create valid",
			handler: "ticker.create <name> <period> <commands>",
			args:    []string{"myticker", "2", "help; events.show"},
			wantErr: false,
		},
		{
			name:    "ticker.create invalid period",
			handler: "ticker.create <name> <period> <commands>",
			args:    []string{"myticker", "notanumber", "help"},
			wantErr: true,
		},
		{
			name:    "ticker.destroy",
			handler: "ticker.destroy <name>",
			args:    []string{"myticker"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find the handler
			var handler *session.ModuleHandler
			for _, h := range mod.Handlers() {
				if h.Name == tt.handler {
					handler = &h
					break
				}
			}

			if handler == nil {
				t.Fatalf("Handler '%s' not found", tt.handler)
			}

			// Create ticker if needed for destroy test
			if tt.handler == "ticker.destroy <name>" && len(tt.args) > 0 && tt.args[0] == "myticker" {
				mod.createNamed("myticker", 1, "help")
			}

			// Execute handler
			err := handler.Exec(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler execution error = %v, wantErr %v", err, tt.wantErr)
			}

			// Cleanup
			if tt.handler == "ticker on" || tt.handler == "ticker.create <name> <period> <commands>" {
				mod.Stop()
			}
		})
	}
}

func TestTickerWorker(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Create params for testing
	params := &Params{
		Commands: []string{"help"},
		Period:   100 * time.Millisecond,
		Running:  true,
	}

	// Start worker in goroutine
	done := make(chan bool)
	go func() {
		mod.worker("test", params)
		done <- true
	}()

	// Let it tick at least once
	time.Sleep(150 * time.Millisecond)

	// Stop the worker
	params.Running = false

	// Wait for worker to finish
	select {
	case <-done:
		// Worker finished successfully
	case <-time.After(1 * time.Second):
		t.Error("Worker did not stop in time")
	}
}

func TestTickerParams(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Test setting invalid period
	mod.Session.Env.Set("ticker.period", "invalid")
	if err := mod.Configure(); err == nil {
		t.Error("Configure should fail with invalid period")
	}

	// Test empty commands
	mod.Session.Env.Set("ticker.period", "1")
	mod.Session.Env.Set("ticker.commands", "")
	if err := mod.Configure(); err != nil {
		t.Errorf("Configure should work with empty commands: %v", err)
	}
}

func TestTickerMultipleNamed(t *testing.T) {
	s := createMockSession(t)
	mod := NewTicker(s)

	// Start the ticker first
	if err := mod.Start(); err != nil {
		t.Fatalf("Failed to start ticker: %v", err)
	}

	// Create multiple named tickers
	names := []string{"ticker1", "ticker2", "ticker3"}
	for _, name := range names {
		if err := mod.createNamed(name, 1, "help"); err != nil {
			t.Errorf("Failed to create ticker '%s': %v", name, err)
		}
	}

	// Check all were created
	if len(mod.named) != len(names) {
		t.Errorf("Expected %d named tickers, got %d", len(names), len(mod.named))
	}

	// Stop all via Stop()
	if err := mod.Stop(); err != nil {
		t.Fatalf("Failed to stop: %v", err)
	}

	// Check all were stopped
	for name, params := range mod.named {
		if params.Running {
			t.Errorf("Ticker '%s' still running after Stop()", name)
		}
	}
}

func TestTickEvent(t *testing.T) {
	// Simple test for TickEvent struct
	event := TickEvent{}
	// TickEvent is empty, just ensure it can be created
	_ = event
}

// Benchmark tests
func BenchmarkTickerCreate(b *testing.B) {
	// Use existing session to avoid flag redefinition
	s := testSession
	if s == nil {
		var err error
		s, err = session.New()
		if err != nil {
			b.Fatal(err)
		}
		testSession = s
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod := NewTicker(s)
		_ = mod
	}
}

func BenchmarkTickerStartStop(b *testing.B) {
	// Use existing session to avoid flag redefinition
	s := testSession
	if s == nil {
		var err error
		s, err = session.New()
		if err != nil {
			b.Fatal(err)
		}
		testSession = s
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod := NewTicker(s)
		// Set period parameter
		mod.Session.Env.Set("ticker.period", "1")
		mod.Start()
		mod.Stop()
	}
}
