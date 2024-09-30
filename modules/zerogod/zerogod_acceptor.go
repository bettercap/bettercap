package zerogod

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
)

type Handler struct {
	TLS    bool
	Handle func(ctx *HandlerContext)
}

// TODO: possibly autodetect from peeking at the first bytes sent by the client
var TCP_HANDLERS = map[string]Handler{
	"_ipp": {
		Handle: ippClientHandler,
	},
	"_ipps": {
		TLS:    true,
		Handle: ippClientHandler,
	},
	"_http": {
		Handle: httpClientHandler,
	},
	"_https": {
		TLS:    true,
		Handle: httpClientHandler,
	},
}

type Acceptor struct {
	mod           *ZeroGod
	srvHost       string
	proto         string
	port          uint16
	service       string
	tlsConfig     *tls.Config
	tcpListener   net.Listener
	udpListener   *net.UDPConn
	running       bool
	context       context.Context
	ctxCancel     context.CancelFunc
	handler       Handler
	ippAttributes map[string]string
	httpPaths     map[string]string
}

type HandlerContext struct {
	service       string
	mod           *ZeroGod
	client        net.Conn
	srvHost       string
	srvPort       int
	srvTLS        bool
	ippAttributes map[string]string
	httpPaths     map[string]string
}

func NewAcceptor(mod *ZeroGod, service string, srvHost string, port uint16, tlsConfig *tls.Config, ippAttributes map[string]string, httpPaths map[string]string) *Acceptor {
	context, ctcCancel := context.WithCancel(context.Background())
	proto := ops.Ternary(strings.Contains(service, "_tcp"), "tcp", "udp").(string)

	acceptor := Acceptor{
		mod:           mod,
		port:          port,
		proto:         proto,
		service:       service,
		context:       context,
		ctxCancel:     ctcCancel,
		srvHost:       srvHost,
		ippAttributes: ippAttributes,
		httpPaths:     httpPaths,
	}

	for svcName, svcHandler := range TCP_HANDLERS {
		if strings.Contains(service, svcName) {
			if svcHandler.TLS {
				acceptor.tlsConfig = tlsConfig
			}
			acceptor.handler = svcHandler
			break
		}
	}

	if acceptor.handler.Handle == nil {
		mod.Warning("no protocol handler found for service %s, using generic %s dump handler", tui.Yellow(service), proto)
		acceptor.handler.Handle = handleGenericTCP
	} else {
		mod.Info("found %s %s protocol handler (tls=%v)", proto, tui.Green(service), acceptor.tlsConfig != nil)
	}

	return &acceptor
}

func (a *Acceptor) startTCP() (err error) {
	var lc net.ListenConfig
	if a.tcpListener, err = lc.Listen(a.context, "tcp", fmt.Sprintf("0.0.0.0:%d", a.port)); err != nil {
		return err
	}
	if a.tlsConfig != nil {
		a.tcpListener = tls.NewListener(a.tcpListener, a.tlsConfig)
	}

	a.running = true
	go func() {
		a.mod.Debug("%s listener for port %d (%s) started", a.proto, a.port, tui.Green(a.service))
		for a.running {
			if conn, err := a.tcpListener.Accept(); err != nil {
				if a.running {
					a.mod.Error("%v", err)
				}
			} else {
				a.mod.Debug("accepted %s connection for service %s (port %d): %v", a.proto, tui.Green(a.service), a.port, conn.RemoteAddr())
				go a.handler.Handle(&HandlerContext{
					service:       a.service,
					mod:           a.mod,
					client:        conn,
					srvHost:       a.srvHost,
					srvPort:       int(a.port),
					srvTLS:        a.tlsConfig != nil,
					ippAttributes: a.ippAttributes,
					httpPaths:     a.httpPaths,
				})
			}
		}
		a.mod.Debug("%s listener for port %d (%s) stopped", a.proto, a.port, tui.Green(a.service))
	}()

	return nil
}

func (a *Acceptor) startUDP() (err error) {
	if udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", a.port)); err != nil {
		return err
	} else if a.udpListener, err = net.ListenUDP("udp", udpAddr); err != nil {
		return err
	} else {
		a.running = true
		go func() {
			var buffer [4096]byte

			a.mod.Info("%s listener for port %d (%s) started", a.proto, a.port, tui.Green(a.service))

			for a.running {
				if n, addr, err := a.udpListener.ReadFromUDP(buffer[0:]); err != nil {
					a.mod.Warning("error reading udp packet: %v", err)
				} else if n <= 0 {
					a.mod.Info("empty read")
				} else {
					a.mod.Info("%v:\n%s", addr, Dump(buffer[0:n]))
				}
			}

			a.mod.Info("%s listener for port %d (%s) stopped", a.proto, a.port, tui.Green(a.service))
		}()
	}

	return nil
}

func (a *Acceptor) Start() (err error) {
	if a.proto == "tcp" {
		return a.startTCP()
	} else {
		return a.startUDP()
	}
}

func (a *Acceptor) Stop() {
	a.mod.Debug("stopping %s listener for port %d", a.proto, a.port)
	a.running = false

	if a.proto == "tcp" {
		a.ctxCancel()
		<-a.context.Done()
		a.tcpListener.Close()
	} else {
		a.udpListener.Close()
	}

	a.mod.Debug("%s listener for port %d stopped", a.proto, a.port)
}
