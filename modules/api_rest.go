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
		username:      "",
		password:      "",
	}

	api.AddParam(session.NewStringParameter("api.rest.address",
		"<interface address>",
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
	return "REST API"
}

func (api *RestAPI) Description() string {
	return "Expose a RESTful API."
}

func (api *RestAPI) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (api *RestAPI) OnSessionStarted(s *session.Session) {
	// refresh the address after session has been created
	s.Env.Set("api.rest.address", s.Interface.IpAddress)
}

func (api *RestAPI) OnSessionEnded(s *session.Session) {
	if api.Running() {
		api.Stop()
	}
}

func (api *RestAPI) configure() error {
	var address string
	var port int

	if err, v := api.Param("api.rest.address").Get(api.Session); err != nil {
		return err
	} else {
		address = v.(string)
	}

	if err, v := api.Param("api.rest.port").Get(api.Session); err != nil {
		return err
	} else {
		port = v.(int)
	}

	api.server.Addr = fmt.Sprintf("%s:%d", address, port)

	if err, v := api.Param("api.rest.certificate").Get(api.Session); err != nil {
		return err
	} else {
		api.certFile = v.(string)
		if api.certFile, err = core.ExpandPath(api.certFile); err != nil {
			return err
		}
	}

	if err, v := api.Param("api.rest.key").Get(api.Session); err != nil {
		return err
	} else {
		api.keyFile = v.(string)
		if api.keyFile, err = core.ExpandPath(api.keyFile); err != nil {
			return err
		}
	}

	if err, v := api.Param("api.rest.username").Get(api.Session); err != nil {
		return err
	} else {
		api.username = v.(string)
		if api.username == "" {
			return fmt.Errorf("api.rest.username is empty.")
		}
	}

	if err, v := api.Param("api.rest.password").Get(api.Session); err != nil {
		return err
	} else {
		api.password = v.(string)
		if api.password == "" {
			return fmt.Errorf("api.rest.password is empty.")
		}
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
	if err := api.configure(); err != nil {
		return err
	}

	if api.Running() == false {
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

	return fmt.Errorf("REST API server already started.")
}

func (api *RestAPI) Stop() error {
	if api.Running() == true {
		api.SetRunning(false)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		return api.server.Shutdown(ctx)
	} else {
		return fmt.Errorf("REST API server already stopped.")
	}
}
