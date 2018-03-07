package modules

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
	"github.com/bettercap/bettercap/tls"
)

type HttpServer struct {
	session.SessionModule
	server   *http.Server
	certFile string
	keyFile  string
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

	httpd.AddParam(session.NewStringParameter("http.server.certificate",
		"",
		"",
		"TLS certificate file, if not empty will configure this as a HTTPS server (will be auto generated if filled but not existing)."))

	httpd.AddParam(session.NewStringParameter("http.server.key",
		"",
		"",
		"TLS key file, if not empty will configure this as a HTTPS server (will be auto generated if filled but not existing)."))

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

func (httpd *HttpServer) isTLS() bool {
	return httpd.certFile != "" && httpd.keyFile != ""
}

func (httpd *HttpServer) Configure() error {
	var err error
	var path string
	var address string
	var port int
	var certFile string
	var keyFile string

	if httpd.Running() == true {
		return session.ErrAlreadyStarted
	}

	if err, path = httpd.StringParam("http.server.path"); err != nil {
		return err
	}

	router := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(path))

	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("(%s) %s %s %s%s", core.Green("httpd"), core.Bold(strings.Split(r.RemoteAddr, ":")[0]), r.Method, r.Host, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	}))

	httpd.server.Handler = router

	if err, address = httpd.StringParam("http.server.address"); err != nil {
		return err
	}

	if err, port = httpd.IntParam("http.server.port"); err != nil {
		return err
	}

	httpd.server.Addr = fmt.Sprintf("%s:%d", address, port)

	if err, certFile = httpd.StringParam("http.server.certificate"); err != nil {
		return err
	} else if certFile, err = core.ExpandPath(certFile); err != nil {
		return err
	}

	if err, keyFile = httpd.StringParam("http.server.key"); err != nil {
		return err
	} else if keyFile, err = core.ExpandPath(keyFile); err != nil {
		return err
	}

	if certFile != "" && keyFile != "" {
		if core.Exists(certFile) == false || core.Exists(keyFile) == false {
			log.Info("Generating server TLS key to %s", keyFile)
			log.Info("Generating server TLS certificate to %s", certFile)
			if err := tls.Generate(certFile, keyFile); err != nil {
				return err
			}
		} else {
			log.Info("Loading server TLS key from %s", keyFile)
			log.Info("Loading server TLS certificate from %s", certFile)
		}
	}

	httpd.certFile = certFile
	httpd.keyFile = keyFile

	return nil
}

func (httpd *HttpServer) Start() error {
	if err := httpd.Configure(); err != nil {
		return err
	}

	return httpd.SetRunning(true, func() {
		var err error

		if httpd.isTLS() {
			log.Info("HTTPS server starting on https://%s", httpd.server.Addr)
			err = httpd.server.ListenAndServeTLS(httpd.certFile, httpd.keyFile)
		} else {
			log.Info("HTTP server starting on http://%s", httpd.server.Addr)
			err = httpd.server.ListenAndServe()
		}
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
