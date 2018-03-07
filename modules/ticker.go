package modules

import (
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

type Ticker struct {
	session.SessionModule
	Period   time.Duration
	Commands []string
}

func NewTicker(s *session.Session) *Ticker {
	t := &Ticker{
		SessionModule: session.NewSessionModule("ticker", s),
	}

	t.AddParam(session.NewStringParameter("ticker.commands",
		"clear; net.show; events.show 20",
		"",
		"List of commands separated by a ;"))

	t.AddParam(session.NewIntParameter("ticker.period",
		"1",
		"Ticker period in seconds"))

	t.AddHandler(session.NewModuleHandler("ticker on", "",
		"Start the ticker.",
		func(args []string) error {
			return t.Start()
		}))

	t.AddHandler(session.NewModuleHandler("ticker off", "",
		"Stop the ticker.",
		func(args []string) error {
			return t.Stop()
		}))

	return t
}

func (t *Ticker) Name() string {
	return "ticker"
}

func (t *Ticker) Description() string {
	return "A module to execute one or more commands every given amount of seconds."
}

func (t *Ticker) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (t *Ticker) Configure() error {
	var err error
	var commands string
	var period int

	if t.Running() == true {
		return session.ErrAlreadyStarted
	} else if err, commands = t.StringParam("ticker.commands"); err != nil {
		return err
	} else if err, period = t.IntParam("ticker.period"); err != nil {
		return err
	}

	t.Commands = session.ParseCommands(commands)
	t.Period = time.Duration(period) * time.Second

	return nil
}

func (t *Ticker) Start() error {
	if err := t.Configure(); err != nil {
		return err
	}

	return t.SetRunning(true, func() {
		log.Info("Ticker running with period %.fs.", t.Period.Seconds())
		tick := time.Tick(t.Period)
		for range tick {
			if t.Running() == false {
				break
			}

			for _, cmd := range t.Commands {
				if err := t.Session.Run(cmd); err != nil {
					log.Error("%s", err)
				}
			}
		}
	})
}

func (t *Ticker) Stop() error {
	return t.SetRunning(false, nil)
}
