package modules

import (
	"github.com/bettercap/bettercap/session"
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
	return session.ErrNotSupported
}

func (pp *PacketProxy) Start() error {
	return session.ErrNotSupported
}

func (pp *PacketProxy) Stop() error {
	return session.ErrNotSupported
}
