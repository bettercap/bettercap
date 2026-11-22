package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/bettercap/bettercap/v2/session"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// empty is used as the output type for tools that only return textual content.
type empty struct{}

// textResult builds a tool result carrying a single text block.
func textResult(text string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: text}},
	}
}

// errorResult builds a tool-level error result (IsError) so the agent can see
// the failure and self-correct, rather than receiving a protocol error.
func errorResult(text string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		IsError: true,
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: text}},
	}
}

// runAndCapture executes the given (possibly ';'-separated) commands through the
// session dispatcher while capturing everything they print, then returns the
// ANSI-stripped output. The global os.Stdout swap is serialized with cmdLock so
// concurrent tool calls cannot interleave their captures.
func (mod *MCPModule) runAndCapture(command string) (string, error) {
	mod.cmdLock.Lock()
	defer mod.cmdLock.Unlock()

	rescueStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = writer

	var runErr error
	for _, cmd := range session.ParseCommands(command) {
		if runErr = mod.Session.Run(cmd); runErr != nil {
			break
		}
	}

	os.Stdout = rescueStdout
	writer.Close()
	out, _ := io.ReadAll(reader)
	reader.Close()

	return ansiEscapeRegex.ReplaceAllString(string(out), ""), runErr
}

// --- run_command ---

type runCommandInput struct {
	Command string `json:"command" jsonschema:"one or more bettercap commands separated by ';', e.g. 'net.probe on' or 'set wifi.show.sort rssi; wifi.show'"`
}

func (mod *MCPModule) toolRunCommand(ctx context.Context, req *mcpsdk.CallToolRequest, in runCommandInput) (*mcpsdk.CallToolResult, empty, error) {
	out, err := mod.runAndCapture(in.Command)
	if err != nil {
		msg := fmt.Sprintf("error: %s", err)
		if out != "" {
			msg = fmt.Sprintf("%s\n\noutput before error:\n%s", msg, out)
		}
		return errorResult(msg), empty{}, nil
	}
	if out == "" {
		out = "(command produced no output)"
	}
	return textResult(out), empty{}, nil
}

// --- list_commands ---

type coreCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (mod *MCPModule) toolListCommands(ctx context.Context, req *mcpsdk.CallToolRequest, in empty) (*mcpsdk.CallToolResult, empty, error) {
	core := make([]coreCommand, 0, len(mod.Session.CoreHandlers))
	for _, h := range mod.Session.CoreHandlers {
		core = append(core, coreCommand{Name: h.Name, Description: h.Description})
	}

	// mod.Session.Modules (ModuleList) marshals each module with its name,
	// description, running state, handlers and parameters.
	payload := struct {
		CoreCommands []coreCommand      `json:"core_commands"`
		Modules      session.ModuleList `json:"modules"`
	}{
		CoreCommands: core,
		Modules:      mod.Session.Modules,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("error marshaling commands: %s", err)), empty{}, nil
	}
	return textResult(string(data)), empty{}, nil
}

// --- help ---

type helpInput struct {
	Module string `json:"module,omitempty" jsonschema:"optional module name (e.g. 'wifi'); leave empty for the general help"`
}

func (mod *MCPModule) toolHelp(ctx context.Context, req *mcpsdk.CallToolRequest, in helpInput) (*mcpsdk.CallToolResult, empty, error) {
	cmd := "help"
	if in.Module != "" {
		cmd = "help " + in.Module
	}
	out, err := mod.runAndCapture(cmd)
	if err != nil {
		return errorResult(fmt.Sprintf("error: %s", err)), empty{}, nil
	}
	return textResult(out), empty{}, nil
}

// --- get_session ---

func (mod *MCPModule) toolGetSession(ctx context.Context, req *mcpsdk.CallToolRequest, in empty) (*mcpsdk.CallToolResult, empty, error) {
	mod.Session.Lock()
	data, err := json.MarshalIndent(mod.Session, "", "  ")
	mod.Session.Unlock()
	if err != nil {
		return errorResult(fmt.Sprintf("error marshaling session: %s", err)), empty{}, nil
	}
	return textResult(string(data)), empty{}, nil
}

// --- read_events ---

type readEventsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"maximum number of most-recent events to return; 0 or less returns all"`
}

func (mod *MCPModule) toolReadEvents(ctx context.Context, req *mcpsdk.CallToolRequest, in readEventsInput) (*mcpsdk.CallToolResult, empty, error) {
	all := mod.Session.Events.Sorted()
	events := make([]session.Event, 0, len(all))
	for _, e := range all {
		if !mod.Session.EventsIgnoreList.Ignored(e) {
			events = append(events, e)
		}
	}

	if in.Limit > 0 && in.Limit < len(events) {
		events = events[len(events)-in.Limit:]
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("error marshaling events: %s", err)), empty{}, nil
	}
	return textResult(string(data)), empty{}, nil
}
