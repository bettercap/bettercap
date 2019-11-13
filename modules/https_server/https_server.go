package https_server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	mod := &HttpsServer{
		SessionModule: session.NewSessionModule("https.server", s),
		server:        &http.Server{},
	}

	mod.AddParam(session.NewStringParameter("https.server.path",
		".",
		"",
		"Server folder."))

	mod.AddParam(session.NewStringParameter("https.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the http server to."))

	mod.AddParam(session.NewIntParameter("https.server.port",
		"443",
		"Port to bind the http server to."))

	mod.AddParam(session.NewStringParameter("https.server.certificate",
		"~/.bettercap-httpd.cert.pem",
		"",
		"TLS certificate file (will be auto generated if filled but not existing)."))

	mod.AddParam(session.NewStringParameter("https.server.key",
		"~/.bettercap-httpd.key.pem",
		"",
		"TLS key file (will be auto generated if filled but not existing)."))

	tls.CertConfigToModule("https.server", &mod.SessionModule, tls.DefaultLegitConfig)

	mod.AddHandler(session.NewModuleHandler("https.server on", "",
		"Start https server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("https.server off", "",
		"Stop https server.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *HttpsServer) Name() string {
	return "https.server"
}

func (mod *HttpsServer) Description() string {
	return "A simple HTTPS server, to be used to serve files and scripts across the network."
}

func (mod *HttpsServer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *HttpsServer) Configure() error {
	var err error
	var path string
	var address string
	var port int
	var certFile string
	var keyFile string

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if err, path = mod.StringParam("https.server.path"); err != nil {
		return err
	}

	router := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(path))

	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mod.Debug("%s %s %s%s", tui.Bold(strings.Split(r.RemoteAddr, ":")[0]), r.Method, r.Host, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	}))

	mod.server.Handler = router

	if err, address = mod.StringParam("https.server.address"); err != nil {
		return err
	}

	if err, port = mod.IntParam("https.server.port"); err != nil {
		return err
	}

	mod.server.Addr = fmt.Sprintf("%s:%d", address, port)

	if err, certFile = mod.StringParam("https.server.certificate"); err != nil {
		return err
	} else if certFile, err = fs.Expand(certFile); err != nil {
		return err
	}

	if err, keyFile = mod.StringParam("https.server.key"); err != nil {
		return err
	} else if keyFile, err = fs.Expand(keyFile); err != nil {
		return err
	}

	if !fs.Exists(certFile) || !fs.Exists(keyFile) {
		cfg, err := tls.CertConfigFromModule("https.server", mod.SessionModule)
		if err != nil {
			return err
		}

		mod.Debug("%+v", cfg)
		mod.Info("generating server TLS key to %s", keyFile)
		mod.Info("generating server TLS certificate to %s", certFile)
		if err := tls.Generate(cfg, certFile, keyFile, false); err != nil {
			return err
		}
	} else {
		mod.Info("loading server TLS key from %s", keyFile)
		mod.Info("loading server TLS certificate from %s", certFile)
	}

	mod.certFile = certFile
	mod.keyFile = keyFile

	return nil
}

func (mod *HttpsServer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("starting on https://%s", mod.server.Addr)
		if err := mod.server.ListenAndServeTLS(mod.certFile, mod.keyFile); err != nil && err != http.ErrServerClosed {
			mod.Error("%v", err)
			mod.Stop()
		}
	})
}

func (mod *HttpsServer) Stop() error {
	return mod.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		mod.server.Shutdown(ctx)
	})
}
