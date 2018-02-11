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

	"github.com/gin-gonic/gin"
)

type RestAPI struct {
	session.SessionModule
	router   *gin.Engine
	server   *http.Server
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
		session.IPv4Validator,
		"Address to bind the API REST server to."))

	api.AddParam(session.NewIntParameter("api.rest.port",
		"8083",
		"Port to bind the API REST server to."))

	api.AddParam(session.NewStringParameter("api.rest.username",
		"",
		".+",
		"API authentication username."))

	api.AddParam(session.NewStringParameter("api.rest.password",
		"",
		".+",
		"API authentication password."))

	api.AddParam(session.NewStringParameter("api.rest.certificate",
		"~/.bcap-api.rest.certificate.pem",
		"",
		"API TLS certificate."))

	api.AddParam(session.NewStringParameter("api.rest.key",
		"~/.bcap-api.rest.key.pem",
		"",
		"API TLS key"))

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
	var username string
	var password string
	var ip string
	var port int

	if err, ip = api.StringParam("api.rest.address"); err != nil {
		return err
	} else if err, port = api.IntParam("api.rest.port"); err != nil {
		return err
	}
	api.server.Addr = fmt.Sprintf("%s:%d", ip, port)

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

	if err, username = api.StringParam("api.rest.username"); err != nil {
		return err
	}

	if err, password = api.StringParam("api.rest.password"); err != nil {
		return err
	}

	if core.Exists(api.certFile) == false || core.Exists(api.keyFile) == false {
		log.Info("Generating TLS key to %s", api.keyFile)
		log.Info("Generating TLS certificate to %s", api.certFile)
		if err := tls.Generate(api.certFile, api.keyFile); err != nil {
			return err
		}
	} else {
		log.Info("Loading TLS key from %s", api.keyFile)
		log.Info("Loading TLS certificate from %s", api.certFile)
	}

	gin.SetMode(gin.ReleaseMode)

	api.router = gin.New()
	api.router.Use(SecurityMiddleware())
	api.router.Use(gin.BasicAuth(gin.Accounts{username: password}))

	group := api.router.Group("/api")
	group.GET("/session", ShowRestSession)
	group.POST("/session", RunRestCommand)
	group.GET("/events", ShowRestEvents)
	group.DELETE("/events", ClearRestEvents)

	api.server.Handler = api.router

	return nil
}

func (api *RestAPI) Start() error {
	if api.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := api.Configure(); err != nil {
		return err
	}

	api.SetRunning(true, func() {
		log.Info("API server starting on https://%s", api.server.Addr)
		err := api.server.ListenAndServeTLS(api.certFile, api.keyFile)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})

	return nil
}

func (api *RestAPI) Stop() error {
	return api.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		api.server.Shutdown(ctx)
	})
}
