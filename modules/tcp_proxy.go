package modules

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

type TcpProxy struct {
	session.SessionModule
	Redirection *firewall.Redirection
	localAddr   *net.TCPAddr
	remoteAddr  *net.TCPAddr
	tunnelAddr  *net.TCPAddr
	listener    *net.TCPListener
	script      *TcpProxyScript
}

func NewTcpProxy(s *session.Session) *TcpProxy {
	p := &TcpProxy{
		SessionModule: session.NewSessionModule("tcp.proxy", s),
	}

	p.AddParam(session.NewIntParameter("tcp.port",
		"443",
		"Remote port to redirect when the TCP proxy is activated."))

	p.AddParam(session.NewStringParameter("tcp.address",
		"",
		session.IPv4Validator,
		"Remote address of the TCP proxy."))

	p.AddParam(session.NewStringParameter("tcp.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the TCP proxy to."))

	p.AddParam(session.NewIntParameter("tcp.proxy.port",
		"8443",
		"Port to bind the TCP proxy to."))

	p.AddParam(session.NewStringParameter("tcp.proxy.script",
		"",
		"",
		"Path of a TCP proxy JS script."))

	p.AddParam(session.NewStringParameter("tcp.tunnel.address",
		"",
		"",
		"Address to redirect the TCP tunnel to (optional)."))

	p.AddParam(session.NewIntParameter("tcp.tunnel.port",
		"0",
		"Port to redirect the TCP tunnel to (optional)."))

	p.AddHandler(session.NewModuleHandler("tcp.proxy on", "",
		"Start TCP proxy.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("tcp.proxy off", "",
		"Stop TCP proxy.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p *TcpProxy) Name() string {
	return "tcp.proxy"
}

func (p *TcpProxy) Description() string {
	return "A full featured TCP proxy and tunnel, all TCP traffic to a given remote address and port will be redirected to it."
}

func (p *TcpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *TcpProxy) Configure() error {
	var err error
	var port string
	var proxyPort int
	var address string
	var proxyAddress string
	var scriptPath string
	var tunnelAddress string
	var tunnelPort int

	if p.Running() {
		return session.ErrAlreadyStarted
	} else if err, address = p.StringParam("tcp.address"); err != nil {
		return err
	} else if err, proxyAddress = p.StringParam("tcp.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = p.IntParam("tcp.proxy.port"); err != nil {
		return err
	} else if err, port = p.StringParam("tcp.port"); err != nil {
		return err
	} else if err, tunnelAddress = p.StringParam("tcp.tunnel.address"); err != nil {
		return err
	} else if err, tunnelPort = p.IntParam("tcp.tunnel.port"); err != nil {
		return err
	} else if err, scriptPath = p.StringParam("tcp.proxy.script"); err != nil {
		return err
	} else if p.localAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", proxyAddress, proxyPort)); err != nil {
		return err
	} else if p.remoteAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", address, port)); err != nil {
		return err
	} else if p.tunnelAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", tunnelAddress, tunnelPort)); err != nil {
		return err
	} else if p.listener, err = net.ListenTCP("tcp", p.localAddr); err != nil {
		return err
	}

	if scriptPath != "" {
		if err, p.script = LoadTcpProxyScript(scriptPath, p.Session); err != nil {
			return err
		} else {
			log.Debug("TCP proxy script %s loaded.", scriptPath)
		}
	}

	if !p.Session.Firewall.IsForwardingEnabled() {
		log.Info("Enabling forwarding.")
		p.Session.Firewall.EnableForwarding(true)
	}

	p.Redirection = firewall.NewRedirection(p.Session.Interface.Name(),
		"TCP",
		port,
		proxyAddress,
		proxyPort)

	p.Redirection.SrcAddress = address

	if err := p.Session.Firewall.EnableRedirection(p.Redirection, true); err != nil {
		return err
	}

	log.Debug("Applied redirection %s", p.Redirection.String())

	return nil
}

func (p *TcpProxy) doPipe(from, to net.Addr, src, dst io.ReadWriter, wg *sync.WaitGroup) {
	defer wg.Done()

	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			if err.Error() != "EOF" {
				log.Warning("Read failed: %s", err)
			}
			return
		}
		b := buff[:n]

		if p.script != nil {
			ret := p.script.OnData(from, to, b)

			if ret != nil {
				nret := len(ret)
				log.Info("Overriding %d bytes of data from %s to %s with %d bytes of new data.",
					n, from.String(), to.String(), nret)
				b = make([]byte, nret)
				copy(b, ret)
			}
		}

		n, err = dst.Write(b)
		if err != nil {
			log.Warning("Write failed: %s", err)
			return
		}

		log.Debug("%s -> %s : %d bytes", from.String(), to.String(), n)
	}
}

func (p *TcpProxy) handleConnection(c *net.TCPConn) {
	defer c.Close()

	log.Info("TCP proxy got a connection from %s", c.RemoteAddr().String())

	// tcp tunnel enabled
	if p.tunnelAddr.IP.To4() != nil {
		log.Info("TCP tunnel started ( %s -> %s )", p.remoteAddr.String(), p.tunnelAddr.String())
		p.remoteAddr = p.tunnelAddr
	}

	remote, err := net.DialTCP("tcp", nil, p.remoteAddr)
	if err != nil {
		log.Warning("Error while connecting to remote %s: %s", p.remoteAddr.String(), err)
		return
	}
	defer remote.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// start pipeing
	go p.doPipe(c.RemoteAddr(), p.remoteAddr, c, remote, &wg)
	go p.doPipe(p.remoteAddr, c.RemoteAddr(), remote, c, &wg)

	wg.Wait()
}

func (p *TcpProxy) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		log.Info("TCP proxy started ( x -> %s -> %s )", p.localAddr.String(), p.remoteAddr.String())

		for p.Running() {
			conn, err := p.listener.AcceptTCP()
			if err != nil {
				log.Warning("Error while accepting TCP connection: %s", err)
				continue
			}

			go p.handleConnection(conn)
		}
	})
}

func (p *TcpProxy) Stop() error {
	return p.SetRunning(false, func() {
		p.listener.Close()
	})
}
