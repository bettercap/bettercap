package api_rest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bettercap/bettercap/session"
	"github.com/bettercap/bettercap/tls"

	"github.com/bettercap/recording"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/evilsocket/islazy/fs"
)

type RestAPI struct {
	session.SessionModule
	server       *http.Server
	username     string
	password     string
	certFile     string
	keyFile      string
	allowOrigin  string
	useWebsocket bool
	upgrader     websocket.Upgrader
	quit         chan bool

	recClock       int
	recording      bool
	recTime        int
	loading        bool
	replaying      bool
	recordFileName string
	recordWait     *sync.WaitGroup
	record         *recording.Archive
	recStarted     time.Time
	recStopped     time.Time
}

func NewRestAPI(s *session.Session) *RestAPI {
	mod := &RestAPI{
		SessionModule: session.NewSessionModule("api.rest", s),
		server:        &http.Server{},
		quit:          make(chan bool),
		useWebsocket:  false,
		allowOrigin:   "*",
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		recClock:       1,
		recording:      false,
		recTime:        0,
		loading:        false,
		replaying:      false,
		recordFileName: "",
		recordWait:     &sync.WaitGroup{},
		record:         nil,
	}

	mod.State.Store("recording", &mod.recording)
	mod.State.Store("rec_clock", &mod.recClock)
	mod.State.Store("replaying", &mod.replaying)
	mod.State.Store("loading", &mod.loading)
	mod.State.Store("load_progress", 0)
	mod.State.Store("rec_time", &mod.recTime)
	mod.State.Store("rec_filename", &mod.recordFileName)
	mod.State.Store("rec_frames", 0)
	mod.State.Store("rec_cur_frame", 0)
	mod.State.Store("rec_started", &mod.recStarted)
	mod.State.Store("rec_stopped", &mod.recStopped)

	mod.AddParam(session.NewStringParameter("api.rest.address",
		"127.0.0.1",
		session.IPv4Validator,
		"Address to bind the API REST server to."))

	mod.AddParam(session.NewIntParameter("api.rest.port",
		"8081",
		"Port to bind the API REST server to."))

	mod.AddParam(session.NewStringParameter("api.rest.alloworigin",
		mod.allowOrigin,
		"",
		"Value of the Access-Control-Allow-Origin header of the API server."))

	mod.AddParam(session.NewStringParameter("api.rest.username",
		"",
		"",
		"API authentication username."))

	mod.AddParam(session.NewStringParameter("api.rest.password",
		"",
		"",
		"API authentication password."))

	mod.AddParam(session.NewStringParameter("api.rest.certificate",
		"",
		"",
		"API TLS certificate."))

	tls.CertConfigToModule("api.rest", &mod.SessionModule, tls.DefaultLegitConfig)

	mod.AddParam(session.NewStringParameter("api.rest.key",
		"",
		"",
		"API TLS key"))

	mod.AddParam(session.NewBoolParameter("api.rest.websocket",
		"false",
		"If true the /api/events route will be available as a websocket endpoint instead of HTTPS."))

	mod.AddHandler(session.NewModuleHandler("api.rest on", "",
		"Start REST API server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("api.rest off", "",
		"Stop REST API server.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddParam(session.NewIntParameter("api.rest.record.clock",
		"1",
		"Number of seconds to wait while recording with api.rest.record between one sample and the next one."))

	mod.AddHandler(session.NewModuleHandler("api.rest.record off", "",
		"Stop recording the session.",
		func(args []string) error {
			return mod.stopRecording()
		}))

	mod.AddHandler(session.NewModuleHandler("api.rest.record FILENAME", `api\.rest\.record (.+)`,
		"Start polling the rest API periodically recording each sample in a compressed file that can be later replayed.",
		func(args []string) error {
			return mod.startRecording(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("api.rest.replay off", "",
		"Stop replaying the recorded session.",
		func(args []string) error {
			return mod.stopReplay()
		}))

	mod.AddHandler(session.NewModuleHandler("api.rest.replay FILENAME", `api\.rest\.replay (.+)`,
		"Start the rest API module in replay mode using FILENAME as the recorded session file, will revert to normal mode once the replay is over.",
		func(args []string) error {
			return mod.startReplay(args[0])
		}))

	return mod
}

type JSSessionRequest struct {
	Command string `json:"cmd"`
}

type JSSessionResponse struct {
	Error string `json:"error"`
}

func (mod *RestAPI) Name() string {
	return "api.rest"
}

func (mod *RestAPI) Description() string {
	return "Expose a RESTful API."
}

func (mod *RestAPI) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *RestAPI) isTLS() bool {
	return mod.certFile != "" && mod.keyFile != ""
}

func (mod *RestAPI) Configure() error {
	var err error
	var ip string
	var port int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, ip = mod.StringParam("api.rest.address"); err != nil {
		return err
	} else if err, port = mod.IntParam("api.rest.port"); err != nil {
		return err
	} else if err, mod.allowOrigin = mod.StringParam("api.rest.alloworigin"); err != nil {
		return err
	} else if err, mod.certFile = mod.StringParam("api.rest.certificate"); err != nil {
		return err
	} else if mod.certFile, err = fs.Expand(mod.certFile); err != nil {
		return err
	} else if err, mod.keyFile = mod.StringParam("api.rest.key"); err != nil {
		return err
	} else if mod.keyFile, err = fs.Expand(mod.keyFile); err != nil {
		return err
	} else if err, mod.username = mod.StringParam("api.rest.username"); err != nil {
		return err
	} else if err, mod.password = mod.StringParam("api.rest.password"); err != nil {
		return err
	} else if err, mod.useWebsocket = mod.BoolParam("api.rest.websocket"); err != nil {
		return err
	}

	if mod.isTLS() {
		if !fs.Exists(mod.certFile) || !fs.Exists(mod.keyFile) {
			cfg, err := tls.CertConfigFromModule("api.rest", mod.SessionModule)
			if err != nil {
				return err
			}

			mod.Debug("%+v", cfg)
			mod.Info("generating TLS key to %s", mod.keyFile)
			mod.Info("generating TLS certificate to %s", mod.certFile)
			if err := tls.Generate(cfg, mod.certFile, mod.keyFile, false); err != nil {
				return err
			}
		} else {
			mod.Info("loading TLS key from %s", mod.keyFile)
			mod.Info("loading TLS certificate from %s", mod.certFile)
		}
	}

	mod.server.Addr = fmt.Sprintf("%s:%d", ip, port)

	router := mux.NewRouter()

	router.Methods("OPTIONS").HandlerFunc(mod.corsRoute)

	router.HandleFunc("/api/file", mod.fileRoute)

	router.HandleFunc("/api/events", mod.eventsRoute)

	router.HandleFunc("/api/session", mod.sessionRoute)
	router.HandleFunc("/api/session/ble", mod.sessionRoute)
	router.HandleFunc("/api/session/ble/{mac}", mod.sessionRoute)
	router.HandleFunc("/api/session/hid", mod.sessionRoute)
	router.HandleFunc("/api/session/hid/{mac}", mod.sessionRoute)
	router.HandleFunc("/api/session/env", mod.sessionRoute)
	router.HandleFunc("/api/session/gateway", mod.sessionRoute)
	router.HandleFunc("/api/session/interface", mod.sessionRoute)
	router.HandleFunc("/api/session/modules", mod.sessionRoute)
	router.HandleFunc("/api/session/lan", mod.sessionRoute)
	router.HandleFunc("/api/session/lan/{mac}", mod.sessionRoute)
	router.HandleFunc("/api/session/options", mod.sessionRoute)
	router.HandleFunc("/api/session/packets", mod.sessionRoute)
	router.HandleFunc("/api/session/started-at", mod.sessionRoute)
	router.HandleFunc("/api/session/wifi", mod.sessionRoute)
	router.HandleFunc("/api/session/wifi/{mac}", mod.sessionRoute)

	mod.server.Handler = router

	if mod.username == "" || mod.password == "" {
		mod.Warning("api.rest.username and/or api.rest.password parameters are empty, authentication is disabled.")
	}

	return nil
}

func (mod *RestAPI) Start() error {
	if mod.replaying {
		return fmt.Errorf("the api is currently in replay mode, run api.rest.replay off before starting it")
	} else if err := mod.Configure(); err != nil {
		return err
	}

	mod.SetRunning(true, func() {
		var err error

		if mod.isTLS() {
			mod.Info("api server starting on https://%s", mod.server.Addr)
			err = mod.server.ListenAndServeTLS(mod.certFile, mod.keyFile)
		} else {
			mod.Info("api server starting on http://%s", mod.server.Addr)
			err = mod.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})

	return nil
}

func (mod *RestAPI) Stop() error {
	if mod.recording {
		mod.stopRecording()
	} else if mod.replaying {
		mod.stopReplay()
	}

	return mod.SetRunning(false, func() {
		go func() {
			mod.quit <- true
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		mod.server.Shutdown(ctx)
	})
}
