package ticker

import (
	"time"

	"github.com/bettercap/bettercap/session"
)

type Ticker struct {
	session.SessionModule
	Period   time.Duration
	Commands []string
}

func NewTicker(s *session.Session) *Ticker {
	mod := &Ticker{
		SessionModule: session.NewSessionModule("ticker", s),
	}

	mod.AddParam(session.NewStringParameter("ticker.commands",
		"clear; net.show; events.show 20",
		"",
		"List of commands separated by a ;"))

	mod.AddParam(session.NewIntParameter("ticker.period",
		"1",
		"Ticker period in seconds"))

	mod.AddHandler(session.NewModuleHandler("ticker on", "",
		"Start the ticker.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("ticker off", "",
		"Stop the ticker.",
		func(args []string) error {
			return mod.Stop()
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

	mod.Commands = session.ParseCommands(commands)
	mod.Period = time.Duration(period) * time.Second

	return nil
}

type TickEvent struct {}

func (mod *Ticker) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("running with period %.fs", mod.Period.Seconds())
		tick := time.NewTicker(mod.Period)
		for range tick.C {
			if !mod.Running() {
				break
			}

			session.I.Events.Add("tick", TickEvent{})

			for _, cmd := range mod.Commands {
				if err := mod.Session.Run(cmd); err != nil {
					mod.Error("%s", err)
				}
			}
		}
	})
}

func (mod *Ticker) Stop() error {
	return mod.SetRunning(false, nil)
}
