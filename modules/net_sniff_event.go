package modules

import (
	"fmt"
	"time"

	"github.com/bettercap/bettercap/session"
)

type SniffData map[string]interface{}

type SnifferEvent struct {
	PacketTime  time.Time
	Protocol    string
	Source      string
	Destination string
	Message     string
	Data        interface{}
}

func NewSnifferEvent(t time.Time, proto string, src string, dst string, data interface{}, format string, args ...interface{}) SnifferEvent {
	return SnifferEvent{
		PacketTime:  t,
		Protocol:    proto,
		Source:      src,
		Destination: dst,
		Message:     fmt.Sprintf(format, args...),
		Data:        data,
	}
}

func (e SnifferEvent) Push() {
	session.I.Events.Add("net.sniff.leak."+e.Protocol, e)
	session.I.Refresh()
}
