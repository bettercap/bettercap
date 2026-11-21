package firewall

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bettercap/bettercap/v2/core"
	"github.com/bettercap/bettercap/v2/network"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
)

type LinuxFirewall struct {
	iface        *network.Endpoint
	forwarding   bool
	restore      bool
	redirections map[string]*Redirection
}

const (
	IPV4ForwardingFile = "/proc/sys/net/ipv4/ip_forward"
	IPV6ForwardingFile = "/proc/sys/net/ipv6/conf/all/forwarding"
)

func Make(iface *network.Endpoint) FirewallManager {
	firewall := &LinuxFirewall{
		iface:        iface,
		forwarding:   false,
		restore:      false,
		redirections: make(map[string]*Redirection),
	}

	firewall.forwarding = firewall.IsForwardingEnabled()

	return firewall
}

func (f LinuxFirewall) enableFeature(filename string, enable bool) error {
	var value string
	if enable {
		value = "1"
	} else {
		value = "0"
	}

	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.WriteString(value)
	return err
}

func (f LinuxFirewall) IsForwardingEnabled() bool {

	if out, err := ioutil.ReadFile(IPV4ForwardingFile); err != nil {
		return false
	} else {
		return str.Trim(string(out)) == "1"
	}
}

func (f LinuxFirewall) EnableForwarding(enabled bool) error {
	if err := f.enableFeature(IPV4ForwardingFile, enabled); err != nil {
		return err
	}

	if fs.Exists(IPV6ForwardingFile) {
		return f.enableFeature(IPV6ForwardingFile, enabled)
	}

	f.restore = true
	return nil
}

func (f *LinuxFirewall) getCommandLine(r *Redirection, enabled bool) (cmdLine []string) {
	action := "-A"
	destination := ""

	if !enabled {
		action = "-D"
	}

	if r.Drop {
		// insert at the top of FORWARD so the drop wins over any ACCEPT rules
		// an existing firewall manager (ufw, firewalld, libvirt, ...) may have
		// placed earlier in the chain; an appended rule could never be reached.
		dropAction := "-I"
		if !enabled {
			dropAction = "-D"
		}
		cmdLine = []string{
			dropAction, "FORWARD",
			"-i", r.Interface,
			"-p", r.Protocol,
			"--dport", fmt.Sprintf("%d", r.SrcPort),
		}
		if r.DropMatch != "" {
			cmdLine = append(cmdLine, "-m", "string", "--algo", "bm", "--hex-string", r.DropMatch)
		}
		cmdLine = append(cmdLine, "-j", "DROP")
		return
	}

	if strings.Count(r.DstAddress, ":") < 2 {
		destination = r.DstAddress
	} else {
		destination = fmt.Sprintf("[%s]", r.DstAddress)
	}

	if r.SrcAddress == "" {
		cmdLine = []string{
			"-t", "nat",
			action, "PREROUTING",
			"-i", r.Interface,
			"-p", r.Protocol,
			"--dport", fmt.Sprintf("%d", r.SrcPort),
		}
	} else {
		cmdLine = []string{
			"-t", "nat",
			action, "PREROUTING",
			"-i", r.Interface,
			"-p", r.Protocol,
			"-d", r.SrcAddress,
			"--dport", fmt.Sprintf("%d", r.SrcPort),
		}
	}

	// Exclude traffic destined to the proxy machine itself (prevents redirect loops)
	if r.ExcludeAddress != "" {
		cmdLine = append(cmdLine, "!", "-d", r.ExcludeAddress)
	}

	cmdLine = append(cmdLine,
		"-j", "DNAT",
		"--to", fmt.Sprintf("%s:%d", destination, r.DstPort),
	)

	return
}

func (f *LinuxFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	cmdLine := f.getCommandLine(r, enabled)
	rkey := r.String()
	_, found := f.redirections[rkey]

	cmds := []string{}
	if r.Drop {
		cmds = append(cmds, "iptables")
		if fs.Exists(IPV6ForwardingFile) {
			cmds = append(cmds, "ip6tables")
		}
	} else if strings.Count(r.DstAddress, ":") < 2 {
		cmds = append(cmds, "iptables")
	} else {
		cmds = append(cmds, "ip6tables")
	}

	if enabled {
		if found {
			return fmt.Errorf("Redirection '%s' already enabled.", rkey)
		}
		f.redirections[rkey] = r
		for i, cmd := range cmds {
			if !r.Drop {
				if _, err := core.Exec(cmd, []string{"-P", "FORWARD", "ACCEPT"}); err != nil {
					return err
				}
			}
			if _, err := core.Exec(cmd, cmdLine); err != nil {
				if i == 0 {
					delete(f.redirections, rkey)
					return err
				}
				// ip6tables is best-effort
			}
		}
	} else {
		if !found {
			return nil
		}
		delete(f.redirections, rkey)
		for i, cmd := range cmds {
			if _, err := core.Exec(cmd, cmdLine); err != nil {
				if i == 0 && !r.Drop {
					return err
				}
			}
		}
	}

	return nil
}

func (f LinuxFirewall) Restore() {
	if f.restore == false {
		return
	}
	for _, r := range f.redirections {
		if err := f.EnableRedirection(r, false); err != nil {
			fmt.Printf("%s", err)
		}
	}

	if err := f.EnableForwarding(f.forwarding); err != nil {
		fmt.Printf("%s", err)
	}
}
