package mcp

import (
	"crypto/subtle"
	"io"
	"net/http"
	"os"

	"github.com/bettercap/bettercap/v2/core"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// buildServer creates the MCP server and registers the tools exposed to agents.
// It is transport-agnostic: the same server is served over Streamable HTTP or stdio.
func (mod *MCPModule) buildServer() *mcpsdk.Server {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    core.Name,
		Version: core.Version,
	}, nil)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "run_command",
		Description: "Execute one or more bettercap session commands (separated by ';', exactly as typed at the interactive prompt) and return their textual output. This is the primary way to drive bettercap: any command of any module is available. Use list_commands and help to discover what to run.",
	}, mod.toolRunCommand)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "list_commands",
		Description: "List every available command: the core commands and, for each module, its name, description, running state and command handlers. Reflects all currently loaded modules.",
	}, mod.toolListCommands)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "help",
		Description: "Return bettercap's help text. With no module name, returns the general help (core commands and modules). With a module name, returns that module's commands and parameters.",
	}, mod.toolHelp)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "get_session",
		Description: "Return a JSON snapshot of the current session state: network interface, gateway, discovered LAN/Wi-Fi/BLE hosts, environment variables and module status.",
	}, mod.toolGetSession)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "read_events",
		Description: "Return the most recent session events (log messages and module events) as JSON, newest last. Use the 'limit' argument to cap how many are returned.",
	}, mod.toolReadEvents)

	return srv
}

// withAuth wraps the MCP HTTP handler with optional bearer-token authentication.
// When mcp.token is empty, authentication is disabled (parity with api.rest).
func (mod *MCPModule) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mod.token != "" {
			const prefix = "Bearer "
			auth := r.Header.Get("Authorization")
			if len(auth) <= len(prefix) || subtle.ConstantTimeCompare([]byte(auth[len(prefix):]), []byte(mod.token)) != 1 {
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// stdioTransport builds an IOTransport bound to the real stdin and the stdout
// captured at Start. Using the captured handle keeps the JSON-RPC stream on the
// terminal even while run_command temporarily swaps the global os.Stdout to
// capture command output. The closers are no-ops so the SDK shutdown never
// closes the process stdio.
func (mod *MCPModule) stdioTransport() *mcpsdk.IOTransport {
	out := mod.origStdout
	if out == nil {
		out = os.Stdout
	}
	return &mcpsdk.IOTransport{
		Reader: nopReadCloser{os.Stdin},
		Writer: nopWriteCloser{out},
	}
}

type nopReadCloser struct{ io.Reader }

func (nopReadCloser) Close() error { return nil }

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
