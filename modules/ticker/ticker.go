package ticker

import (
	"errors"
	"strconv"
	"time"

	"github.com/bettercap/bettercap/v2/session"
)

type Params struct {
	Period   time.Duration
	Commands []string
	Running  bool
}

type Ticker struct {
	session.SessionModule

	main  Params
	named map[string]*Params
}

func NewTicker(s *session.Session) *Ticker {
	mod := &Ticker{
		SessionModule: session.NewSessionModule("ticker", s),
		named:         make(map[string]*Params),
	}

	mod.AddParam(session.NewStringParameter("ticker.commands",
		"clear; net.show; events.show 20",
		"",
		"List of commands for the main ticker separated by a ;"))

	mod.AddParam(session.NewIntParameter("ticker.period",
		"1",
		"Main ticker period in seconds"))

	mod.AddHandler(session.NewModuleHandler("ticker on", "",
		"Start the main ticker.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("ticker off", "",
		"Stop the main ticker.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("ticker.create <name> <period> <commands>",
		`(?i)^ticker\.create\s+([^\s]+)\s+(\d+)\s+(.+)$`,
		"Create and start a named ticker.",
		func(args []string) error {
			if period, err := strconv.Atoi(args[1]); err != nil {
				return err
			} else {
				return mod.createNamed(args[0], period, args[2])
			}
		}))

	mod.AddHandler(session.NewModuleHandler("ticker.destroy <name>",
		`(?i)^ticker\.destroy\s+([^\s]+)$`,
		"Stop a named ticker.",
		func(args []string) error {
			return mod.destroyNamed(args[0])
		}))

	return mod
}

func (mod *Ticker) Name() string {
	return "ticker"
}

func (mod *Ticker) Description() string {
	return "A module to execute one or more commands every given amount of seconds."
}

func (mod *Ticker) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *Ticker) Configure() error {
	var err error
	var commands string
	var period int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, commands = mod.StringParam("ticker.commands"); err != nil {
		return err
	} else if err, period = mod.IntParam("ticker.period"); err != nil {
		return err
	}

	mod.main = Params{
		Commands: session.ParseCommands(commands),
		Period:   time.Duration(period) * time.Second,
		Running:  true,
	}

	return nil
}

type TickEvent struct{}

func (mod *Ticker) worker(name string, params *Params) {
	isMain := name == "main"
	eventName := "tick"

	if isMain {
		mod.Info("main ticker running with period %.fs", params.Period.Seconds())
	} else {
		eventName = "ticker." + name
		mod.Info("ticker '%s' running with period %.fs", name, params.Period.Seconds())
	}

	tick := time.NewTicker(params.Period)
	for range tick.C {
		if !params.Running {
			break
		}

		session.I.Events.Add(eventName, TickEvent{})
		for _, cmd := range params.Commands {
			if err := mod.Session.Run(cmd); err != nil {
				mod.Error("%s", err)
			}
		}
	}

	if isMain {
		mod.Info("main ticker stopped")
	} else {
		mod.Info("ticker '%s' stopped", name)
	}
}

func (mod *Ticker) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.worker("main", &mod.main)
	})
}

func (mod *Ticker) Stop() error {
	mod.main.Running = false
	for _, params := range mod.named {
		params.Running = false
	}

	return mod.SetRunning(false, nil)
}

func (mod *Ticker) createNamed(name string, period int, commands string) error {
	if _, found := mod.named[name]; found {
		return errors.New("ticker '" + name + "' already exists")
	}

	params := &Params{
		Commands: session.ParseCommands(commands),
		Period:   time.Duration(period) * time.Second,
		Running:  true,
	}

	mod.named[name] = params

	go mod.worker(name, params)

	return nil
}

func (mod *Ticker) destroyNamed(name string) error {
	if _, found := mod.named[name]; !found {
		return errors.New("ticker '" + name + "' not found")
	}

	mod.named[name].Running = false
	delete(mod.named, name)

	return nil
}
