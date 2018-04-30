package firewall

import (
	"fmt"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
)

type WindowsFirewall struct {
	iface        *network.Endpoint
	forwarding   bool
	redirections map[string]*Redirection
}

func Make(iface *network.Endpoint) FirewallManager {
	firewall := &WindowsFirewall{
		iface:        iface,
		forwarding:   false,
		redirections: make(map[string]*Redirection, 0),
	}

	firewall.forwarding = firewall.IsForwardingEnabled()

	return firewall
}

func (f WindowsFirewall) IsForwardingEnabled() bool {
	if out, err := core.Exec("netsh", []string{"interface", "ipv4", "dump"}); err != nil {
		fmt.Printf("%s\n", err)
		return false
	} else {
		return strings.Contains(out, "forwarding=enabled")
	}
}

func (f WindowsFirewall) EnableForwarding(enabled bool) error {
	v := "enabled"
	if enabled == false {
		v = "disabled"
	}

	if _, err := core.Exec("netsh", []string{"interface", "ipv4", "set", "interface", fmt.Sprintf("%d", f.iface.Index), fmt.Sprintf("forwarding=\"%s\"", v)}); err != nil {
		return err
	}

	return nil
}

func (f WindowsFirewall) generateRule(r *Redirection, enabled bool) []string {
	// https://stackoverflow.com/questions/24646165/netsh-port-forwarding-from-local-port-to-local-port-not-working
	rule := []string{
		fmt.Sprintf("listenport=%d", r.SrcPort),
	}

	if enabled {
		rule = append(rule, fmt.Sprintf("connectport=%d", r.DstPort))
		rule = append(rule, fmt.Sprintf("connectaddress=%s", r.DstAddress))
		rule = append(rule, fmt.Sprintf("protocol=%s", r.Protocol))
	}

	return rule
}

func (f *WindowsFirewall) AllowPort(port int, address string, proto string, allow bool) error {
	ruleName := fmt.Sprintf("bettercap-rule-%s-%s-%d", address, proto, port)
	nameField := fmt.Sprintf(`name="%s"`, ruleName)
	protoField := fmt.Sprintf("protocol=%s", proto)
	// ipField := fmt.Sprintf("lolcalip=%s", address)
	portField := fmt.Sprintf("localport=%d", port)

	cmd := []string{}

	if allow {
		cmd = []string{"advfirewall", "firewall", "add", "rule", nameField, protoField, "dir=in", portField, "action=allow"}
	} else {
		cmd = []string{"advfirewall", "firewall", "delete", "rule", nameField, protoField, portField}
	}

	if _, err := core.Exec("netsh", cmd); err != nil {
		return err
	}

	return nil
}

func (f *WindowsFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	if err := f.AllowPort(r.SrcPort, r.DstAddress, r.Protocol, enabled); err != nil {
		return err
	} else if err := f.AllowPort(r.DstPort, r.DstAddress, r.Protocol, enabled); err != nil {
		return err
	}

	rule := f.generateRule(r, enabled)
	if enabled {
		rule = append([]string{"interface", "portproxy", "add", "v4tov4"}, rule...)
	} else {
		rule = append([]string{"interface", "portproxy", "delete", "v4tov4"}, rule...)
	}

	if _, err := core.Exec("netsh", rule); err != nil {
		return err
	}

	return nil
}

func (f WindowsFirewall) Restore() {
	for _, r := range f.redirections {
		if err := f.EnableRedirection(r, false); err != nil {
			fmt.Printf("%s", err)
		}
	}

	if err := f.EnableForwarding(f.forwarding); err != nil {
		fmt.Printf("%s", err)
	}
}
