package tcp_proxy

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/session"

	"github.com/robertkrimen/otto"
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
	mod := &TcpProxy{
		SessionModule: session.NewSessionModule("tcp.proxy", s),
	}

	mod.AddParam(session.NewIntParameter("tcp.port",
		"443",
		"Remote port to redirect when the TCP proxy is activated."))

	mod.AddParam(session.NewStringParameter("tcp.address",
		"",
		session.IPv4Validator,
		"Remote address of the TCP proxy."))

	mod.AddParam(session.NewStringParameter("tcp.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the TCP proxy to."))

	mod.AddParam(session.NewIntParameter("tcp.proxy.port",
		"8443",
		"Port to bind the TCP proxy to."))

	mod.AddParam(session.NewStringParameter("tcp.proxy.script",
		"",
		"",
		"Path of a TCP proxy JS script."))

	mod.AddParam(session.NewStringParameter("tcp.tunnel.address",
		"",
		"",
		"Address to redirect the TCP tunnel to (optional)."))

	mod.AddParam(session.NewIntParameter("tcp.tunnel.port",
		"0",
		"Port to redirect the TCP tunnel to (optional)."))

	mod.AddHandler(session.NewModuleHandler("tcp.proxy on", "",
		"Start TCP proxy.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("tcp.proxy off", "",
		"Stop TCP proxy.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *TcpProxy) Name() string {
	return "tcp.proxy"
}

func (mod *TcpProxy) Description() string {
	return "A full featured TCP proxy and tunnel, all TCP traffic to a given remote address and port will be redirected to it."
}

func (mod *TcpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *TcpProxy) Configure() error {
	var err error
	var port int
	var proxyPort int
	var address string
	var proxyAddress string
	var scriptPath string
	var tunnelAddress string
	var tunnelPort int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, address = mod.StringParam("tcp.address"); err != nil {
		return err
	} else if err, proxyAddress = mod.StringParam("tcp.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = mod.IntParam("tcp.proxy.port"); err != nil {
		return err
	} else if err, port = mod.IntParam("tcp.port"); err != nil {
		return err
	} else if err, tunnelAddress = mod.StringParam("tcp.tunnel.address"); err != nil {
		return err
	} else if err, tunnelPort = mod.IntParam("tcp.tunnel.port"); err != nil {
		return err
	} else if err, scriptPath = mod.StringParam("tcp.proxy.script"); err != nil {
		return err
	} else if mod.localAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", proxyAddress, proxyPort)); err != nil {
		return err
	} else if mod.remoteAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port)); err != nil {
		return err
	} else if mod.tunnelAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", tunnelAddress, tunnelPort)); err != nil {
		return err
	} else if mod.listener, err = net.ListenTCP("tcp", mod.localAddr); err != nil {
		return err
	}

	if scriptPath != "" {
		if err, mod.script = LoadTcpProxyScript(scriptPath, mod.Session); err != nil {
			return err
		} else {
			mod.Debug("script %s loaded.", scriptPath)
		}
	}

	if !mod.Session.Firewall.IsForwardingEnabled() {
		mod.Info("enabling forwarding.")
		mod.Session.Firewall.EnableForwarding(true)
	}

	mod.Redirection = firewall.NewRedirection(mod.Session.Interface.Name(),
		"TCP",
		port,
		proxyAddress,
		proxyPort)

	mod.Redirection.SrcAddress = address

	if err := mod.Session.Firewall.EnableRedirection(mod.Redirection, true); err != nil {
		return err
	}

	mod.Debug("applied redirection %s", mod.Redirection.String())

	return nil
}

func (mod *TcpProxy) doPipe(from, to net.Addr, src *net.TCPConn, dst io.ReadWriter, wg *sync.WaitGroup) {
	defer wg.Done()

	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			if err.Error() != "EOF" {
				mod.Warning("read failed: %s", err)
			}
			return
		}
		b := buff[:n]

		if mod.script != nil {
			ret := mod.script.OnData(from, to, b, func(call otto.FunctionCall) otto.Value {
				mod.Debug("onData dropCallback called")
				src.Close()
				return otto.Value{}
			})

			if ret != nil {
				nret := len(ret)
				mod.Info("overriding %d bytes of data from %s to %s with %d bytes of new data.",
					n, from.String(), to.String(), nret)
				b = make([]byte, nret)
				copy(b, ret)
			}
		}

		n, err = dst.Write(b)
		if err != nil {
			mod.Warning("write failed: %s", err)
			return
		}

		mod.Debug("%s -> %s : %d bytes", from.String(), to.String(), n)
	}
}

func (mod *TcpProxy) handleConnection(c *net.TCPConn) {
	defer c.Close()

	mod.Info("got a connection from %s", c.RemoteAddr().String())

	// tcp tunnel enabled
	if mod.tunnelAddr.IP.To4() != nil {
		mod.Info("tcp tunnel started ( %s -> %s )", mod.remoteAddr.String(), mod.tunnelAddr.String())
		mod.remoteAddr = mod.tunnelAddr
	}

	remote, err := net.DialTCP("tcp", nil, mod.remoteAddr)
	if err != nil {
		mod.Warning("error while connecting to remote %s: %s", mod.remoteAddr.String(), err)
		return
	}
	defer remote.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// start pipeing
	go mod.doPipe(c.RemoteAddr(), mod.remoteAddr, c, remote, &wg)
	go mod.doPipe(mod.remoteAddr, c.RemoteAddr(), remote, c, &wg)

	wg.Wait()
}

func (mod *TcpProxy) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started ( x -> %s -> %s )", mod.localAddr.String(), mod.remoteAddr.String())

		for mod.Running() {
			conn, err := mod.listener.AcceptTCP()
			if err != nil {
				mod.Warning("error while accepting TCP connection: %s", err)
				continue
			}

			go mod.handleConnection(conn)
		}
	})
}

func (mod *TcpProxy) Stop() error {

	if mod.Redirection != nil {
		mod.Debug("disabling redirection %s", mod.Redirection.String())
		if err := mod.Session.Firewall.EnableRedirection(mod.Redirection, false); err != nil {
			return err
		}
		mod.Redirection = nil
	}

	return mod.SetRunning(false, func() {
		mod.listener.Close()
	})
}
