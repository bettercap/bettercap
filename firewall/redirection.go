package firewall

import "fmt"

type Redirection struct {
	Interface  string
	Protocol   string
	SrcAddress string
	SrcPort    int
	DstAddress string
	DstPort    int
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

func (r Redirection) String() string {
	return fmt.Sprintf("[%s] (%s) %s:%d -> %s:%d", r.Interface, r.Protocol, r.SrcAddress, r.SrcPort, r.DstAddress, r.DstPort)
}
