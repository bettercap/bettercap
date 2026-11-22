package mcp

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bettercap/bettercap/v2/session"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestNewMCPModule(t *testing.T) {
	mod := NewMCPModule(createMockSession(t))

	if mod == nil {
		t.Fatal("NewMCPModule returned nil")
	}
	if mod.Name() != "mcp" {
		t.Errorf("expected name 'mcp', got '%s'", mod.Name())
	}
	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("unexpected author: %s", mod.Author())
	}
	if mod.Description() == "" {
		t.Error("empty description")
	}

	wantHandlers := map[string]bool{"mcp on": false, "mcp off": false}
	for _, h := range mod.Handlers() {
		if _, ok := wantHandlers[h.Name]; ok {
			wantHandlers[h.Name] = true
		}
	}
	for name, found := range wantHandlers {
		if !found {
			t.Errorf("missing handler %q", name)
		}
	}

	params := mod.Parameters()
	for _, name := range []string{"mcp.transport", "mcp.address", "mcp.port", "mcp.token"} {
		if _, ok := params[name]; !ok {
			t.Errorf("missing parameter %q", name)
		}
	}
}

func TestRunAndCaptureRestoresStdout(t *testing.T) {
	mod := NewMCPModule(createMockSession(t))

	real := os.Stdout

	// successful command (core 'help' handler is always available)
	out, err := mod.runAndCapture("help")
	if err != nil {
		t.Fatalf("unexpected error running help: %v", err)
	}
	if os.Stdout != real {
		t.Fatal("os.Stdout was not restored after a successful command")
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty help output")
	}

	// failing command must still restore stdout and surface the error
	_, err = mod.runAndCapture("this.command.does.not.exist")
	if err == nil {
		t.Error("expected an error for an unknown command")
	}
	if os.Stdout != real {
		t.Fatal("os.Stdout was not restored after a failing command")
	}
}

func TestToolRunCommand(t *testing.T) {
	mod := NewMCPModule(createMockSession(t))

	res, _, err := mod.toolRunCommand(context.Background(), nil, runCommandInput{Command: "help"})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", firstText(res))
	}
	if strings.TrimSpace(firstText(res)) == "" {
		t.Error("expected non-empty output")
	}

	// an invalid command should come back as a tool-level error, not a Go error
	res, _, err = mod.toolRunCommand(context.Background(), nil, runCommandInput{Command: "nope.nope"})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !res.IsError {
		t.Error("expected IsError=true for an invalid command")
	}
}

func TestToolListCommandsAndHelp(t *testing.T) {
	mod := NewMCPModule(createMockSession(t))

	res, _, err := mod.toolListCommands(context.Background(), nil, empty{})
	if err != nil {
		t.Fatalf("list_commands error: %v", err)
	}
	text := firstText(res)
	if !strings.Contains(text, "core_commands") || !strings.Contains(text, "modules") {
		t.Errorf("list_commands output missing expected keys: %s", text)
	}

	res, _, err = mod.toolHelp(context.Background(), nil, helpInput{})
	if err != nil {
		t.Fatalf("help error: %v", err)
	}
	if strings.TrimSpace(firstText(res)) == "" {
		t.Error("expected non-empty general help")
	}
}

func TestBuildServer(t *testing.T) {
	mod := NewMCPModule(createMockSession(t))
	if srv := mod.buildServer(); srv == nil {
		t.Fatal("buildServer returned nil")
	}
}

func firstText(res *mcpsdk.CallToolResult) string {
	for _, c := range res.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
