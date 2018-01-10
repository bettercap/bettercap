package session

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"sort"
	"strings"
	"syscall"

	"github.com/chzyer/readline"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/firewall"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
)

var (
	I = (*Session)(nil)
)

type Session struct {
	Options   core.Options             `json:"options"`
	Interface *net.Endpoint            `json:"interface"`
	Gateway   *net.Endpoint            `json:"gateway"`
	Firewall  firewall.FirewallManager `json:"-"`
	Env       *Environment             `json:"env"`
	Targets   *Targets                 `json:"targets"`
	Queue     *packets.Queue           `json:"-"`
	Input     *readline.Instance       `json:"-"`
	Active    bool                     `json:"active"`

	CoreHandlers []CommandHandler `json:"-"`
	Modules      []Module         `json:"-"`
	HelpPadding  int              `json:"-"`

	Events *EventPool `json:"-"`
}

func New() (*Session, error) {
	var err error

	s := &Session{
		Env:    nil,
		Active: false,
		Queue:  nil,

		CoreHandlers: make([]CommandHandler, 0),
		Modules:      make([]Module, 0),
		HelpPadding:  0,

		Events: nil,
	}

	if s.Options, err = core.ParseOptions(); err != nil {
		return nil, err
	}

	s.Env = NewEnvironment(s)
	s.Events = NewEventPool(*s.Options.Debug, *s.Options.Silent)

	if u, err := user.Current(); err != nil {
		return nil, err
	} else if u.Uid != "0" {
		return nil, fmt.Errorf("This software must run as root.")
	}

	s.registerCoreHandlers()

	if I == nil {
		I = s
	}

	return s, nil
}

func (s *Session) Module(name string) (err error, mod Module) {
	for _, m := range s.Modules {
		if m.Name() == name {
			return nil, m
		}
	}
	return fmt.Errorf("Module %s not found", name), mod
}

func (s *Session) setupInput() error {
	var err error

	pcompleters := make([]readline.PrefixCompleterInterface, 0)
	for _, h := range s.CoreHandlers {
		pcompleters = append(pcompleters, readline.PcItem(h.Name))
	}

	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			pcompleters = append(pcompleters, readline.PcItem(h.Name))
		}
	}

	history := ""
	if *s.Options.NoHistory == false {
		history = "bettercap.history"
	}

	cfg := readline.Config{
		HistoryFile:       history,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		AutoComplete:      readline.NewPrefixCompleter(pcompleters...),
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			// block CtrlZ feature
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
	}

	s.Input, err = readline.NewEx(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) Close() {
	s.Events.Add("session.closing", nil)

	for _, m := range s.Modules {
		if m.Running() {
			m.Stop()
		}
	}

	s.Firewall.Restore()
	s.Queue.Stop()
}

func (s *Session) Register(mod Module) error {
	s.Modules = append(s.Modules, mod)
	return nil
}

func (s *Session) Start() error {
	var err error

	// make sure modules are always sorted by name
	sort.Slice(s.Modules, func(i, j int) bool {
		return s.Modules[i].Name() < s.Modules[j].Name()
	})

	net.OuiInit()

	if s.Interface, err = net.FindInterface(*s.Options.InterfaceName); err != nil {
		return err
	}

	s.Env.Set("iface.name", s.Interface.Name())
	s.Env.Set("iface.address", s.Interface.IpAddress)
	s.Env.Set("iface.mac", s.Interface.HwAddress)

	if s.Queue, err = packets.NewQueue(s.Interface.Name()); err != nil {
		return err
	}

	if s.Gateway, err = net.FindGateway(s.Interface); err != nil {
		s.Events.Log(WARNING, "%s", err.Error())
	}

	if s.Gateway == nil || s.Gateway.IpAddress == s.Interface.IpAddress {
		s.Gateway = s.Interface
	}

	s.Env.Set("gateway.address", s.Gateway.IpAddress)
	s.Env.Set("gateway.mac", s.Gateway.HwAddress)

	s.Targets = NewTargets(s, s.Interface, s.Gateway)
	s.Firewall = firewall.Make()

	if err := s.setupInput(); err != nil {
		return err
	}

	for _, h := range s.CoreHandlers {
		if len(h.Name) > s.HelpPadding {
			s.HelpPadding = len(h.Name)
		}
	}
	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			if len(h.Name) > s.HelpPadding {
				s.HelpPadding = len(h.Name)
			}
		}

		for _, p := range m.Parameters() {
			if len(p.Name) > s.HelpPadding {
				s.HelpPadding = len(p.Name)
			}
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		s.Events.Log(WARNING, "Got SIGTERM")
		s.Close()
		os.Exit(0)
	}()

	s.Active = true
	s.Events.Add("session.started", nil)

	return nil
}

func (s *Session) ReadLine() (string, error) {
	prompt := FG_WHITE + BG_YELLOW + " " + s.Interface.CIDR() +
		FG_BLACK +
		" > " +
		s.Interface.IpAddress +
		" " + core.RESET +
		BOLD + " » " + RESET

	s.Input.SetPrompt(prompt)
	s.Input.Refresh()
	return s.Input.Readline()
}

func (s *Session) RunCaplet(filename string) error {
	s.Events.Log(INFO, "Reading from caplet %s ...", filename)

	input, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer input.Close()

	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}

		if err = s.Run(line); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) Run(line string) error {
	line = strings.TrimRight(line, " ")
	for _, h := range s.CoreHandlers {
		if parsed, args := h.Parse(line); parsed == true {
			return h.Exec(args, s)
		}
	}

	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			if parsed, args := h.Parse(line); parsed == true {
				return h.Exec(args)
			}
		}
	}

	return fmt.Errorf("Unknown command %s%s%s, type %shelp%s for the help menu.", core.BOLD, line, core.RESET, core.BOLD, core.RESET)
}
