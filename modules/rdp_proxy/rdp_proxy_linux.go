// +build !amd64

package rdp_proxy

import (
	"github.com/bettercap/bettercap/session"
)

type RdpProxy struct {
	session.SessionModule
}

func NewRdpProxy(s *session.Session) *RdpProxy {
	return &RdpProxy{
		SessionModule: session.NewSessionModule("rdp.proxy", s),
	}
}

func (mod RdpProxy) Name() string {
	return "rdp.proxy"
}

func (mod RdpProxy) Description() string {
	return "Not supported on this OS"
}

func (mod RdpProxy) Author() string {
	return "Alexandre Beaulieu <alex@segfault.me> && Maxime Carbonneau <pourliver@gmail.com>"
}

func (mod *RdpProxy) Configure() (err error) {
	return session.ErrNotSupported
}

func (mod *RdpProxy) Start() error {
	return session.ErrNotSupported
}

func (mod *RdpProxy) Stop() error {
	return session.ErrNotSupported
}
