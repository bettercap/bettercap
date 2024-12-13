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

func (f *LinuxFirewall) enableFeature(filename string, enable bool) error {
	value := "0"
	if enable {
		value = "1"
	}

	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer fd.Close()

	if _, err := fd.WriteString(value); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	return nil
}

func (f *LinuxFirewall) IsForwardingEnabled() bool {
	content, err := ioutil.ReadFile(IPV4ForwardingFile)
	if err != nil {
		return false
	}
	return str.Trim(string(content)) == "1"
}

func (f *LinuxFirewall) EnableForwarding(enabled bool) error {
	if err := f.enableFeature(IPV4ForwardingFile, enabled); err != nil {
		return err
	}

	if fs.Exists(IPV6ForwardingFile) {
		if err := f.enableFeature(IPV6ForwardingFile, enabled); err != nil {
			return err
		}
	}

	f.restore = true
	return nil
