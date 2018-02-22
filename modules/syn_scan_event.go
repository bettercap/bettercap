package modules

import (
	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/session"
)

type SynScanEvent struct {
	Host *network.Endpoint
	Port int
}

func NewSynScanEvent(h *network.Endpoint, port int) SynScanEvent {
	return SynScanEvent{
		Host: h,
		Port: port,
	}
}

func (e SynScanEvent) Push() {
	session.I.Events.Add("syn.scan", e)
	session.I.Refresh()
}
