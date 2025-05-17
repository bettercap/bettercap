package net_sniff

import (
	"fmt"
	"time"

	"github.com/bettercap/bettercap/v2/session"
)

type SniffData map[string]any

type SnifferEvent struct {
	PacketTime  time.Time `json:"time"`
	Protocol    string    `json:"protocol"`
	Source      string    `json:"from"`
	Destination string    `json:"to"`
	Message     string    `json:"message"`
	Data        any       `json:"data"`
}

func NewSnifferEvent(t time.Time, proto string, src string, dst string, data any, format string, args ...any) SnifferEvent {
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
	session.I.Events.Add("net.sniff."+e.Protocol, e)
	session.I.Refresh()
}
