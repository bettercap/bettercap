package packet_proxy

import (
	"fmt"
	"io/ioutil"
	golog "log"
	"plugin"
	"strings"
	"syscall"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/session"

	"github.com/chifflier/nfqueue-go/nfqueue"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
)

type PacketProxy struct {
	session.SessionModule
	done       chan bool
	chainName  string
	rule       string
	queue      *nfqueue.Queue
	queueNum   int
	queueCb    nfqueue.Callback
	pluginPath string
	plugin     *plugin.Plugin
}

// this is ugly, but since we can only pass a function
// (not a struct function) as a callback to nfqueue,
// we need this in order to recover the state.
var mod *PacketProxy

func NewPacketProxy(s *session.Session) *PacketProxy {
	mod = &PacketProxy{
		SessionModule: session.NewSessionModule("packet.proxy", s),
		done:          make(chan bool),
		queue:         nil,
		queueCb:       nil,
		queueNum:      0,
		chainName:     "OUTPUT",
	}

	mod.AddHandler(session.NewModuleHandler("packet.proxy on", "",
		"Start the NFQUEUE based packet proxy.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("packet.proxy off", "",
		"Stop the NFQUEUE based packet proxy.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddParam(session.NewIntParameter("packet.proxy.queue.num",
		"0",
		"NFQUEUE number to bind to."))

	mod.AddParam(session.NewStringParameter("packet.proxy.chain",
		"OUTPUT",
		"",
		"Chain name of the iptables rule."))

	mod.AddParam(session.NewStringParameter("packet.proxy.plugin",
		"",
		"",
		"Go plugin file to load and call for every packet."))

	mod.AddParam(session.NewStringParameter("packet.proxy.rule",
		"",
		"",
		"Any additional iptables rule to make the queue more selective (ex. --destination 8.8.8.8)."))

	return mod
}

func (mod PacketProxy) Name() string {
	return "packet.proxy"
}

func (mod PacketProxy) Description() string {
	return "A Linux only module that relies on NFQUEUEs in order to filter packets."
}

func (mod PacketProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *PacketProxy) destroyQueue() {
	if mod.queue == nil {
		return
	}

	mod.queue.DestroyQueue()
	mod.queue.Close()
	mod.queue = nil
}

func (mod *PacketProxy) runRule(enable bool) (err error) {
	action := "-A"
	if !enable {
		action = "-D"
	}

	args := []string{
		action, mod.chainName,
	}

	if mod.rule != "" {
		rule := strings.Split(mod.rule, " ")
		args = append(args, rule...)
	}

	args = append(args, []string{
		"-j", "NFQUEUE",
		"--queue-num", fmt.Sprintf("%d", mod.queueNum),
		"--queue-bypass",
	}...)

	mod.Debug("iptables %s", args)

	_, err = core.Exec("iptables", args)
	return
}

func (mod *PacketProxy) Configure() (err error) {
	golog.SetOutput(ioutil.Discard)

	mod.destroyQueue()

	if err, mod.queueNum = mod.IntParam("packet.proxy.queue.num"); err != nil {
		return
	} else if err, mod.chainName = mod.StringParam("packet.proxy.chain"); err != nil {
		return
	} else if err, mod.rule = mod.StringParam("packet.proxy.rule"); err != nil {
		return
	} else if err, mod.pluginPath = mod.StringParam("packet.proxy.plugin"); err != nil {
		return
	}

	if mod.pluginPath == "" {
		return fmt.Errorf("The parameter %s can not be empty.", tui.Bold("packet.proxy.plugin"))
	} else if !fs.Exists(mod.pluginPath) {
		return fmt.Errorf("%s does not exist.", mod.pluginPath)
	}

	mod.Info("loading packet proxy plugin from %s ...", mod.pluginPath)

	var ok bool
	var sym plugin.Symbol

	if mod.plugin, err = plugin.Open(mod.pluginPath); err != nil {
		return
	} else if sym, err = mod.plugin.Lookup("OnPacket"); err != nil {
		return
	} else if mod.queueCb, ok = sym.(func(*nfqueue.Payload) int); !ok {
		return fmt.Errorf("Symbol OnPacket is not a valid callback function.")
	}

	if sym, err = mod.plugin.Lookup("OnStart"); err == nil {
		var onStartCb func() int
		if onStartCb, ok = sym.(func() int); !ok {
			return fmt.Errorf("OnStart signature does not match expected signature: 'func() int'")
		} else {
			var result int
			if result = onStartCb(); result != 0 {
				return fmt.Errorf("OnStart returned non-zero result. result=%d", result)
			}
		}
	}

	mod.queue = new(nfqueue.Queue)
	if err = mod.queue.SetCallback(dummyCallback); err != nil {
		return
	} else if err = mod.queue.Init(); err != nil {
		return
	} else if err = mod.queue.Unbind(syscall.AF_INET); err != nil {
		return
	} else if err = mod.queue.Bind(syscall.AF_INET); err != nil {
		return
	} else if err = mod.queue.CreateQueue(mod.queueNum); err != nil {
		return
	} else if err = mod.queue.SetMode(nfqueue.NFQNL_COPY_PACKET); err != nil {
		return
	} else if err = mod.runRule(true); err != nil {
		return
	}

	return nil
}

// we need this because for some reason we can't directly
// pass the symbol loaded from the plugin as a direct
// CGO callback ... ¯\_(ツ)_/¯
func dummyCallback(payload *nfqueue.Payload) int {
	return mod.queueCb(payload)
}

func (mod *PacketProxy) Start() error {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started on queue number %d", mod.queueNum)

		defer mod.destroyQueue()

		mod.queue.Loop()

		mod.done <- true
	})
}

func (mod *PacketProxy) Stop() (err error) {
	return mod.SetRunning(false, func() {
		mod.queue.StopLoop()
		mod.runRule(false)

		var sym plugin.Symbol
		if sym, err = mod.plugin.Lookup("OnStop"); err == nil {
			var onStopCb func()
			var ok bool
			if onStopCb, ok = sym.(func()); !ok {
				mod.Error("OnStop signature does not match expected signature: 'func()', unable to call OnStop.")
			} else {
				onStopCb()
			}
		}

		<-mod.done
	})
}
