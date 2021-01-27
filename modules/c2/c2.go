package c2

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/acarl005/stripansi"
	"github.com/bettercap/bettercap/modules/events_stream"
	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
	irc "github.com/thoj/go-ircevent"
	"strings"
	"text/template"
)

type settings struct {
	server         string
	tls            bool
	tlsVerify      bool
	nick           string
	user           string
	password       string
	saslUser       string
	saslPassword   string
	operator       string
	controlChannel string
	eventsChannel  string
	outputChannel  string
}

type C2 struct {
	session.SessionModule

	settings  settings
	stream    *events_stream.EventsStream
	templates map[string]*template.Template
	channels  map[string]string
	client    *irc.Connection
	eventBus  session.EventBus
	quit      chan bool
}

type eventContext struct {
	Session *session.Session
	Event   session.Event
}

func NewC2(s *session.Session) *C2 {
	mod := &C2{
		SessionModule: session.NewSessionModule("c2", s),
		stream:        events_stream.NewEventsStream(s),
		templates:     make(map[string]*template.Template),
		channels:      make(map[string]string),
		quit:          make(chan bool),
		settings: settings{
			server:         "localhost:6697",
			tls:            true,
			tlsVerify:      false,
			nick:           "bettercap",
			user:           "bettercap",
			password:       "password",
			operator:       "admin",
			eventsChannel:  "#events",
			outputChannel:  "#events",
			controlChannel: "#events",
		},
	}

	mod.AddParam(session.NewStringParameter("c2.server",
		mod.settings.server,
		"",
		"IRC server address and port."))

	mod.AddParam(session.NewBoolParameter("c2.server.tls",
		"true",
		"Enable TLS."))

	mod.AddParam(session.NewBoolParameter("c2.server.tls.verify",
		"false",
		"Enable TLS certificate validation."))

	mod.AddParam(session.NewStringParameter("c2.operator",
		mod.settings.operator,
		"",
		"IRC nickname of the user allowed to run commands."))

	mod.AddParam(session.NewStringParameter("c2.nick",
		mod.settings.nick,
		"",
		"IRC nickname."))

	mod.AddParam(session.NewStringParameter("c2.username",
		mod.settings.user,
		"",
		"IRC username."))

	mod.AddParam(session.NewStringParameter("c2.password",
		mod.settings.password,
		"",
		"IRC server password."))

	mod.AddParam(session.NewStringParameter("c2.sasl.username",
		mod.settings.saslUser,
		"",
		"IRC SASL username."))

	mod.AddParam(session.NewStringParameter("c2.sasl.password",
		mod.settings.saslPassword,
		"",
		"IRC server SASL password."))

	mod.AddParam(session.NewStringParameter("c2.channel.output",
		mod.settings.outputChannel,
		"",
		"IRC channel to send commands output to."))

	mod.AddParam(session.NewStringParameter("c2.channel.events",
		mod.settings.eventsChannel,
		"",
		"IRC channel to send events to."))

	mod.AddParam(session.NewStringParameter("c2.channel.control",
		mod.settings.controlChannel,
		"",
		"IRC channel to receive commands from."))

	mod.AddHandler(session.NewModuleHandler("c2 on", "",
		"Start the C2 module.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("c2 off", "",
		"Stop the C2 module.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("c2.channel.set EVENT_TYPE CHANNEL",
		"c2.channel.set ([^\\s]+) (.+)",
		"Set a specific channel to report events of this type.",
		func(args []string) error {
			eventType := args[0]
			channel := args[1]

			mod.Debug("setting channel for event %s: %v", eventType, channel)
			mod.channels[eventType] = channel
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("c2.channel.clear EVENT_TYPE",
		"c2.channel.clear ([^\\s]+)",
		"Clear the channel to use for a specific event type.",
		func(args []string) error {
			eventType := args[0]
			if _, found := mod.channels[args[0]]; found {
				delete(mod.channels, eventType)
				mod.Debug("cleared channel for %s", eventType)
			} else {
				return fmt.Errorf("channel for event %s not set", args[0])
			}
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("c2.template.set EVENT_TYPE TEMPLATE",
		"c2.template.set ([^\\s]+) (.+)",
		"Set the reporting template to use for a specific event type.",
		func(args []string) error {
			eventType := args[0]
			eventTemplate := args[1]

			parsed, err := template.New(eventType).Parse(eventTemplate)
			if err != nil {
				return err
			}

			mod.Debug("setting template for event %s: %v", eventType, parsed)
			mod.templates[eventType] = parsed
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("c2.template.clear EVENT_TYPE",
		"c2.template.clear ([^\\s]+)",
		"Clear the reporting template to use for a specific event type.",
		func(args []string) error {
			eventType := args[0]
			if _, found := mod.templates[args[0]]; found {
				delete(mod.templates, eventType)
				mod.Debug("cleared template for %s", eventType)
			} else {
				return fmt.Errorf("template for event %s not set", args[0])
			}
			return nil
		}))

	mod.Session.Events.OnPrint(mod.onPrint)

	return mod
}

func (mod *C2) Name() string {
	return "c2"
}

func (mod *C2) Description() string {
	return "A CnC module that connects to an IRC server for reporting and commands."
}

func (mod *C2) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *C2) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if err, mod.settings.server = mod.StringParam("c2.server"); err != nil {
		return err
	} else if err, mod.settings.tls = mod.BoolParam("c2.server.tls"); err != nil {
		return err
	} else if err, mod.settings.tlsVerify = mod.BoolParam("c2.server.tls.verify"); err != nil {
		return err
	} else if err, mod.settings.nick = mod.StringParam("c2.nick"); err != nil {
		return err
	} else if err, mod.settings.user = mod.StringParam("c2.username"); err != nil {
		return err
	} else if err, mod.settings.password = mod.StringParam("c2.password"); err != nil {
		return err
	} else if err, mod.settings.saslUser = mod.StringParam("c2.sasl.username"); err != nil {
		return err
	} else if err, mod.settings.saslPassword = mod.StringParam("c2.sasl.password"); err != nil {
		return err
	} else if err, mod.settings.operator = mod.StringParam("c2.operator"); err != nil {
		return err
	} else if err, mod.settings.eventsChannel = mod.StringParam("c2.channel.events"); err != nil {
		return err
	} else if err, mod.settings.controlChannel = mod.StringParam("c2.channel.control"); err != nil {
		return err
	} else if err, mod.settings.outputChannel = mod.StringParam("c2.channel.output"); err != nil {
		return err
	}

	mod.eventBus = mod.Session.Events.Listen()

	mod.client = irc.IRC(mod.settings.nick, mod.settings.user)

	if log.Level == log.DEBUG {
		mod.client.VerboseCallbackHandler = true
		mod.client.Debug = true
	}

	mod.client.Password = mod.settings.password
	mod.client.UseTLS = mod.settings.tls
	mod.client.TLSConfig = &tls.Config{
		InsecureSkipVerify: !mod.settings.tlsVerify,
	}

	if mod.settings.saslUser != "" || mod.settings.saslPassword != "" {
		mod.client.SASLLogin = mod.settings.saslUser
		mod.client.SASLPassword = mod.settings.saslPassword
		mod.client.UseSASL = true
	}

	mod.client.AddCallback("PRIVMSG", func(event *irc.Event) {
		channel := event.Arguments[0]
		message := event.Message()
		from := event.Nick

		if from != mod.settings.operator {
			mod.client.Privmsg(event.Nick, "nope")
			return
		}

		if channel != mod.settings.controlChannel && channel != mod.settings.nick {
			mod.Debug("from:%s on:%s - '%s'", from, channel, message)
			return
		}

		mod.Debug("from:%s on:%s - '%s'", from, channel, message)

		parts := strings.SplitN(message, " ", 2)
		cmd := parts[0]
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		if cmd == "join" {
			mod.client.Join(args)
		} else if cmd == "part" {
			mod.client.Part(args)
		} else if cmd == "nick" {
			mod.client.Nick(args)
		} else if err = mod.Session.Run(message); err == nil {

		} else {
			mod.client.Privmsgf(event.Nick, "error: %v", stripansi.Strip(err.Error()))
		}
	})

	mod.client.AddCallback("001", func(e *irc.Event) {
		mod.Debug("got 101")
		mod.client.Join(mod.settings.controlChannel)
		mod.client.Join(mod.settings.outputChannel)
		mod.client.Join(mod.settings.eventsChannel)
	})

	return mod.client.Connect(mod.settings.server)
}

func (mod *C2) onPrint(format string, args ...interface{}) {
	if !mod.Running() {
		return
	}

	msg := stripansi.Strip(str.Trim(fmt.Sprintf(format, args...)))

	for _, line := range strings.Split(msg, "\n") {
		mod.client.Privmsg(mod.settings.outputChannel, line)
	}
}

func (mod *C2) onEvent(e session.Event) {
	if mod.Session.EventsIgnoreList.Ignored(e) {
		return
	}

	// default channel or event specific channel?
	channel := mod.settings.eventsChannel
	if custom, found := mod.channels[e.Tag]; found {
		channel = custom
	}

	var out bytes.Buffer
	if tpl, found := mod.templates[e.Tag]; found {
		// use a custom template to render this event
		if err := tpl.Execute(&out, eventContext{
			Session: mod.Session,
			Event:   e,
		}); err != nil {
			fmt.Fprintf(&out, "%v", err)
		}
	} else {
		// use the default view to render this event
		mod.stream.Render(&out, e)
	}

	// make sure colors and in general bash escape sequences are removed
	msg := stripansi.Strip(str.Trim(string(out.Bytes())))

	mod.client.Privmsg(channel, msg)
}

func (mod *C2) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started")

		for mod.Running() {
			var e session.Event
			select {
			case e = <-mod.eventBus:
				mod.onEvent(e)

			case <-mod.quit:
				mod.Debug("got quit")
				return
			}
		}
	})
}

func (mod *C2) Stop() error {
	return mod.SetRunning(false, func() {
		mod.quit <- true
		mod.Session.Events.Unlisten(mod.eventBus)
		mod.client.Quit()
		mod.client.Disconnect()
	})
}
