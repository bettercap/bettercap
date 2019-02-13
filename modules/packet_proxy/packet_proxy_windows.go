package packet_proxy

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

func (mod PacketProxy) Name() string {
	return "packet.proxy"
}

func (mod PacketProxy) Description() string {
	return "Not supported on this OS"
}

func (mod PacketProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *PacketProxy) Configure() (err error) {
	return session.ErrNotSupported
}

func (mod *PacketProxy) Start() error {
	return session.ErrNotSupported
}

func (mod *PacketProxy) Stop() error {
	return session.ErrNotSupported
}
