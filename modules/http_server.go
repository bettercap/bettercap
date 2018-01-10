package modules

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"
)

type HttpServer struct {
	session.SessionModule
	server *http.Server
	path   string
}

func NewHttpServer(s *session.Session) *HttpServer {
	httpd := &HttpServer{
		SessionModule: session.NewSessionModule("http.server", s),
		server:        &http.Server{},
		path:          ".",
	}

	httpd.AddParam(session.NewStringParameter("http.server.path",
		".",
		"",
		"Server folder."))

	httpd.AddParam(session.NewStringParameter("http.server.address",
		"<interface address>",
		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`,
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
	return "A simple HTTP server, to be used to serve files and scripts accross the network."
}

func (httpd *HttpServer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (httpd *HttpServer) OnSessionStarted(s *session.Session) {
	// refresh the address after session has been created
	s.Env.Set("http.server.address", s.Interface.IpAddress)
}

func (httpd *HttpServer) OnSessionEnded(s *session.Session) {
	if httpd.Running() {
		httpd.Stop()
	}
}

func (httpd *HttpServer) configure() error {
	var address string
	var port int

	if err, v := httpd.Param("http.server.path").Get(httpd.Session); err != nil {
		return err
	} else {
		httpd.path = v.(string)
	}

	http.Handle("/", http.FileServer(http.Dir(httpd.path)))

	if err, v := httpd.Param("http.server.address").Get(httpd.Session); err != nil {
		return err
	} else {
		address = v.(string)
	}

	if err, v := httpd.Param("http.server.port").Get(httpd.Session); err != nil {
		return err
	} else {
		port = v.(int)
	}

	httpd.server.Addr = fmt.Sprintf("%s:%d", address, port)

	return nil
}

func (httpd *HttpServer) Start() error {
	if err := httpd.configure(); err != nil {
		return err
	}

	if httpd.Running() == false {
		httpd.SetRunning(true)
		go func() {
			log.Info("httpd server starting on http://%s", httpd.server.Addr)
			err := httpd.server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		return nil
	}

	return fmt.Errorf("httpd server already started.")
}

func (httpd *HttpServer) Stop() error {
	if httpd.Running() == true {
		httpd.SetRunning(false)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		return httpd.server.Shutdown(ctx)
	} else {
		return fmt.Errorf("httpd server already stopped.")
	}
}
