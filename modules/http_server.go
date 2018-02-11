package modules

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"
)

type HttpServer struct {
	session.SessionModule
	server *http.Server
}

func NewHttpServer(s *session.Session) *HttpServer {
	httpd := &HttpServer{
		SessionModule: session.NewSessionModule("http.server", s),
		server:        &http.Server{},
	}

	httpd.AddParam(session.NewStringParameter("http.server.path",
		".",
		"",
		"Server folder."))

	httpd.AddParam(session.NewStringParameter("http.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the http server to."))

	httpd.AddParam(session.NewIntParameter("http.server.port",
		"80",
		"Port to bind the http server to."))

	httpd.AddHandler(session.NewModuleHandler("http.server on", "",
		"Start httpd server.",
		func(args []string) error {
			return httpd.Start()
		}))

	httpd.AddHandler(session.NewModuleHandler("http.server off", "",
		"Stop httpd server.",
		func(args []string) error {
			return httpd.Stop()
		}))

	return httpd
}

func (httpd *HttpServer) Name() string {
	return "http.server"
}

func (httpd *HttpServer) Description() string {
	return "A simple HTTP server, to be used to serve files and scripts across the network."
}

func (httpd *HttpServer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func wrapHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("(%s) %s %s %s%s", core.Green("httpd"), core.Bold(strings.Split(r.RemoteAddr, ":")[0]), r.Method, r.Host, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

func (httpd *HttpServer) Configure() error {
	var err error
	var path string
	var address string
	var port int

	if err, path = httpd.StringParam("http.server.path"); err != nil {
		return err
	}

	http.Handle("/", wrapHandler(http.FileServer(http.Dir(path))))

	if err, address = httpd.StringParam("http.server.address"); err != nil {
		return err
	}

	if err, port = httpd.IntParam("http.server.port"); err != nil {
		return err
	}

	httpd.server.Addr = fmt.Sprintf("%s:%d", address, port)

	return nil
}

func (httpd *HttpServer) Start() error {
	if httpd.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := httpd.Configure(); err != nil {
		return err
	}

	return httpd.SetRunning(true, func() {
		log.Info("httpd server starting on http://%s", httpd.server.Addr)
		err := httpd.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})
}

func (httpd *HttpServer) Stop() error {
	return httpd.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		httpd.server.Shutdown(ctx)
	})
}
