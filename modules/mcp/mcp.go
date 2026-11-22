package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bettercap/bettercap/v2/session"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPModule exposes the bettercap command session to AI agents over the Model
// Context Protocol (MCP). Instead of wrapping every module in a dedicated tool,
// it gives agents direct access to the session command dispatcher, so any
// command provided by any current or future module is automatically reachable.
type MCPModule struct {
	session.SessionModule

	transport string
	address   string
	port      int
	token     string

	server    *mcpsdk.Server
	http      *http.Server
	runCancel context.CancelFunc

	// cmdLock serializes command execution because capturing a command's output
	// swaps the process-global os.Stdout; two concurrent captures would interleave.
	cmdLock sync.Mutex
	// origStdout is the real stdout captured before any output-capture swap, used
	// as the MCP JSON-RPC stream for the stdio transport.
	origStdout *os.File
}

func NewMCPModule(s *session.Session) *MCPModule {
	mod := &MCPModule{
		SessionModule: session.NewSessionModule("mcp", s),
		transport:     "http",
		address:       "127.0.0.1",
		port:          8083,
	}

	mod.AddParam(session.NewStringParameter("mcp.transport",
		"http",
		"^(http|stdio)$",
		"MCP transport to expose, either 'http' (Streamable HTTP) or 'stdio' (newline delimited JSON over stdin/stdout, for headless/parent-spawned use)."))

	mod.AddParam(session.NewStringParameter("mcp.address",
		"127.0.0.1",
		session.IPv4Validator,
		"Address to bind the MCP Streamable HTTP server to."))

	mod.AddParam(session.NewIntParameter("mcp.port",
		"8083",
		"Port to bind the MCP Streamable HTTP server to."))

	mod.AddParam(session.NewStringParameter("mcp.token",
		"",
		"",
		"If set, the MCP Streamable HTTP endpoint requires an 'Authorization: Bearer <token>' header."))

	mod.AddHandler(session.NewModuleHandler("mcp on", "",
		"Start the MCP server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("mcp off", "",
		"Stop the MCP server.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *MCPModule) Name() string {
	return "mcp"
}

func (mod *MCPModule) Description() string {
	return "Expose the bettercap interactive session to AI agents over the Model Context Protocol (MCP)."
}

func (mod *MCPModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *MCPModule) Configure() error {
	var err error

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.transport = mod.StringParam("mcp.transport"); err != nil {
		return err
	} else if err, mod.address = mod.StringParam("mcp.address"); err != nil {
		return err
	} else if err, mod.port = mod.IntParam("mcp.port"); err != nil {
		return err
	} else if err, mod.token = mod.StringParam("mcp.token"); err != nil {
		return err
	}

	// build the MCP server and register the tools once
	mod.server = mod.buildServer()

	if mod.transport == "http" {
		handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
			return mod.server
		}, nil)

		mux := http.NewServeMux()
		mux.Handle("/mcp", mod.withAuth(handler))

		mod.http = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", mod.address, mod.port),
			Handler: mux,
		}
	}

	return nil
}

func (mod *MCPModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	if mod.transport == "http" {
		return mod.SetRunning(true, func() {
			defer mod.SetRunning(false, nil)

			mod.Info("MCP server (streamable http) starting on http://%s/mcp", mod.http.Addr)
			if mod.token == "" {
				mod.Warning("mcp.token is empty, the MCP endpoint has no authentication.")
			}

			err := mod.http.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				mod.Error("MCP http server failed: %v", err)
			}
		})
	}

	// The stdio transport speaks newline-delimited JSON-RPC over stdin/stdout, so
	// that channel must carry nothing else. Reserve the real stdout for the
	// protocol and point everything bettercap prints at stderr for the duration,
	// so logs, events and captured command output never corrupt the stream.
	mod.cmdLock.Lock()
	mod.origStdout = os.Stdout
	os.Stdout = os.Stderr
	mod.cmdLock.Unlock()

	return mod.SetRunning(true, func() {
		defer mod.SetRunning(false, nil)

		ctx, cancel := context.WithCancel(context.Background())
		mod.runCancel = cancel

		mod.Info("MCP server (stdio) starting")
		if err := mod.server.Run(ctx, mod.stdioTransport()); err != nil && ctx.Err() == nil {
			mod.Error("MCP stdio server failed: %v", err)
		}
	})
}

func (mod *MCPModule) Stop() error {
	return mod.SetRunning(false, func() {
		if mod.http != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			mod.http.Shutdown(ctx)
			mod.http = nil
		}
		if mod.runCancel != nil {
			mod.runCancel()
			mod.runCancel = nil
		}
		// restore stdout if the stdio transport had redirected it
		mod.cmdLock.Lock()
		if mod.origStdout != nil {
			os.Stdout = mod.origStdout
			mod.origStdout = nil
		}
		mod.cmdLock.Unlock()
	})
}
