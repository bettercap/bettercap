package session

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/bettercap/readline"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/adrianmo/go-nmea"
)

const (
	HistoryFile = "~/bettercap.history"
)

var (
	I = (*Session)(nil)

	ErrAlreadyStarted = errors.New("Module is already running.")
	ErrAlreadyStopped = errors.New("Module is not running.")
	ErrNotSupported   = errors.New("This component is not supported on this OS.")

	reCmdSpaceCleaner = regexp.MustCompile(`^([^\s]+)\s+(.+)$`)
	reEnvVarCapture   = regexp.MustCompile(`{env\.([^}]+)}`)
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

	if s.Env, err = NewEnvironment(*s.Options.EnvFile); err != nil {
		return nil, err
	}

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

func (s *Session) Close() {
	if *s.Options.Debug {
		fmt.Printf("\nStopping modules and cleaning session state ...\n")
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
	if ip.IsLoopback() {
		return true
	} else if ip.Equal(s.Interface.IP) {
		return true
	} else if ip.Equal(s.Gateway.IP) {
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
	p, _ := s.parseEnvTokens(s.Prompt.Render(s))
	s.Input.SetPrompt(p)
	s.Input.Refresh()
}

func (s *Session) ReadLine() (string, error) {
	s.Refresh()
	return s.Input.Readline()
}

func (s *Session) RunCaplet(filename string) error {
	err, caplet := LoadCaplet(filename)
	if err != nil {
		return err
	}

	return caplet.Eval(s, nil)
}

func (s *Session) Run(line string) error {
	line = core.TrimRight(line)
	// remove extra spaces after the first command
	// so that 'arp.spoof      on' is normalized
	// to 'arp.spoof on' (fixes #178)
	line = reCmdSpaceCleaner.ReplaceAllString(line, "$1 $2")

	// replace all {env.something} with their values
	line, err := s.parseEnvTokens(line)
	if err != nil {
		return err
	}

	// is it a core command?
	for _, h := range s.CoreHandlers {
		if parsed, args := h.Parse(line); parsed {
			return h.Exec(args, s)
		}
	}

	// is it a module command?
	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			if parsed, args := h.Parse(line); parsed {
				return h.Exec(args)
			}
		}
	}

	// is it a caplet command?
	if parsed, caplet, argv := parseCapletCommand(line); parsed {
		return caplet.Eval(s, argv)
	}

	// is it a proxy module custom command?
	if s.UnkCmdCallback != nil && s.UnkCmdCallback(line) {
		return nil
	}

	return fmt.Errorf("unknown or invalid syntax \"%s%s%s\", type %shelp%s for the help menu.", core.BOLD, line, core.RESET, core.BOLD, core.RESET)
}
