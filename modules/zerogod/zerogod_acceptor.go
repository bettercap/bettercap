package zerogod

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/evilsocket/islazy/tui"
)

type Handler struct {
	TLS    bool
	Handle func(mod *ZeroGod, client net.Conn, srvHost string, srvPort int, srvTLS bool)
}

// TODO: add more and possibly autodetect from peeking at the first bytes sent by the client
var TCP_HANDLERS = map[string]Handler{
	"_ipp": {
		Handle: ippClientHandler,
	},
	"_ipps": {
		TLS:    true,
		Handle: ippClientHandler,
	},
	// TODO: _http at least
}

type Acceptor struct {
	mod       *ZeroGod
	srvHost   string
	port      uint16
	service   string
	tlsConfig *tls.Config
	listener  net.Listener
	running   bool
	context   context.Context
	ctxCancel context.CancelFunc
	handler   Handler
}

func NewAcceptor(mod *ZeroGod, service string, srvHost string, port uint16, tlsConfig *tls.Config) *Acceptor {
	context, ctcCancel := context.WithCancel(context.Background())
	acceptor := Acceptor{
		mod:       mod,
		port:      port,
		service:   service,
		context:   context,
		ctxCancel: ctcCancel,
		srvHost:   srvHost,
	}

	for svcName, svcHandler := range TCP_HANDLERS {
		if strings.Contains(service, svcName) {
			acceptor.tlsConfig = tlsConfig
			acceptor.handler = svcHandler
			break
		}
	}

	if acceptor.handler.Handle == nil {
		mod.Warning("no protocol handler found for service %s, using generic dump handler", tui.Yellow(service))
		acceptor.handler.Handle = handleGenericTCP
	} else {
		mod.Info("found %s protocol handler", tui.Green(service))
	}

	return &acceptor
}

func (a *Acceptor) Start() (err error) {
	var lc net.ListenConfig

	if a.listener, err = lc.Listen(a.context, "tcp", fmt.Sprintf("0.0.0.0:%d", a.port)); err != nil {
		return err
	}

	if a.tlsConfig != nil {
		a.listener = tls.NewListener(a.listener, a.tlsConfig)
	}

	a.running = true
	go func() {
		a.mod.Debug("tcp listener for port %d (%s) started", a.port, tui.Green(a.service))
		for a.running {
			if conn, err := a.listener.Accept(); err != nil {
				if a.running {
					a.mod.Error("%v", err)
				}
			} else {
				a.mod.Info("accepted connection for service %s (port %d): %v", tui.Green(a.service), a.port, conn.RemoteAddr())
				go a.handler.Handle(a.mod, conn, a.srvHost, int(a.port), a.tlsConfig != nil)
			}
		}
		a.mod.Debug("tcp listener for port %d (%s) stopped", a.port, tui.Green(a.service))
	}()

	return nil
}

func (a *Acceptor) Stop() {
	a.mod.Debug("stopping tcp listener for port %d", a.port)
	a.running = false
	a.ctxCancel()
	<-a.context.Done()
	a.listener.Close()
	a.mod.Debug("tcp listener for port %d stopped", a.port)
}
