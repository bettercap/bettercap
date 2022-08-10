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
	"strings"
	"time"

	"github.com/bettercap/readline"

	"github.com/bettercap/bettercap/caplets"
	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/evilsocket/islazy/data"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/plugin"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

const (
	HistoryFile = "~/bettercap.history"
)

var (
	I = (*Session)(nil)

	ErrNotSupported = errors.New("this component is not supported on this OS")

	reCmdSpaceCleaner = regexp.MustCompile(`^([^\s]+)\s+(.+)$`)
	reEnvVarCapture   = regexp.MustCompile(`{env\.([^}]+)}`)
)

func ErrAlreadyStarted(name string) error {
	return fmt.Errorf("module %s is already running", name)
}

func ErrAlreadyStopped(name string) error {
	return fmt.Errorf("module %s is not running", name)
}

type UnknownCommandCallback func(cmd string) bool

type GPS struct {
	Updated       time.Time
	Latitude      float64 // Latitude.
	Longitude     float64 // Longitude.
	FixQuality    string  // Quality of fix.
	NumSatellites int64   // Number of satellites in use.
	HDOP          float64 // Horizontal dilution of precision.
	Altitude      float64 // Altitude.
	Separation    float64 // Geoidal separation
}

const AliasesFile = "~/bettercap.aliases"

var aliasesFileName, _ = fs.Expand(AliasesFile)

type Session struct {
	Options   core.Options
	Interface *network.Endpoint
	Gateway   *network.Endpoint
	Env       *Environment
	Lan       *network.LAN
	WiFi      *network.WiFi
	BLE       *network.BLE
	HID       *network.HID
	Queue     *packets.Queue
	StartedAt time.Time
	Active    bool
	GPS       GPS
	Modules   ModuleList
	Aliases   *data.UnsortedKV

	Input            *readline.Instance
	Prompt           Prompt
	CoreHandlers     []CommandHandler
	Events           *EventPool
	EventsIgnoreList *EventsIgnoreList
	UnkCmdCallback   UnknownCommandCallback
	Firewall         firewall.FirewallManager

	script *Script
}

func New() (*Session, error) {
	opts, err := core.ParseOptions()
	if err != nil {
		return nil, err
	}

	if *opts.NoColors || !tui.Effects() {
		tui.Disable()
		log.NoEffects = true
	}

	s := &Session{
		Prompt:  NewPrompt(),
		Options: opts,
		Env:     nil,
		Active:  false,
		Queue:   nil,

		CoreHandlers:     make([]CommandHandler, 0),
		Modules:          make([]Module, 0),
		Events:           nil,
		EventsIgnoreList: NewEventsIgnoreList(),
		UnkCmdCallback:   nil,
	}

	if *s.Options.CpuProfile != "" {
		if f, err := os.Create(*s.Options.CpuProfile); err != nil {
			return nil, err
		} else if err := pprof.StartCPUProfile(f); err != nil {
			return nil, err
		}
	}

	if bufSize := *s.Options.PcapBufSize; bufSize != -1 {
		network.CAPTURE_DEFAULTS.Bufsize = bufSize
	}

	if s.Env, err = NewEnvironment(*s.Options.EnvFile); err != nil {
		return nil, err
	}

	if s.Aliases, err = data.NewUnsortedKV(aliasesFileName, data.FlushOnEdit); err != nil {
		return nil, err
	}

	s.Events = NewEventPool(*s.Options.Debug, *s.Options.Silent)

	s.registerCoreHandlers()

	if I == nil {
		I = s
	}

	return s, nil
}

func (s *Session) Lock() {
	s.Env.Lock()
	s.Lan.Lock()
	s.WiFi.Lock()
}

func (s *Session) Unlock() {
	s.Env.Unlock()
	s.Lan.Unlock()
	s.WiFi.Unlock()
}

func (s *Session) Module(name string) (err error, mod Module) {
	for _, m := range s.Modules {
		if m.Name() == name {
			return nil, m
		}
	}
	return fmt.Errorf("module %s not found", name), mod
}

func (s *Session) Close() {
	if *s.Options.PrintVersion {
		return
	}

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
		envFile, _ := fs.Expand(*s.Options.EnvFile)
		if err := s.Env.Save(envFile); err != nil {
			fmt.Printf("error while storing the environment to %s: %s", envFile, err)
		}
	}

	if *s.Options.CpuProfile != "" {
		pprof.StopCPUProfile()
	}

	if *s.Options.MemProfile != "" {
		f, err := os.Create(*s.Options.MemProfile)
		if err != nil {
			fmt.Printf("could not create memory profile: %s\n", err)
			return
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("could not write memory profile: %s\n", err)
		}
	}
}

func (s *Session) Register(mod Module) error {
	s.Modules = append(s.Modules, mod)
	return nil
}

func (s *Session) Start() error {
	var err error

	network.Debug = func(format string, args ...interface{}) {
		s.Events.Log(log.DEBUG, format, args...)
	}

	// make sure modules are always sorted by name
	sort.Slice(s.Modules, func(i, j int) bool {
		return s.Modules[i].Name() < s.Modules[j].Name()
	})

	if *s.Options.CapletsPath != "" {
		if err = caplets.Setup(*s.Options.CapletsPath); err != nil {
			return err
		}
	}

	if s.Interface, err = network.FindInterface(*s.Options.InterfaceName); err != nil {
		return err
	}

	if s.Queue, err = packets.NewQueue(s.Interface); err != nil {
		return err
	}

	if *s.Options.Gateway != "" {
		if s.Gateway, err = network.GatewayProvidedByUser(s.Interface, *s.Options.Gateway); err != nil {
			s.Events.Log(log.WARNING, "%s", err.Error())
			s.Gateway, err = network.FindGateway(s.Interface)
		}
	} else {
		s.Gateway, err = network.FindGateway(s.Interface)
	}

	if err != nil {
		level := ops.Ternary(s.Interface.IsMonitor(), log.DEBUG, log.WARNING).(log.Verbosity)
		s.Events.Log(level, "%s", err.Error())
	}

	// we are the gateway
	if s.Gateway == nil || s.Gateway.IpAddress == s.Interface.IpAddress {
		s.Gateway = s.Interface
	} else {
		// start monitoring for gateway changes
		go s.routeMon()
	}

	s.Firewall = firewall.Make(s.Interface)

	s.HID = network.NewHID(s.Aliases, func(dev *network.HIDDevice) {
		s.Events.Add("hid.device.new", dev)
	}, func(dev *network.HIDDevice) {
		s.Events.Add("hid.device.lost", dev)
	})

	s.BLE = network.NewBLE(s.Aliases, func(dev *network.BLEDevice) {
		s.Events.Add("ble.device.new", dev)
	}, func(dev *network.BLEDevice) {
		s.Events.Add("ble.device.lost", dev)
	})

	s.WiFi = network.NewWiFi(s.Interface, s.Aliases, func(ap *network.AccessPoint) {
		s.Events.Add("wifi.ap.new", ap)
	}, func(ap *network.AccessPoint) {
		s.Events.Add("wifi.ap.lost", ap)
	})

	s.Lan = network.NewLAN(s.Interface, s.Gateway, s.Aliases, func(e *network.Endpoint) {
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

	s.Events.Add("session.started", nil)

	// register js functions here to avoid cyclic dependency between
	// js and session
	plugin.Defines["env"] = jsEnvFunc
	plugin.Defines["run"] = jsRunFunc
	plugin.Defines["fileExists"] = jsFileExistsFunc
	plugin.Defines["loadJSON"] = jsLoadJSONFunc
	plugin.Defines["saveJSON"] = jsSaveJSONFunc
	plugin.Defines["onEvent"] = jsOnEventFunc
	plugin.Defines["session"] = s

	// load the script here so the session and its internal objects are ready
	if *s.Options.Script != "" {
		if s.script, err = LoadScript(*s.Options.Script); err != nil {
			return fmt.Errorf("error loading %s: %v", *s.Options.Script, err)
		}
		log.Debug("session script %s loaded", *s.Options.Script)
	}

	return nil
}

func (s *Session) Skip(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	} else if ip.Equal(s.Interface.IP) || ip.Equal(s.Interface.IPv6) {
		return true
	} else if ip.Equal(s.Gateway.IP) {
		return true
	}
	return false
}

func (s *Session) FindMAC(ip net.IP, probe bool) (net.HardwareAddr, error) {
	var mac string
	var hw net.HardwareAddr
	var err error

	// do we have this ip mac address?
	mac, err = network.ArpLookup(s.Interface.Name(), ip.String(), false)
	if err != nil && probe {
		from := s.Interface.IP
		from_hw := s.Interface.HW

		if ip.To4() == nil {
			from = s.Interface.IPv6
		}

		if err, probe := packets.NewUDPProbe(from, from_hw, ip, 139); err != nil {
			log.Error("Error while creating UDP probe packet for %s: %s", ip.String(), err)
		} else {
			s.Queue.Send(probe)
		}

		time.Sleep(500 * time.Millisecond)
		mac, _ = network.ArpLookup(s.Interface.Name(), ip.String(), false)
	}

	if mac == "" {
		return nil, fmt.Errorf("Could not find hardware address for %s.", ip.String())
	}

	mac = network.NormalizeMac(mac)
	hw, err = net.ParseMAC(mac)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing hardware address '%s' for %s: %s", mac, ip.String(), err)
	}
	return hw, nil
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
	caplet, err := caplets.Load(filename)
	if err != nil {
		return err
	}

	return caplet.Eval(nil, func(line string) error {
		return s.Run(line + "\n")
	})
}

func parseCapletCommand(line string) (is bool, caplet *caplets.Caplet, argv []string) {
	file := str.Trim(line)
	parts := strings.Split(file, " ")
	argc := len(parts)
	argv = make([]string, 0)
	// check for any arguments
	if argc > 1 {
		file = str.Trim(parts[0])
		argv = parts[1:]
	}

	if cap, err := caplets.Load(file); err == nil {
		return true, cap, argv
	}

	return false, nil, nil
}

func (s *Session) Run(line string) error {
	line = str.TrimRight(line)
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
		return caplet.Eval(argv, func(line string) error {
			return s.Run(line + "\n")
		})
	}

	// is it a proxy module custom command?
	if s.UnkCmdCallback != nil && s.UnkCmdCallback(line) {
		return nil
	}

	return fmt.Errorf("unknown or invalid syntax \"%s%s%s\", type %shelp%s for the help menu.", tui.BOLD, line, tui.RESET, tui.BOLD, tui.RESET)
}
