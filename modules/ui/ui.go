package ui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/bettercap/bettercap/v2/session"
)

var (
	//go:embed ui
	web embed.FS
)

type UIModule struct {
	session.SessionModule

	server *http.Server
}

func NewUIModule(s *session.Session) *UIModule {
	mod := &UIModule{
		SessionModule: session.NewSessionModule("ui", s),
		server:        &http.Server{},
	}

	mod.SessionModule.Requires("api.rest")

	mod.AddHandler(session.NewModuleHandler("ui on", "",
		"Start the web user interface.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("ui off", "",
		"Stop the web user interface.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddParam(session.NewStringParameter("ui.address",
		"127.0.0.1",
		session.IPv4Validator,
		"Address to bind the web ui to."))

	mod.AddParam(session.NewIntParameter("ui.port",
		"8080",
		"Port to bind the web ui server to."))

	return mod
}

func (mod *UIModule) Name() string {
	return "ui"
}

func (mod *UIModule) Description() string {
	return "Web User Interface."
}

func (mod *UIModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *UIModule) Configure() (err error) {
	var ip string
	var port int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, ip = mod.StringParam("ui.address"); err != nil {
		return err
	} else if err, port = mod.IntParam("ui.port"); err != nil {
		return err
	}

	dist, _ := fs.Sub(web, "ui")
	mod.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", ip, port),
		Handler: http.FileServer(http.FS(dist)),
	}

	return nil
}

func (mod *UIModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		defer mod.SetRunning(false, nil)

		var err error

		mod.Info("web ui starting on http://%s", mod.server.Addr)
		err = mod.server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			mod.Error("web ui failed: %v", err)
		}
	})
}

func (mod *UIModule) Stop() error {
	return mod.SetRunning(false, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		mod.server.Shutdown(ctx)
	})
}
