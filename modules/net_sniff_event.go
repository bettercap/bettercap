package modules

import (
	"fmt"
	"time"

	"github.com/evilsocket/bettercap-ng/session"
)

type SniffData map[string]interface{}

type SnifferEvent struct {
	PacketTime  time.Time
	Protocol    string
	Source      string
	Destination string
	Data        SniffData
	Message     string
}

func NewSnifferEvent(t time.Time, proto string, src string, dst string, data SniffData, format string, args ...interface{}) SnifferEvent {
	return SnifferEvent{
		PacketTime:  t,
		Protocol:    proto,
		Source:      src,
		Destination: dst,
		Data:        data,
		Message:     fmt.Sprintf(format, args...),
	}
}

func (e SnifferEvent) Push() {
	session.I.Events.Add("net.sniff.leak."+e.Protocol, e)
	session.I.Refresh()
}
