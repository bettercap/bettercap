package firewall

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
)

type LinuxFirewall struct {
	iface        *network.Endpoint
	forwarding   bool
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

	return nil
}

func (f *LinuxFirewall) getCommandLine(r *Redirection, enabled bool) (cmdLine []string) {
	action := "-A"
	destination := ""

	if !enabled {
		action = "-D"
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
			"-j", "DNAT",
			"--to", fmt.Sprintf("%s:%d", destination, r.DstPort),
		}
	} else {
		cmdLine = []string{
			"-t", "nat",
			action, "PREROUTING",
			"-i", r.Interface,
			"-p", r.Protocol,
			"-d", r.SrcAddress,
			"--dport", fmt.Sprintf("%d", r.SrcPort),
			"-j", "DNAT",
			"--to", fmt.Sprintf("%s:%d", destination, r.DstPort),
		}
	}

	return
}

func (f *LinuxFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	cmdLine := f.getCommandLine(r, enabled)
	rkey := r.String()
	_, found := f.redirections[rkey]
	cmd := ""

	if strings.Count(r.DstAddress, ":") < 2 {
		cmd = "iptables"
	} else {
		cmd = "ip6tables"
	}

	if enabled {
		if found {
			return fmt.Errorf("Redirection '%s' already enabled.", rkey)
		}

		f.redirections[rkey] = r

		// accept all
		if _, err := core.Exec(cmd, []string{"-P", "FORWARD", "ACCEPT"}); err != nil {
			return err
		} else if _, err := core.Exec(cmd, cmdLine); err != nil {
			return err
		}
	} else {
		if !found {
			return nil
		}

		delete(f.redirections, r.String())

		if _, err := core.Exec(cmd, cmdLine); err != nil {
			return err
		}
	}

	return nil
}

func (f LinuxFirewall) Restore() {
	for _, r := range f.redirections {
		if err := f.EnableRedirection(r, false); err != nil {
			fmt.Printf("%s", err)
		}
	}

	if err := f.EnableForwarding(f.forwarding); err != nil {
		fmt.Printf("%s", err)
	}
}
