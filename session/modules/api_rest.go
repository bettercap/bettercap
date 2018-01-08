package session_modules

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
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
		"~/.bettercap-ng.certificate.pem",
		"",
		"API TLS certificate."))

	api.AddParam(session.NewStringParameter("api.rest.key",
		"~/.bettercap-ng.key.pem",
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

	http.HandleFunc("/api/session", api.sessRoute)
	http.HandleFunc("/api/events", api.eventsRoute)

	return api
}

type JSSessionRequest struct {
	Command string `json:"cmd"`
}

type JSSessionResponse struct {
	Error string `json:"error"`
}

func (api *RestAPI) sessRoute(w http.ResponseWriter, r *http.Request) {
	if api.checkAuth(w, r) == false {
		return
	}

	if r.Method == "GET" {
		js, err := json.Marshal(api.Session)
		if err != nil {
			api.Session.Events.Log(session.ERROR, "Error while returning session: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if r.Method == "POST" && r.Body != nil {
		var req JSSessionRequest
		var res JSSessionResponse

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = api.Session.Run(req.Command)
		if err != nil {
			res.Error = err.Error()
		}
		js, err := json.Marshal(res)
		if err != nil {
			api.Session.Events.Log(session.ERROR, "Error while returning response: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	if api.checkAuth(w, r) == false {
		return
	}

	if r.Method == "GET" {
		var err error

		events := api.Session.Events.Events()
		nmax := len(events)
		n := nmax

		keys, ok := r.URL.Query()["n"]
		if len(keys) == 1 && ok {
			sn := keys[0]
			n, err = strconv.Atoi(sn)
			if err == nil {
				if n > nmax {
					n = nmax
				}
			} else {
				n = nmax
			}
		}

		js, err := json.Marshal(events[0:n])
		if err != nil {
			api.Session.Events.Log(session.ERROR, "Error while returning events: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if r.Method == "DELETE" {
		api.Session.Events.Clear()
		api.Session.Events.Add("sys.log.cleared", nil)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api RestAPI) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if api.Authenticated(w, r) == false {
		api.Session.Events.Log(session.WARNING, "Unauthenticated access!")
		http.Error(w, "Not authorized", 401)
		return false
	}
	return true
}

func (api RestAPI) Authenticated(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	if pair[0] != api.username || pair[1] != api.password {
		return false
	}

	return true
}

func (api RestAPI) Name() string {
	return "REST API"
}

func (api RestAPI) Description() string {
	return "Expose a RESTful API."
}

func (api RestAPI) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (api RestAPI) OnSessionStarted(s *session.Session) {
	// refresh the address after session has been created
	s.Env.Set("api.rest.address", s.Interface.IpAddress)
}

func (api RestAPI) OnSessionEnded(s *session.Session) {
	if api.Running() {
		api.Stop()
	}
}

func (api *RestAPI) Start() error {
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
		api.Session.Events.Log(session.INFO, "Generating RSA key to %s", api.keyFile)
		api.Session.Events.Log(session.INFO, "Generating TLS certificate to %s", api.certFile)
		if err := tls.Generate(api.certFile, api.keyFile); err != nil {
			return err
		}
	} else {
		api.Session.Events.Log(session.INFO, "Loading RSA key from %s", api.keyFile)
		api.Session.Events.Log(session.INFO, "Loading TLS certificate from %s", api.certFile)
	}

	if api.Running() == false {
		api.SetRunning(true)
		api.server.Addr = fmt.Sprintf("%s:%d", address, port)
		go func() {
			api.Session.Events.Log(session.INFO, "API server starting on https://%s", api.server.Addr)
			err := api.server.ListenAndServeTLS(api.certFile, api.keyFile)
			if err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		return nil
	} else {
		return fmt.Errorf("REST API server already started.")
	}
}

func (api *RestAPI) Stop() error {
	if api.Running() == true {
		api.SetRunning(false)
		return api.server.Shutdown(nil)
	} else {
		return fmt.Errorf("REST API server already stopped.")
	}
}
