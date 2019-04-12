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

type RdpProxy struct {
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

var mod *RdpProxy

func NewRdpProxy(s *session.Session) *RdpProxy {
	mod = &RdpProxy{
		SessionModule: session.NewSessionModule("rdp.proxy", s),
		done:          make(chan bool),
		queue:         nil,
		queueCb:       nil,
		queueNum:      0,
		chainName:     "OUTPUT",
	}

	mod.AddHandler(session.NewModuleHandler("rdp.proxy on", "",
		"Start the RDP proxy.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("rdp.proxy off", "",
		"Stop the RDP proxy.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddParam(session.NewIntParameter("rdp.proxy.queue.num",
		"0",
		"NFQUEUE number to bind to."))

    mod.AddParam(session.NewIntParameter("rdp.proxy.port",
        "3389",
        "RDP port to intercept."
    ))

	// mod.AddParam(session.NewStringParameter("rdp.proxy.targets",
        // session.ParamSubnet,
        // "",
        // "Comma separated list of IP addresses, also supports nmap style IP ranges."))

    // TODO: Should support comma separated subnets
    // mod.AddParam(session.NewStringParameter("rdp.proxy.targets", "3389", session.IPv4RangeValidator "RDP port to intercept."))

	mod.AddParam(session.NewIntParameter("rdp.proxy.start",
		"40000",
		"",
		"Starting port for pyrdp sessionss"))

    mod.AddParam(session.NewStringParameter("rdp.proxy.command",
        "pyrdp-mitm",
        "The PyRDP base command to launch the man-in-the-middle."))

	return mod
}

func (mod RdpProxy) Name() string {
	return "rdp.proxy"
}

func (mod RdpProxy) Description() string {
	return "A Linux only module that relies on NFQUEUEs in order to man-in-the-middle RDP sessions."
}

func (mod RdpProxy) Author() string {
	return "Alexandre Beaulieu <alex@segfault.me>"
}

func (mod *RdpProxy) destroyQueue() {
	if mod.queue == nil {
		return
	}

	mod.queue.DestroyQueue()
	mod.queue.Close()
	mod.queue = nil
}

func (mod *RdpProxy) runRule(enable bool) (err error) {
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

func (mod *RdpProxy) Configure() (err error) {
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

func (mod *RdpProxy) Start() error {
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

func (mod *RdpProxy) Stop() error {
	return mod.SetRunning(false, func() {
		mod.queue.StopLoop()
		mod.runRule(false)
		<-mod.done
	})
}
