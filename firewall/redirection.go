package firewall

import "fmt"

type Redirection struct {
	Interface      string
	Protocol       string
	SrcAddress     string
	SrcPort        int
	DstAddress     string
	DstPort        int
	ExcludeAddress string // If set, traffic destined to this address is NOT redirected
	Drop           bool
	DropMatch      string
}

func NewRedirection(iface string, proto string, port_from int, addr_to string, port_to int) *Redirection {
	return &Redirection{
		Interface:  iface,
		Protocol:   proto,
		SrcAddress: "",
		SrcPort:    port_from,
		DstAddress: addr_to,
		DstPort:    port_to,
	}
}

func NewDropForwarding(iface string, proto string, port int, match string) *Redirection {
	return &Redirection{
		Interface: iface,
		Protocol:  proto,
		SrcPort:   port,
		Drop:      true,
		DropMatch: match,
	}
}

func (r Redirection) String() string {
	if r.Drop {
		return fmt.Sprintf("[%s] (%s) DROP FORWARD dport %d match=%s", r.Interface, r.Protocol, r.SrcPort, r.DropMatch)
	}
	return fmt.Sprintf("[%s] (%s) %s:%d -> %s:%d", r.Interface, r.Protocol, r.SrcAddress, r.SrcPort, r.DstAddress, r.DstPort)
}
