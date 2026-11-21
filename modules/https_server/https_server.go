package https_server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/session"
	btls "github.com/bettercap/bettercap/v2/tls"

	"github.com/evilsocket/islazy/tui"
)

type HttpsServer struct {
	session.SessionModule
	server *http.Server
	cert   *tls.Certificate
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
		"Address to bind the HTTPS server to."))

	mod.AddParam(session.NewIntParameter("https.server.port",
		"443",
		"Port to bind the HTTPS server to."))

	mod.AddParam(session.NewStringParameter("https.server.certificate",
		"~/.bettercap-httpd.cert.pem",
		"",
		"Server TLS certificate for 'auto' mode (generated if missing); ignored in 'emudhra' mode."))

	mod.AddParam(session.NewStringParameter("https.server.key",
		"~/.bettercap-httpd.key.pem",
		"",
		"Server TLS key for 'auto' mode (generated if missing); ignored in 'emudhra' mode."))

	btls.CertConfigToModule("https.server", &mod.SessionModule, btls.DefaultLegitConfig)
	btls.CAConfigToModule("https.server", &mod.SessionModule)

	mod.AddHandler(session.NewModuleHandler("https.server on", "",
		"Start HTTPS server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("https.server off", "",
		"Stop HTTPS server.",
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

	caCfg, err := btls.CAConfigFromModule("https.server", mod.SessionModule)
	if err != nil {
		return err
	}

	if caCfg.IsEmudhra() {
		ca, err := caCfg.Resolve()
		if err != nil {
			return err
		}
		mod.Warning("serving with revoked eMudhra CA %s", ca)
	} else if exists, err := caCfg.FilesExist(); err != nil {
		return err
	} else if exists {
		mod.Info("loading server TLS key from %s", caCfg.KeyFile)
		mod.Info("loading server TLS certificate from %s", caCfg.CertFile)
	} else {
		mod.Debug("%+v", caCfg.Config)
		mod.Info("generating server TLS key to %s", caCfg.KeyFile)
		mod.Info("generating server TLS certificate to %s", caCfg.CertFile)
	}

	cert, err := caCfg.LoadForHost(address, caCfg.Config.Bits)
	if err != nil {
		return err
	}

	mod.cert = cert

	return nil
}

func (mod *HttpsServer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("starting on https://%s", mod.server.Addr)
		mod.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*mod.cert},
		}
		if err := mod.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
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
