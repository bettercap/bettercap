package rdp_proxy

import (
	"github.com/bettercap/bettercap/session"
)

type RdpProxyEvent struct {
	Source      string
	Destination string
	Message     string
}

func NewRdpProxyEvent(src string, dst string, msg string) RdpProxyEvent {
	return RdpProxyEvent{
		Source:      src,
		Destination: dst,
		Message:     msg,
	}
}

func (e RdpProxyEvent) Push() {
	session.I.Events.Add("rdp.proxy", e)
	session.I.Refresh()
}
