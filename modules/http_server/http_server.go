package http_server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

type HttpServer struct {
	session.SessionModule
	server *http.Server
}

func NewHttpServer(s *session.Session) *HttpServer {
	mod := &HttpServer{
		SessionModule: session.NewSessionModule("http.server", s),
		server:        &http.Server{},
	}

	mod.AddParam(session.NewStringParameter("http.server.path",
		".",
		"",
		"Server folder."))

	mod.AddParam(session.NewStringParameter("http.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the http server to."))

	mod.AddParam(session.NewIntParameter("http.server.port",
		"80",
		"Port to bind the http server to."))

	mod.AddHandler(session.NewModuleHandler("http.server on", "",
		"Start httpd server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("http.server off", "",
		"Stop httpd server.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *HttpServer) Name() string {
	return "http.server"
}

func (mod *HttpServer) Description() string {
	return "A simple HTTP server, to be used to serve files and scripts across the network."
}

func (mod *HttpServer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *HttpServer) Configure() error {
	var err error
	var path string
	var address string
	var port int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if err, path = mod.StringParam("http.server.path"); err != nil {
		return err
	}

	router := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(path))

	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mod.Debug("%s %s %s%s", tui.Bold(strings.Split(r.RemoteAddr, ":")[0]), r.Method, r.Host, r.URL.Path)
		if r.URL.Path == "/proxy.pac" || r.URL.Path == "/wpad.dat" {
			w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		}
		fileServer.ServeHTTP(w, r)
	}))

	mod.server.Handler = router

	if err, address = mod.StringParam("http.server.address"); err != nil {
		return err
	}

	if err, port = mod.IntParam("http.server.port"); err != nil {
		return err
	}

	mod.server.Addr = fmt.Sprintf("%s:%d", address, port)

	return nil
}

func (mod *HttpServer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		var err error
		mod.Info("starting on http://%s", mod.server.Addr)
		if err = mod.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mod.Error("%v", err)
			mod.Stop()
		}
	})
}

func (mod *HttpServer) Stop() error {
	return mod.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		mod.server.Shutdown(ctx)
	})
}
