package https_server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
	"github.com/bettercap/bettercap/tls"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
)

type HttpsServer struct {
	session.SessionModule
	server   *http.Server
	certFile string
	keyFile  string
}

func NewHttpsServer(s *session.Session) *HttpsServer {
	httpd := &HttpsServer{
		SessionModule: session.NewSessionModule("https.server", s),
		server:        &http.Server{},
	}

	httpd.AddParam(session.NewStringParameter("https.server.path",
		".",
		"",
		"Server folder."))

	httpd.AddParam(session.NewStringParameter("https.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the http server to."))

	httpd.AddParam(session.NewIntParameter("https.server.port",
		"443",
		"Port to bind the http server to."))

	httpd.AddParam(session.NewStringParameter("https.server.certificate",
		"~/.bettercap-https.cert.pem",
		"",
		"TLS certificate file (will be auto generated if filled but not existing)."))

	httpd.AddParam(session.NewStringParameter("https.server.key",
		"~/.bettercap-https.key.pem",
		"",
		"TLS key file (will be auto generated if filled but not existing)."))

	tls.CertConfigToModule("https.server", &httpd.SessionModule, tls.DefaultLegitConfig)

	httpd.AddHandler(session.NewModuleHandler("https.server on", "",
		"Start https server.",
		func(args []string) error {
			return httpd.Start()
		}))

	httpd.AddHandler(session.NewModuleHandler("https.server off", "",
		"Stop https server.",
		func(args []string) error {
			return httpd.Stop()
		}))

	return httpd
}

func (httpd *HttpsServer) Name() string {
	return "https.server"
}

func (httpd *HttpsServer) Description() string {
	return "A simple HTTPS server, to be used to serve files and scripts across the network."
}

func (httpd *HttpsServer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (httpd *HttpsServer) Configure() error {
	var err error
	var path string
	var address string
	var port int
	var certFile string
	var keyFile string

	if httpd.Running() {
		return session.ErrAlreadyStarted
	}

	if err, path = httpd.StringParam("https.server.path"); err != nil {
		return err
	}

	router := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(path))

	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("(%s) %s %s %s%s", tui.Green("https"), tui.Bold(strings.Split(r.RemoteAddr, ":")[0]), r.Method, r.Host, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	}))

	httpd.server.Handler = router

	if err, address = httpd.StringParam("https.server.address"); err != nil {
		return err
	}

	if err, port = httpd.IntParam("https.server.port"); err != nil {
		return err
	}

	httpd.server.Addr = fmt.Sprintf("%s:%d", address, port)

	if err, certFile = httpd.StringParam("https.server.certificate"); err != nil {
		return err
	} else if certFile, err = fs.Expand(certFile); err != nil {
		return err
	}

	if err, keyFile = httpd.StringParam("https.server.key"); err != nil {
		return err
	} else if keyFile, err = fs.Expand(keyFile); err != nil {
		return err
	}

	if !fs.Exists(certFile) || !fs.Exists(keyFile) {
		err, cfg := tls.CertConfigFromModule("https.server", httpd.SessionModule)
		if err != nil {
			return err
		}

		log.Debug("%+v", cfg)
		log.Info("generating server TLS key to %s", keyFile)
		log.Info("generating server TLS certificate to %s", certFile)
		if err := tls.Generate(cfg, certFile, keyFile); err != nil {
			return err
		}
	} else {
		log.Info("loading server TLS key from %s", keyFile)
		log.Info("loading server TLS certificate from %s", certFile)
	}

	httpd.certFile = certFile
	httpd.keyFile = keyFile

	return nil
}

func (httpd *HttpsServer) Start() error {
	if err := httpd.Configure(); err != nil {
		return err
	}

	return httpd.SetRunning(true, func() {
		log.Info("HTTPS server starting on https://%s", httpd.server.Addr)
		if err := httpd.server.ListenAndServeTLS(httpd.certFile, httpd.keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})
}

func (httpd *HttpsServer) Stop() error {
	return httpd.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		httpd.server.Shutdown(ctx)
	})
}
