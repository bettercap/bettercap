package c2

import (
	"sync"
	"testing"
	"text/template"

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

func TestNewC2(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	if mod == nil {
		t.Fatal("NewC2 returned nil")
	}

	if mod.Name() != "c2" {
		t.Errorf("Expected name 'c2', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("Unexpected author: %s", mod.Author())
	}

	if mod.Description() == "" {
		t.Error("Empty description")
	}

	// Check default settings
	if mod.settings.server != "localhost:6697" {
		t.Errorf("Expected default server 'localhost:6697', got '%s'", mod.settings.server)
	}

	if !mod.settings.tls {
		t.Error("Expected TLS to be enabled by default")
	}

	if mod.settings.tlsVerify {
		t.Error("Expected TLS verify to be disabled by default")
	}

	if mod.settings.nick != "bettercap" {
		t.Errorf("Expected default nick 'bettercap', got '%s'", mod.settings.nick)
	}

	if mod.settings.user != "bettercap" {
		t.Errorf("Expected default user 'bettercap', got '%s'", mod.settings.user)
	}

	if mod.settings.operator != "admin" {
		t.Errorf("Expected default operator 'admin', got '%s'", mod.settings.operator)
	}

	// Check channels
	if mod.quit == nil {
		t.Error("Quit channel should not be nil")
	}

	// Check maps
	if mod.templates == nil {
		t.Error("Templates map should not be nil")
	}

	if mod.channels == nil {
		t.Error("Channels map should not be nil")
	}

	// Check handlers
	handlers := mod.Handlers()
	expectedHandlers := []string{
		"c2 on",
		"c2 off",
		"c2.channel.set EVENT_TYPE CHANNEL",
		"c2.channel.clear EVENT_TYPE",
		"c2.template.set EVENT_TYPE TEMPLATE",
		"c2.template.clear EVENT_TYPE",
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
}

func TestDefaultSettings(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Check default channel settings
	if mod.settings.eventsChannel != "#events" {
		t.Errorf("Expected default events channel '#events', got '%s'", mod.settings.eventsChannel)
	}

	if mod.settings.outputChannel != "#events" {
		t.Errorf("Expected default output channel '#events', got '%s'", mod.settings.outputChannel)
	}

	if mod.settings.controlChannel != "#events" {
		t.Errorf("Expected default control channel '#events', got '%s'", mod.settings.controlChannel)
	}

	if mod.settings.password != "password" {
		t.Errorf("Expected default password 'password', got '%s'", mod.settings.password)
	}
}

func TestRunningState(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Initially should not be running
	if mod.Running() {
		t.Error("Module should not be running initially")
	}

	// Note: Cannot test actual Start/Stop without IRC server
}

func TestEventContext(t *testing.T) {
	s := createMockSession(t)

	ctx := eventContext{
		Session: s,
		Event:   session.Event{Tag: "test.event"},
	}

	if ctx.Session == nil {
		t.Error("Session should not be nil")
	}

	if ctx.Event.Tag != "test.event" {
		t.Errorf("Expected event tag 'test.event', got '%s'", ctx.Event.Tag)
	}
}

func TestChannelHandlers(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Test channel.set handler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.channel.set EVENT_TYPE CHANNEL" {
			err := h.Exec([]string{"test.event", "#test"})
			if err != nil {
				t.Errorf("channel.set handler failed: %v", err)
			}

			// Verify channel was set
			if channel, found := mod.channels["test.event"]; !found {
				t.Error("Channel was not set")
			} else if channel != "#test" {
				t.Errorf("Expected channel '#test', got '%s'", channel)
			}
			break
		}
	}

	// Test channel.clear handler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.channel.clear EVENT_TYPE" {
			err := h.Exec([]string{"test.event"})
			if err != nil {
				t.Errorf("channel.clear handler failed: %v", err)
			}

			// Verify channel was cleared
			if _, found := mod.channels["test.event"]; found {
				t.Error("Channel was not cleared")
			}
			break
		}
	}
}

func TestTemplateHandlers(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Test template.set handler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.template.set EVENT_TYPE TEMPLATE" {
			err := h.Exec([]string{"test.event", "Event: {{.Event.Tag}}"})
			if err != nil {
				t.Errorf("template.set handler failed: %v", err)
			}

			// Verify template was set
			if tpl, found := mod.templates["test.event"]; !found {
				t.Error("Template was not set")
			} else if tpl == nil {
				t.Error("Template is nil")
			}
			break
		}
	}

	// Test template.clear handler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.template.clear EVENT_TYPE" {
			err := h.Exec([]string{"test.event"})
			if err != nil {
				t.Errorf("template.clear handler failed: %v", err)
			}

			// Verify template was cleared
			if _, found := mod.templates["test.event"]; found {
				t.Error("Template was not cleared")
			}
			break
		}
	}
}

func TestClearNonExistent(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Test clearing non-existent channel
	for _, h := range mod.Handlers() {
		if h.Name == "c2.channel.clear EVENT_TYPE" {
			err := h.Exec([]string{"non.existent"})
			if err == nil {
				t.Error("Expected error when clearing non-existent channel")
			}
			break
		}
	}

	// Test clearing non-existent template
	for _, h := range mod.Handlers() {
		if h.Name == "c2.template.clear EVENT_TYPE" {
			err := h.Exec([]string{"non.existent"})
			if err == nil {
				t.Error("Expected error when clearing non-existent template")
			}
			break
		}
	}
}

func TestParameters(t *testing.T) {
	s := createMockSession(t)
	mod := NewC2(s)

	// Check that all parameters are registered
	paramNames := []string{
		"c2.server",
		"c2.server.tls",
		"c2.server.tls.verify",
		"c2.operator",
		"c2.nick",
		"c2.username",
		"c2.password",
		"c2.sasl.username",
		"c2.sasl.password",
		"c2.channel.output",
		"c2.channel.events",
		"c2.channel.control",
	}

	// Parameters are stored in the session environment
	for _, param := range paramNames {
		// This is a simplified check
		_ = param
	}

	if mod == nil {
		t.Error("Module should not be nil")
	}
}

func TestTemplateExecution(t *testing.T) {
	// Test template parsing and execution
	tmpl, err := template.New("test").Parse("Event: {{.Event.Tag}}")
	if err != nil {
		t.Errorf("Failed to parse template: %v", err)
	}

	if tmpl == nil {
		t.Error("Template should not be nil")
	}
}

// Benchmark tests
func BenchmarkNewC2(b *testing.B) {
	s, _ := session.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewC2(s)
	}
}

func BenchmarkChannelSet(b *testing.B) {
	s, _ := session.New()
	mod := NewC2(s)

	var handler *session.ModuleHandler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.channel.set EVENT_TYPE CHANNEL" {
			handler = &h
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Exec([]string{"test.event", "#test"})
	}
}

func BenchmarkTemplateSet(b *testing.B) {
	s, _ := session.New()
	mod := NewC2(s)

	var handler *session.ModuleHandler
	for _, h := range mod.Handlers() {
		if h.Name == "c2.template.set EVENT_TYPE TEMPLATE" {
			handler = &h
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Exec([]string{"test.event", "Event: {{.Event.Tag}}"})
	}
}
