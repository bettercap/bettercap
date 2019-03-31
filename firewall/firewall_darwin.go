package firewall

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/str"
)

var (
	sysCtlParser = regexp.MustCompile(`([^:]+):\s*(.+)`)
	pfFilePath   = fmt.Sprintf("/tmp/bcap_pf_%d.conf", os.Getpid())
)

type PfFirewall struct {
	iface      *network.Endpoint
	filename   string
	forwarding bool
	enabled    bool
}

func Make(iface *network.Endpoint) FirewallManager {
	firewall := &PfFirewall{
		iface:      iface,
		filename:   pfFilePath,
		forwarding: false,
		enabled:    false,
	}

	firewall.forwarding = firewall.IsForwardingEnabled()

	return firewall
}

func (f PfFirewall) sysCtlRead(param string) (string, error) {
	if out, err := core.Exec("sysctl", []string{param}); err != nil {
		return "", err
	} else if m := sysCtlParser.FindStringSubmatch(out); len(m) == 3 && m[1] == param {
		return m[2], nil
	} else {
		return "", fmt.Errorf("Unexpected sysctl output: %s", out)
	}
}

func (f PfFirewall) sysCtlWrite(param string, value string) (string, error) {
	args := []string{"-w", fmt.Sprintf("%s=%s", param, value)}
	_, err := core.Exec("sysctl", args)
	if err != nil {
		return "", err
	}

	// make sure we actually wrote the value
	if out, err := f.sysCtlRead(param); err != nil {
		return "", err
	} else if out != value {
		return "", fmt.Errorf("Expected value for '%s' is %s, found %s", param, value, out)
	} else {
		return out, nil
	}
}

func (f PfFirewall) IsForwardingEnabled() bool {
	out, err := f.sysCtlRead("net.inet.ip.forwarding")
	if err != nil {
		log.Printf("ERROR: %s", err)
		return false
	}

	return strings.HasSuffix(out, ": 1")
}

func (f PfFirewall) enableParam(param string, enabled bool) error {
	var value string
	if enabled {
		value = "1"
	} else {
		value = "0"
	}

	if _, err := f.sysCtlWrite(param, value); err != nil {
		return err
	} else {
		return nil
	}
}

func (f PfFirewall) EnableForwarding(enabled bool) error {
	return f.enableParam("net.inet.ip.forwarding", enabled)
}

func (f PfFirewall) generateRule(r *Redirection) string {
	src_a := "any"
	dst_a := "any"

	if r.SrcAddress != "" {
		src_a = r.SrcAddress
	}

	if r.DstAddress != "" {
		dst_a = r.DstAddress
	}

	return fmt.Sprintf("rdr pass on %s proto %s from any to %s port %d -> %s port %d",
		r.Interface, r.Protocol, src_a, r.SrcPort, dst_a, r.DstPort)
}

func (f *PfFirewall) enable(enabled bool) {
	f.enabled = enabled
	if enabled {
		core.Exec("pfctl", []string{"-e"})
	} else {
		core.Exec("pfctl", []string{"-d"})
	}
}

func (f PfFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	rule := f.generateRule(r)

	if enabled {
		fd, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer fd.Close()

		if _, err = fd.WriteString(rule + "\n"); err != nil {
			return err
		}

		// enable pf
		f.enable(true)

		// load the rule
		if _, err := core.Exec("pfctl", []string{"-f", f.filename}); err != nil {
			return err
		}
	} else {
		fd, err := os.Open(f.filename)
		if err == nil {
			defer fd.Close()

			lines := ""
			scanner := bufio.NewScanner(fd)
			for scanner.Scan() {
				line := str.Trim(scanner.Text())
				if line != rule {
					lines += line + "\n"
				}
			}

			if str.Trim(lines) == "" {
				os.Remove(f.filename)
				f.enable(false)
			} else {
				ioutil.WriteFile(f.filename, []byte(lines), 0600)
			}
		}
	}

	return nil
}

func (f PfFirewall) Restore() {
	f.EnableForwarding(f.forwarding)
	if f.enabled {
		f.enable(false)
	}
	os.Remove(f.filename)
}
