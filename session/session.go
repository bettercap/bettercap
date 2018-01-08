package session

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/op/go-logging"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/firewall"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
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

		Events: NewEventPool(),
	}

	s.Env = NewEnvironment(s)

	if s.Options, err = core.ParseOptions(); err != nil {
		return nil, err
	}

	if u, err := user.Current(); err != nil {
		return nil, err
	} else if u.Uid != "0" {
		return nil, fmt.Errorf("This software must run as root.")
	}

	// setup logging
	if *s.Options.Debug == true {
		logging.SetLevel(logging.DEBUG, "")
	} else if *s.Options.Silent == true {
		logging.SetLevel(logging.ERROR, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}

	s.registerCoreHandlers()

	return s, nil
}

func (s *Session) registerCoreHandlers() {
	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("help", "^(help|\\?)$",
		"Display list of available commands.",
		func(args []string, s *Session) error {
			fmt.Println()
			fmt.Printf("Basic commands:\n\n")
			for _, h := range s.CoreHandlers {
				fmt.Printf("  "+core.Bold("%"+strconv.Itoa(s.HelpPadding)+"s")+" : %s\n", h.Name, h.Description)
			}

			sort.Slice(s.Modules, func(i, j int) bool {
				return s.Modules[i].Name() < s.Modules[j].Name()
			})

			for _, m := range s.Modules {
				fmt.Println()
				status := ""
				if m.Running() {
					status = core.Green("active")
				} else {
					status = core.Red("not active")
				}
				fmt.Printf("%s [%s]\n", m.Name(), status)
				fmt.Println(core.Dim(m.Description()) + "\n")
				for _, h := range m.Handlers() {
					fmt.Printf(h.Help(s.HelpPadding))
				}

				params := m.Parameters()
				if len(params) > 0 {
					fmt.Printf("\n  Parameters\n\n")
					for _, p := range params {
						fmt.Printf(p.Help(s.HelpPadding))
					}
					fmt.Println()
				}
			}

			return nil
		}))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("active", "^active$",
		"Show information about active modules.",
		func(args []string, s *Session) error {
			for _, m := range s.Modules {
				if m.Running() == false {
					continue
				}
				fmt.Printf("[%s] %s (%s)\n", core.Green("active"), m.Name(), core.Dim(m.Description()))
				params := m.Parameters()
				if len(params) > 0 {
					for _, p := range params {
						_, p.Value = s.Env.Get(p.Name)
						fmt.Printf("  %s: '%s'\n", p.Name, core.Yellow(p.Value))
					}
					fmt.Println()
				}
			}

			return nil
		}))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("exit", "^(q|quit|e|exit)$",
		"Close the session and exit.",
		func(args []string, s *Session) error {
			s.Active = false
			s.Input.Close()
			return nil
		}))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("sleep SECONDS", "^sleep\\s+(\\d+)$",
		"Sleep for the given amount of seconds.",
		func(args []string, s *Session) error {
			if secs, err := strconv.Atoi(args[0]); err == nil {
				time.Sleep(time.Duration(secs) * time.Second)
				return nil
			} else {
				return err
			}
		}))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("get NAME", "^get\\s+(.+)",
		"Get the value of variable NAME, use * for all.",
		func(args []string, s *Session) error {
			key := args[0]
			if key == "*" {
				prev_ns := ""

				fmt.Println()
				for _, k := range s.Env.Sorted() {
					ns := ""
					toks := strings.Split(k, ".")
					if len(toks) > 0 {
						ns = toks[0]
					}

					if ns != prev_ns {
						fmt.Println()
						prev_ns = ns
					}

					fmt.Printf("  %"+strconv.Itoa(s.Env.Padding)+"s: '%s'\n", k, s.Env.Storage[k])
				}
				fmt.Println()
			} else if found, value := s.Env.Get(key); found == true {
				fmt.Println()
				fmt.Printf("  %s: '%s'\n", key, value)
				fmt.Println()
			} else {
				return fmt.Errorf("%s not found", key)
			}

			return nil
		}))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("set NAME VALUE", "^set\\s+([^\\s]+)\\s+(.+)",
		"Set the VALUE of variable NAME.",
		func(args []string, s *Session) error {
			key := args[0]
			value := args[1]

			if value == "\"\"" {
				value = ""
			}

			s.Env.Set(key, value)
			fmt.Printf("  %s => '%s'\n", core.Green(key), core.Yellow(value))
			return nil
		}))
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

	// now that we have the readline instance, we can set logging to its
	// console writer so the whole thing gets correctly updated when something
	// is logged to screen (it won't overlap with the prompt).
	log_be := logging.NewLogBackend(s.Input.Stderr(), "", 0)
	log_level := logging.AddModuleLevel(log_be)
	if *s.Options.Debug == true {
		log_level.SetLevel(logging.DEBUG, "")
	} else if *s.Options.Silent == true {
		log_level.SetLevel(logging.ERROR, "")
	} else {
		log_level.SetLevel(logging.INFO, "")
	}

	logging.SetBackend(log_level)

	return nil
}

func (s *Session) Close() {
	s.Events.Add("session.closing", nil)

	for _, m := range s.Modules {
		m.OnSessionEnded(s)
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

	log.Debugf("[%s%s%s] %s\n", core.GREEN, s.Interface.Name(), core.RESET, s.Interface)
	log.Debugf("[%ssubnet%s] %s\n", core.GREEN, core.RESET, s.Interface.CIDR())

	if s.Gateway, err = net.FindGateway(s.Interface); err != nil {
		log.Warningf("%s\n", err)
	}

	if s.Gateway == nil || s.Gateway.IpAddress == s.Interface.IpAddress {
		s.Gateway = s.Interface
	}

	s.Env.Set("gateway.address", s.Gateway.IpAddress)
	s.Env.Set("gateway.mac", s.Gateway.HwAddress)

	log.Debugf("[%sgateway%s] %s\n", core.GREEN, core.RESET, s.Gateway)

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
		log.Warning("Got SIGTERM ...")
		s.Close()
		os.Exit(0)
	}()

	s.Active = true

	for _, m := range s.Modules {
		m.OnSessionStarted(s)
	}

	s.Events.Add("session.started", nil)

	return nil
}

func (s *Session) ReadLine() (string, error) {
	s.Input.SetPrompt(core.GREEN + s.Interface.IpAddress + core.RESET + "» ")
	s.Input.Refresh()
	return s.Input.Readline()
}

func (s *Session) RunCaplet(filename string) error {
	log.Infof("Reading from caplet %s ...\n", filename)

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
