package modules

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"
	"github.com/evilsocket/bettercap-ng/tls"
)

type RestAPI struct {
	session.SessionModule
	server   *http.Server
	username string
	password string
	certFile string
	keyFile  string
}

func NewRestAPI(s *session.Session) *RestAPI {
	api := &RestAPI{
		SessionModule: session.NewSessionModule("api.rest", s),
		server:        &http.Server{},
	}

	api.AddParam(session.NewStringParameter("api.rest.address",
		session.ParamIfaceAddress,
		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`,
		"Address to bind the API REST server to."))

	api.AddParam(session.NewIntParameter("api.rest.port",
		"8083",
		"Port to bind the API REST server to."))

	api.AddParam(session.NewStringParameter("api.rest.username",
		"",
		"",
		"API authentication username."))

	api.AddParam(session.NewStringParameter("api.rest.certificate",
		"~/.bettercap-ng.api.rest.certificate.pem",
		"",
		"API TLS certificate."))

	api.AddParam(session.NewStringParameter("api.rest.key",
		"~/.bettercap-ng.api.rest.key.pem",
		"",
		"API TLS key"))

	api.AddParam(session.NewStringParameter("api.rest.password",
		"",
		"",
		"API authentication password."))

	api.AddHandler(session.NewModuleHandler("api.rest on", "",
		"Start REST API server.",
		func(args []string) error {
			return api.Start()
		}))

	api.AddHandler(session.NewModuleHandler("api.rest off", "",
		"Stop REST API server.",
		func(args []string) error {
			return api.Stop()
		}))

	api.setupRoutes()

	return api
}

type JSSessionRequest struct {
	Command string `json:"cmd"`
}

type JSSessionResponse struct {
	Error string `json:"error"`
}

func (api *RestAPI) Name() string {
	return "api.rest"
}

func (api *RestAPI) Description() string {
	return "Expose a RESTful API."
}

func (api *RestAPI) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (api *RestAPI) Configure() error {
	var err error
	var address string
	var port int

	if err, address = api.StringParam("api.rest.addr"); err != nil {
		return err
	} else if err, port = api.IntParam("api.rest.port"); err != nil {
		return err
	} else {
		api.server.Addr = fmt.Sprintf("%s:%d", address, port)
	}

	if err, api.certFile = api.StringParam("api.rest.certificate"); err != nil {
		return err
	} else if api.certFile, err = core.ExpandPath(api.certFile); err != nil {
		return err
	}

	if err, api.keyFile = api.StringParam("api.rest.key"); err != nil {
		return err
	} else if api.keyFile, err = core.ExpandPath(api.keyFile); err != nil {
		return err
	}

	if err, api.username = api.StringParam("api.rest.username"); err != nil {
		return err
	} else if api.username == "" {
		return fmt.Errorf("api.rest.username is empty.")
	}

	if err, api.password = api.StringParam("api.rest.password"); err != nil {
		return err
	} else if api.password == "" {
		return fmt.Errorf("api.rest.password is empty.")
	}

	if core.Exists(api.certFile) == false || core.Exists(api.keyFile) == false {
		log.Info("Generating RSA key to %s", api.keyFile)
		log.Info("Generating TLS certificate to %s", api.certFile)
		if err := tls.Generate(api.certFile, api.keyFile); err != nil {
			return err
		}
	} else {
		log.Info("Loading RSA key from %s", api.keyFile)
		log.Info("Loading TLS certificate from %s", api.certFile)
	}

	return nil
}

func (api *RestAPI) Start() error {
	if api.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := api.Configure(); err != nil {
		return err
	}

	api.SetRunning(true)
	go func() {
		log.Info("API server starting on https://%s", api.server.Addr)
		err := api.server.ListenAndServeTLS(api.certFile, api.keyFile)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

func (api *RestAPI) Stop() error {
	if api.Running() == false {
		return session.ErrAlreadyStopped
	}
	api.SetRunning(false)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return api.server.Shutdown(ctx)
}
