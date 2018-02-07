package firewall

import (
	"fmt"
)

type WindowsFirewall struct {
	forwarding   bool
	redirections map[string]*Redirection
}

func Make() FirewallManager {
	firewall := &WindowsFirewall{
		forwarding:   false,
		redirections: make(map[string]*Redirection, 0),
	}

	firewall.forwarding = firewall.IsForwardingEnabled()

	return firewall
}

func (f WindowsFirewall) IsForwardingEnabled() bool {
	return false
}

func (f WindowsFirewall) EnableForwarding(enabled bool) error {
	return fmt.Errorf("Not implemented")
}

func (f *WindowsFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	return fmt.Errorf("Not implemented")
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
