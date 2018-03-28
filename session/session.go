package session

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bettercap/readline"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/adrianmo/go-nmea"
)

const HistoryFile = "~/bettercap.history"

var (
	I = (*Session)(nil)

	ErrAlreadyStarted = errors.New("Module is already running.")
	ErrAlreadyStopped = errors.New("Module is not running.")
	ErrNotSupported   = errors.New("This component is not supported on this OS.")

	reCmdSpaceCleaner = regexp.MustCompile(`^([^\s]+)\s+(.+)$`)
)

type UnknownCommandCallback func(cmd string) bool

type Session struct {
	Options   core.Options             `json:"options"`
	Interface *network.Endpoint        `json:"interface"`
	Gateway   *network.Endpoint        `json:"gateway"`
	Firewall  firewall.FirewallManager `json:"-"`
	Env       *Environment             `json:"env"`
	Lan       *network.LAN             `json:"lan"`
	WiFi      *network.WiFi            `json:"wifi"`
	BLE       *network.BLE             `json:"ble"`
	Queue     *packets.Queue           `json:"packets"`
	Input     *readline.Instance       `json:"-"`
	StartedAt time.Time                `json:"started_at"`
	Active    bool                     `json:"active"`
	GPS       nmea.GNGGA               `json:"gps"`
	Prompt    Prompt                   `json:"-"`

	CoreHandlers []CommandHandler `json:"-"`
	Modules      []Module         `json:"-"`

	Events *EventPool `json:"-"`

	UnkCmdCallback UnknownCommandCallback `json:"-"`
}

func ParseCommands(line string) []string {
	args := []string{}
	buf := ""

	singleQuoted := false
	doubleQuoted := false
	finish := false

	for _, c := range line {
		switch c {
		case ';':
			if singleQuoted == false && doubleQuoted == false {
				finish = true
			} else {
				buf += string(c)
			}

		case '"':
			if doubleQuoted {
				// finish of quote
				doubleQuoted = false
			} else if singleQuoted {
				// quote initiated with ', so we ignore it
				buf += string(c)
			} else {
				// quote init here
				doubleQuoted = true
			}

		case '\'':
			if singleQuoted {
				singleQuoted = false
			} else if doubleQuoted {
				buf += string(c)
			} else {
				singleQuoted = true
			}

		default:
			buf += string(c)
		}

		if finish {
			args = append(args, buf)
			finish = false
			buf = ""
		}
	}

	if len(buf) > 0 {
		args = append(args, buf)
	}

	cmds := make([]string, 0)
	for _, cmd := range args {
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

		CoreHandlers:   make([]CommandHandler, 0),
		Modules:        make([]Module, 0),
		Events:         nil,
		UnkCmdCallback: nil,
	}

	if s.Options, err = core.ParseOptions(); err != nil {
		return nil, err
	}

	core.InitSwag(*s.Options.NoColors)

	if *s.Options.CpuProfile != "" {
		if f, err := os.Create(*s.Options.CpuProfile); err != nil {
			return nil, err
		} else if err := pprof.StartCPUProfile(f); err != nil {
			return nil, err
		}
	}

	s.Env = NewEnvironment(s, *s.Options.EnvFile)
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

			var appendedOption = strings.Join(parts[1:], " ")

			if len(appendedOption) > 0 {
				tree[name] = append(tree[name], appendedOption)
			}
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
		HistoryFile:     history,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    readline.NewPrefixCompleter(pcompleters...),
	}

	s.Input, err = readline.NewEx(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) Close() {
	fmt.Printf("\nStopping modules and cleaning session state ...\n")

	if *s.Options.Debug {
		s.Events.Add("session.closing", nil)
	}

	for _, m := range s.Modules {
		if m.Running() {
			m.Stop()
		}
	}

	s.Firewall.Restore()

	if *s.Options.EnvFile != "" {
		envFile, _ := core.ExpandPath(*s.Options.EnvFile)
		if err := s.Env.Save(envFile); err != nil {
			fmt.Printf("Error while storing the environment to %s: %s", envFile, err)
		}
	}

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

				existing := s.Lan.AddIfNew(addr, mac)
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
	s.Env.Set("iface.index", fmt.Sprintf("%d", s.Interface.Index))
	s.Env.Set("iface.name", s.Interface.Name())
	s.Env.Set("iface.ipv4", s.Interface.IpAddress)
	s.Env.Set("iface.ipv6", s.Interface.Ip6Address)
	s.Env.Set("iface.mac", s.Interface.HwAddress)
	s.Env.Set("gateway.address", s.Gateway.IpAddress)
	s.Env.Set("gateway.mac", s.Gateway.HwAddress)

	if found, v := s.Env.Get(PromptVariable); found == false || v == "" {
		s.Env.Set(PromptVariable, DefaultPrompt)
	}

	dbg := "false"
	if *s.Options.Debug {
		dbg = "true"
	}
	s.Env.WithCallback("log.debug", dbg, func(newValue string) {
		newDbg := false
		if newValue == "true" {
			newDbg = true
		}
		s.Events.SetDebug(newDbg)
	})

	silent := "false"
	if *s.Options.Silent {
		silent = "true"
	}
	s.Env.WithCallback("log.silent", silent, func(newValue string) {
		newSilent := false
		if newValue == "true" {
			newSilent = true
		}
		s.Events.SetSilent(newSilent)
	})
}

func (s *Session) Start() error {
	var err error

	// make sure modules are always sorted by name
	sort.Slice(s.Modules, func(i, j int) bool {
		return s.Modules[i].Name() < s.Modules[j].Name()
	})

	if s.Interface, err = network.FindInterface(*s.Options.InterfaceName); err != nil {
		return err
	}

	if s.Queue, err = packets.NewQueue(s.Interface); err != nil {
		return err
	}

	if s.Gateway, err = network.FindGateway(s.Interface); err != nil {
		s.Events.Log(core.WARNING, "%s", err.Error())
	}

	if s.Gateway == nil || s.Gateway.IpAddress == s.Interface.IpAddress {
		s.Gateway = s.Interface
	}

	s.Firewall = firewall.Make(s.Interface)

	s.BLE = network.NewBLE(func(dev *network.BLEDevice) {
		s.Events.Add("ble.device.new", dev)
	}, func(dev *network.BLEDevice) {
		s.Events.Add("ble.device.lost", dev)
	})

	s.WiFi = network.NewWiFi(s.Interface, func(ap *network.AccessPoint) {
		s.Events.Add("wifi.ap.new", ap)
	}, func(ap *network.AccessPoint) {
		s.Events.Add("wifi.ap.lost", ap)
	})

	s.Lan = network.NewLAN(s.Interface, s.Gateway, func(e *network.Endpoint) {
		s.Events.Add("endpoint.new", e)
	}, func(e *network.Endpoint) {
		s.Events.Add("endpoint.lost", e)
	})

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

func (s *Session) isCapletCommand(line string) (is bool, filename string, argv []string) {
	paths := []string{
		"./",
		"./caplets/",
	}

	capspath := core.Trim(os.Getenv("CAPSPATH"))
	for _, folder := range core.SepSplit(capspath, ":") {
		paths = append(paths, folder)
	}

	file := core.Trim(line)
	parts := strings.Split(file, " ")
	argc := len(parts)
	argv = make([]string, 0)
	// check for any arguments
	if argc > 1 {
		file = core.Trim(parts[0])
		if argc >= 2 {
			argv = parts[1:]
		}
	}

	for _, path := range paths {
		filename := filepath.Join(path, file) + ".cap"
		if core.Exists(filename) {
			return true, filename, argv
		}
	}

	return false, "", nil
}

func (s *Session) runCapletCommand(filename string, argv []string) error {
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

		// replace $0 with argv[0], $1 with argv[1] and so on
		for i, arg := range argv {
			line = strings.Replace(line, fmt.Sprintf("$%d", i), arg, -1)
		}

		if err = s.Run(line); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) Run(line string) error {
	line = core.TrimRight(line)
	// remove extra spaces after the first command
	// so that 'arp.spoof      on' is normalized
	// to 'arp.spoof on' (fixes #178)
	line = reCmdSpaceCleaner.ReplaceAllString(line, "$1 $2")

	// is it a core command?
	for _, h := range s.CoreHandlers {
		if parsed, args := h.Parse(line); parsed == true {
			return h.Exec(args, s)
		}
	}

	// is it a module command?
	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			if parsed, args := h.Parse(line); parsed == true {
				return h.Exec(args)
			}
		}
	}

	// is it a caplet command?
	if is, filename, argv := s.isCapletCommand(line); is {
		return s.runCapletCommand(filename, argv)
	}

	// is it a proxy module custom command?
	if s.UnkCmdCallback != nil && s.UnkCmdCallback(line) == true {
		return nil
	}

	return fmt.Errorf("Unknown or invalid syntax \"%s%s%s\", type %shelp%s for the help menu.", core.BOLD, line, core.RESET, core.BOLD, core.RESET)
}
