package modules

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
	btls "github.com/bettercap/bettercap/tls"

	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"
)

const (
	httpReadTimeout  = 5 * time.Second
	httpWriteTimeout = 10 * time.Second
)

type HTTPProxy struct {
	Name        string
	Address     string
	Server      *http.Server
	Redirection *firewall.Redirection
	Proxy       *goproxy.ProxyHttpServer
	Script      *HttpProxyScript
	CertFile    string
	KeyFile     string

	isTLS       bool
	isRunning   bool
	stripper    *SSLStripper
	sniListener net.Listener
	sess        *session.Session
}

func stripPort(s string) string {
	ix := strings.IndexRune(s, ':')
	if ix == -1 {
		return s
	}
	return s[:ix]
}

func NewHTTPProxy(s *session.Session) *HTTPProxy {
	p := &HTTPProxy{
		Name:     "http.proxy",
		Proxy:    goproxy.NewProxyHttpServer(),
		sess:     s,
		stripper: NewSSLStripper(s, false),
		isTLS:    false,
		Server:   nil,
	}

	p.Proxy.Verbose = false
	p.Proxy.Logger.SetOutput(ioutil.Discard)

	p.Proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if p.doProxy(req) == true {
			if p.isTLS == false {
				req.URL.Scheme = "http"
			}
			req.URL.Host = req.Host
			p.Proxy.ServeHTTP(w, req)
		}
	})

	p.Proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	p.Proxy.OnRequest().DoFunc(p.onRequestFilter)
	p.Proxy.OnResponse().DoFunc(p.onResponseFilter)

	return p
}

func (p *HTTPProxy) doProxy(req *http.Request) bool {
	blacklist := []string{
		"localhost",
		"127.0.0.1",
	}

	if req.Host == "" {
		log.Error("Got request with empty host: %v", req)
		return false
	}

	for _, blacklisted := range blacklist {
		if strings.HasPrefix(req.Host, blacklisted) {
			log.Error("Got request with blacklisted host: %s", req.Host)
			return false
		}
	}

	return true
}

func (p *HTTPProxy) Configure(address string, proxyPort int, httpPort int, scriptPath string, stripSSL bool) error {
	var err error

	p.stripper.Enable(stripSSL)
	p.Address = address

	if scriptPath != "" {
		if err, p.Script = LoadHttpProxyScript(scriptPath, p.sess); err != nil {
			return err
		} else {
			log.Debug("Proxy script %s loaded.", scriptPath)
		}
	}

	p.Server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", p.Address, proxyPort),
		Handler:      p.Proxy,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}

	if p.sess.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		p.sess.Firewall.EnableForwarding(true)
	}

	p.Redirection = firewall.NewRedirection(p.sess.Interface.Name(),
		"TCP",
		httpPort,
		p.Address,
		proxyPort)

	if err := p.sess.Firewall.EnableRedirection(p.Redirection, true); err != nil {
		return err
	}

	log.Debug("Applied redirection %s", p.Redirection.String())

	p.sess.UnkCmdCallback = func(cmd string) bool {
		if p.Script != nil {
			return p.Script.OnCommand(cmd)
		}
		return false
	}

	return nil
}

func TLSConfigFromCA(ca *tls.Certificate) func(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
	return func(host string, ctx *goproxy.ProxyCtx) (c *tls.Config, err error) {
		parts := strings.SplitN(host, ":", 2)
		hostname := parts[0]
		port := 443
		if len(parts) > 1 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				port = 443
			}
		}

		cert := getCachedCert(hostname, port)
		if cert == nil {
			log.Debug("Creating spoofed certificate for %s:%d", core.Yellow(hostname), port)
			cert, err = btls.SignCertificateForHost(ca, hostname, port)
			if err != nil {
				log.Warning("Cannot sign host certificate with provided CA: %s", err)
				return nil, err
			}

			setCachedCert(hostname, port, cert)
		}

		config := tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{*cert},
		}

		return &config, nil
	}
}

func (p *HTTPProxy) ConfigureTLS(address string, proxyPort int, httpPort int, scriptPath string, certFile string, keyFile string, stripSSL bool) (err error) {
	if p.Configure(address, proxyPort, httpPort, scriptPath, stripSSL); err != nil {
		return err
	}

	p.isTLS = true
	p.Name = "https.proxy"
	p.CertFile = certFile
	p.KeyFile = keyFile

	rawCert, _ := ioutil.ReadFile(p.CertFile)
	rawKey, _ := ioutil.ReadFile(p.KeyFile)
	ourCa, err := tls.X509KeyPair(rawCert, rawKey)
	if err != nil {
		return err
	}

	if ourCa.Leaf, err = x509.ParseCertificate(ourCa.Certificate[0]); err != nil {
		return err
	}

	goproxy.GoproxyCa = ourCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: TLSConfigFromCA(&ourCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: TLSConfigFromCA(&ourCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: TLSConfigFromCA(&ourCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: TLSConfigFromCA(&ourCa)}

	return nil
}

func (p *HTTPProxy) httpWorker() error {
	p.isRunning = true
	return p.Server.ListenAndServe()
}

type dumbResponseWriter struct {
	net.Conn
}

func (dumb dumbResponseWriter) Header() http.Header {
	panic("Header() should not be called on this ResponseWriter")
}

func (dumb dumbResponseWriter) Write(buf []byte) (int, error) {
	if bytes.Equal(buf, []byte("HTTP/1.0 200 OK\r\n\r\n")) {
		return len(buf), nil // throw away the HTTP OK response from the faux CONNECT request
	}
	return dumb.Conn.Write(buf)
}

func (dumb dumbResponseWriter) WriteHeader(code int) {
	panic("WriteHeader() should not be called on this ResponseWriter")
}

func (dumb dumbResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return dumb, bufio.NewReadWriter(bufio.NewReader(dumb), bufio.NewWriter(dumb)), nil
}

func (p *HTTPProxy) httpsWorker() error {
	var err error

	// listen to the TLS ClientHello but make it a CONNECT request instead
	p.sniListener, err = net.Listen("tcp", p.Server.Addr)
	if err != nil {
		return err
	}

	p.isRunning = true
	for p.isRunning {
		c, err := p.sniListener.Accept()
		if err != nil {
			log.Warning("Error accepting connection: %s.", err)
			continue
		}

		go func(c net.Conn) {
			now := time.Now()
			c.SetReadDeadline(now.Add(httpReadTimeout))
			c.SetWriteDeadline(now.Add(httpWriteTimeout))

			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Warning("Error reading SNI: %s.", err)
				return
			}

			hostname := tlsConn.Host()
			if hostname == "" {
				log.Warning("Client does not support SNI.")
				return
			}

			log.Debug("Got new SNI from %s for %s", core.Bold(stripPort(c.RemoteAddr().String())), core.Yellow(hostname))

			req := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: hostname,
					Host:   net.JoinHostPort(hostname, "443"),
				},
				Host:   hostname,
				Header: make(http.Header),
			}
			resp := dumbResponseWriter{tlsConn}
			p.Proxy.ServeHTTP(resp, req)
		}(c)
	}

	return nil
}

func (p *HTTPProxy) Start() {
	go func() {
		var err error

		strip := core.Yellow("enabled")
		if p.stripper.Enabled() == false {
			strip = core.Dim("disabled")
		}

		log.Info("%s started on %s (sslstrip %s)", core.Green(p.Name), p.Server.Addr, strip)

		if p.isTLS == true {
			err = p.httpsWorker()
		} else {
			err = p.httpWorker()
		}

		if err != nil && err.Error() != "http: Server closed" {
			log.Fatal("%s", err)
		}
	}()
}

func (p *HTTPProxy) Stop() error {
	if p.Redirection != nil {
		log.Debug("Disabling redirection %s", p.Redirection.String())
		if err := p.sess.Firewall.EnableRedirection(p.Redirection, false); err != nil {
			return err
		}
		p.Redirection = nil
	}

	p.sess.UnkCmdCallback = nil

	if p.isTLS == true {
		p.isRunning = false
		p.sniListener.Close()
		return nil
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.Server.Shutdown(ctx)
	}
}
