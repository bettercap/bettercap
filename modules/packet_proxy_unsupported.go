// +build windows darwin

package modules

import (
	"errors"

	"github.com/bettercap/bettercap/session"
)

var (
	notSupported = errors.New("packet.proxy is not supported on this OS")
)

type PacketProxy struct {
	session.SessionModule
}

func NewPacketProxy(s *session.Session) *PacketProxy {
	return &PacketProxy{
		SessionModule: session.NewSessionModule("packet.proxy", s),
	}
}

func (pp PacketProxy) Name() string {
	return "packet.proxy"
}

func (pp PacketProxy) Description() string {
	return "Not supported on this OS"
}

func (pp PacketProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (pp *PacketProxy) Configure() (err error) {
	return notSupported
}

func (pp *PacketProxy) Start() error {
	return notSupported
}

func (pp *PacketProxy) Stop() error {
	return notSupported
}
