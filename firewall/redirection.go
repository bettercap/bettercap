package firewall

import (
	"fmt"
	"strconv"
)

type Redirection struct {
	Interface  string
	Protocol   string
	SrcAddress string
	SrcPort    string
	DstAddress string
	DstPort    int
	MultiPort	bool
}

func NewRedirection(iface string, proto string, port_from string, addr_to string, port_to int) *Redirection {
	_, err := strconv.Atoi(port_from)
	multi_port := false
	if err != nil {
		multi_port = true
	} else {
		multi_port = false
	}
	return &Redirection{
		Interface:  iface,
		Protocol:   proto,
		SrcAddress: "",
		SrcPort:    port_from,
		DstAddress: addr_to,
		DstPort:    port_to,
		MultiPort:  multi_port,
	}
}

func (r Redirection) String() string {
	return fmt.Sprintf("[%s] (%s) %s:%d -> %s:%d", r.Interface, r.Protocol, r.SrcAddress, r.SrcPort, r.DstAddress, r.DstPort)
}
