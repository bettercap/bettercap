package session

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/evilsocket/readline"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/firewall"
	bnet "github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
)

const HistoryFile = "~/bettercap.history"

var (
	I = (*Session)(nil)

	ErrAlreadyStarted = errors.New("Module is already running.")
	ErrAlreadyStopped = errors.New("Module is not running.")
)

type Session struct {
	Options   core.Options             `json:"options"`
	Interface *bnet.Endpoint           `json:"interface"`
	Gateway   *bnet.Endpoint           `json:"gateway"`
	Firewall  firewall.FirewallManager `json:"-"`
	Env       *Environment             `json:"env"`
	Targets   *Targets                 `json:"targets"`
	Queue     *packets.Queue           `json:"packets"`
	Input     *readline.Instance       `json:"-"`
	StartedAt time.Time                `json:"started_at"`
	Active    bool                     `json:"active"`
	Prompt    Prompt                   `json:"-"`

	CoreHandlers []CommandHandler `json:"-"`
	Modules      []Module         `json:"-"`
	HelpPadding  int              `json:"-"`

	Events *EventPool `json:"-"`
}

func ParseCommands(buffer string) []string {
	cmds := make([]string, 0)
	for _, cmd := range strings.Split(buffer, ";") {
		cmd = core.Trim(cmd)
		if cmd != "" || (len(cmd) > 0 && cmd[0] != '#') {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func New() (*Session, error) {
	var err error

	s := &Session{
		Prompt: NewPrompt(),
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

	if *s.Options.CpuProfile != "" {
		if f, err := os.Create(*s.Options.CpuProfile); err != nil {
			return nil, err
		} else if err := pprof.StartCPUProfile(f); err != nil {
			return nil, err
		}
	}

	s.Env = NewEnvironment(s)
	s.Events = NewEventPool(*s.Options.Debug, *s.Options.Silent)

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

func (s *Session) setupReadline() error {
	var err error

	pcompleters := make([]readline.PrefixCompleterInterface, 0)
	for _, h := range s.CoreHandlers {
		if h.Completer == nil {
			pcompleters = append(pcompleters, readline.PcItem(h.Name))
		} else {
			pcompleters = append(pcompleters, h.Completer)
		}
	}

	tree := make(map[string][]string, 0)

	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			parts := strings.Split(h.Name, " ")
			name := parts[0]

			if _, found := tree[name]; found == false {
				tree[name] = []string{}
			}

			tree[name] = append(tree[name], parts[1:]...)
		}
	}

	for root, subElems := range tree {
		item := readline.PcItem(root)
		item.Children = []readline.PrefixCompleterInterface{}

		for _, child := range subElems {
			item.Children = append(item.Children, readline.PcItem(child))
		}

		pcompleters = append(pcompleters, item)
	}

	history := ""
	if *s.Options.NoHistory == false {
		history, _ = core.ExpandPath(HistoryFile)
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
	if *s.Options.Debug {
		s.Events.Add("session.closing", nil)
	}

	for _, m := range s.Modules {
		if m.Running() {
			m.Stop()
		}
	}

	s.Firewall.Restore()
	s.Queue.Stop()

	if *s.Options.CpuProfile != "" {
		pprof.StopCPUProfile()
	}

	if *s.Options.MemProfile != "" {
		f, err := os.Create(*s.Options.MemProfile)
		if err != nil {
			fmt.Printf("Could not create memory profile: %s\n", err)
			return
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("Could not write memory profile: %s\n", err)
		}
	}
}

func (s *Session) Register(mod Module) error {
	s.Modules = append(s.Modules, mod)

	for _, h := range mod.Handlers() {
		if len(h.Name) > s.HelpPadding {
			s.HelpPadding = len(h.Name)
		}
	}

	for _, p := range mod.Parameters() {
		if len(p.Name) > s.HelpPadding {
			s.HelpPadding = len(p.Name)
		}
	}

	return nil
}

func (s *Session) startNetMon() {
	// keep reading network events in order to add / update endpoints
	go func() {
		for event := range s.Queue.Activities {
			if s.Active == false {
				return
			}

			if s.IsOn("net.recon") == true && event.Source == true {
				addr := event.IP.String()
				mac := event.MAC.String()

				existing := s.Targets.AddIfNew(addr, mac)
				if existing != nil {
					existing.LastSeen = time.Now()
				}
			}
		}
	}()
}

func (s *Session) setupSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		s.Events.Log(core.WARNING, "Got SIGTERM")
		s.Close()
		os.Exit(0)
	}()
}

func (s *Session) setupEnv() {
	s.Env.Set(PromptVariable, DefaultPrompt)
	s.Env.Set("iface.index", fmt.Sprintf("%d", s.Interface.Index))
	s.Env.Set("iface.name", s.Interface.Name())
	s.Env.Set("iface.ipv4", s.Interface.IpAddress)
	s.Env.Set("iface.ipv6", s.Interface.Ip6Address)
	s.Env.Set("iface.mac", s.Interface.HwAddress)
	s.Env.Set("gateway.address", s.Gateway.IpAddress)
	s.Env.Set("gateway.mac", s.Gateway.HwAddress)
}

func (s *Session) Start() error {
	var err error

	// make sure modules are always sorted by name
	sort.Slice(s.Modules, func(i, j int) bool {
		return s.Modules[i].Name() < s.Modules[j].Name()
	})

	bnet.OuiInit()

	if s.Interface, err = bnet.FindInterface(*s.Options.InterfaceName); err != nil {
		return err
	}

	if s.Queue, err = packets.NewQueue(s.Interface); err != nil {
		return err
	}

	if s.Gateway, err = bnet.FindGateway(s.Interface); err != nil {
		s.Events.Log(core.WARNING, "%s", err.Error())
	}

	if s.Gateway == nil || s.Gateway.IpAddress == s.Interface.IpAddress {
		s.Gateway = s.Interface
	}

	s.Targets = NewTargets(s, s.Interface, s.Gateway)
	s.Firewall = firewall.Make(s.Interface)

	s.setupEnv()

	if err := s.setupReadline(); err != nil {
		return err
	}

	s.setupSignals()

	s.StartedAt = time.Now()
	s.Active = true

	s.startNetMon()

	if *s.Options.Debug {
		s.Events.Add("session.started", nil)
	}

	return nil
}

func (s *Session) Skip(ip net.IP) bool {
	if ip.IsLoopback() == true {
		return true
	} else if bytes.Compare(ip, s.Interface.IP) == 0 {
		return true
	} else if bytes.Compare(ip, s.Gateway.IP) == 0 {
		return true
	}
	return false
}

func (s *Session) IsOn(moduleName string) bool {
	for _, m := range s.Modules {
		if m.Name() == moduleName {
			return m.Running()
		}
	}
	return false
}

func (s *Session) Refresh() {
	s.Input.SetPrompt(s.Prompt.Render(s))
	s.Input.Refresh()
}

func (s *Session) ReadLine() (string, error) {
	s.Refresh()
	return s.Input.Readline()
}

func (s *Session) RunCaplet(filename string) error {
	s.Events.Log(core.INFO, "Reading from caplet %s ...", filename)

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
	line = core.TrimRight(line)
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
