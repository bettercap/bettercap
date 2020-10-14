package http_proxy

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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/session"
	btls "github.com/bettercap/bettercap/tls"

	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
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
	Blacklist   []string
	Whitelist   []string
	Sess        *session.Session
	Stripper    *SSLStripper

	jsHook      string
	isTLS       bool
	isRunning   bool
	doRedirect  bool
	sniListener net.Listener
	tag         string
}

func stripPort(s string) string {
	ix := strings.IndexRune(s, ':')
	if ix == -1 {
		return s
	}
	return s[:ix]
}

type dummyLogger struct {
	p *HTTPProxy
}

func (l dummyLogger) Printf(format string, v ...interface{}) {
	l.p.Debug("[goproxy.log] %s", str.Trim(fmt.Sprintf(format, v...)))
}

func NewHTTPProxy(s *session.Session, tag string) *HTTPProxy {
	p := &HTTPProxy{
		Name:       "http.proxy",
		Proxy:      goproxy.NewProxyHttpServer(),
		Sess:       s,
		Stripper:   NewSSLStripper(s, false),
		isTLS:      false,
		doRedirect: true,
		Server:     nil,
		Blacklist:  make([]string, 0),
		Whitelist:  make([]string, 0),
		tag:        session.AsTag(tag),
	}

	p.Proxy.Verbose = false
	p.Proxy.Logger = dummyLogger{p}

	p.Proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if p.doProxy(req) {
			if !p.isTLS {
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

func (p *HTTPProxy) Debug(format string, args ...interface{}) {
	p.Sess.Events.Log(log.DEBUG, p.tag+format, args...)
}

func (p *HTTPProxy) Info(format string, args ...interface{}) {
	p.Sess.Events.Log(log.INFO, p.tag+format, args...)
}

func (p *HTTPProxy) Warning(format string, args ...interface{}) {
	p.Sess.Events.Log(log.WARNING, p.tag+format, args...)
}

func (p *HTTPProxy) Error(format string, args ...interface{}) {
	p.Sess.Events.Log(log.ERROR, p.tag+format, args...)
}

func (p *HTTPProxy) Fatal(format string, args ...interface{}) {
	p.Sess.Events.Log(log.FATAL, p.tag+format, args...)
}

func (p *HTTPProxy) doProxy(req *http.Request) bool {
	if req.Host == "" {
		p.Error("got request with empty host: %v", req)
		return false
	}

	hostname := strings.Split(req.Host, ":")[0]
	for _, local := range []string{"localhost", "127.0.0.1"} {
		if hostname == local {
			p.Error("got request with localed host: %s", req.Host)
			return false
		}
	}

	return true
}

func (p *HTTPProxy) shouldProxy(req *http.Request) bool {
	hostname := strings.Split(req.Host, ":")[0]

	// check for the whitelist
	for _, expr := range p.Whitelist {
		if matched, err := filepath.Match(expr, hostname); err != nil {
			p.Error("error while using proxy whitelist expression '%s': %v", expr, err)
		} else if matched {
			p.Debug("hostname '%s' matched whitelisted element '%s'", hostname, expr)
			return true
		}
	}

	// then the blacklist
	for _, expr := range p.Blacklist {
		if matched, err := filepath.Match(expr, hostname); err != nil {
			p.Error("error while using proxy blacklist expression '%s': %v", expr, err)
		} else if matched {
			p.Debug("hostname '%s' matched blacklisted element '%s'", hostname, expr)
			return false
		}
	}

	return true
}

func (p *HTTPProxy) Configure(address string, proxyPort int, httpPort int, doRedirect bool, scriptPath string,
	jsToInject string, stripSSL bool) error {
	var err error

	// check if another http(s) proxy is using sslstrip and merge strippers
	if stripSSL {
		for _, mname := range []string{"http.proxy", "https.proxy"}{
			err, m := p.Sess.Module(mname)
			if err == nil && m.Running() {
				var mextra interface{}
				var mstripper *SSLStripper
				mextra = m.Extra()
				mextramap := mextra.(map[string]interface{})
				mstripper = mextramap["stripper"].(*SSLStripper)
				if mstripper != nil && mstripper.Enabled() {
					p.Info("found another proxy using sslstrip -> merging strippers...")
					p.Stripper = mstripper
					break
				}
			}
		}
	}

	p.Stripper.Enable(stripSSL)
	p.Address = address
	p.doRedirect = doRedirect
	p.jsHook = ""

	if strings.HasPrefix(jsToInject, "http://") || strings.HasPrefix(jsToInject, "https://") {
		p.jsHook = fmt.Sprintf("<script src=\"%s\" type=\"text/javascript\"></script></head>", jsToInject)
	} else if fs.Exists(jsToInject) {
		if data, err := ioutil.ReadFile(jsToInject); err != nil {
			return err
		} else {
			jsToInject = string(data)
		}
	}

	if p.jsHook == "" && jsToInject != "" {
		if !strings.HasPrefix(jsToInject, "<script ") {
			jsToInject = fmt.Sprintf("<script type=\"text/javascript\">%s</script>", jsToInject)
		}
		p.jsHook = fmt.Sprintf("%s</head>", jsToInject)
	}

	if scriptPath != "" {
		if err, p.Script = LoadHttpProxyScript(scriptPath, p.Sess); err != nil {
			return err
		} else {
			p.Debug("proxy script %s loaded.", scriptPath)
		}
	}

	p.Server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", p.Address, proxyPort),
		Handler:      p.Proxy,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}

	if p.doRedirect {
		if !p.Sess.Firewall.IsForwardingEnabled() {
			p.Info("enabling forwarding.")
			p.Sess.Firewall.EnableForwarding(true)
		}

		p.Redirection = firewall.NewRedirection(p.Sess.Interface.Name(),
			"TCP",
			httpPort,
			p.Address,
			proxyPort)

		if err := p.Sess.Firewall.EnableRedirection(p.Redirection, true); err != nil {
			return err
		}

		p.Debug("applied redirection %s", p.Redirection.String())
	} else {
		p.Warning("port redirection disabled, the proxy must be set manually to work")
	}

	p.Sess.UnkCmdCallback = func(cmd string) bool {
		if p.Script != nil {
			return p.Script.OnCommand(cmd)
		}
		return false
	}

	return nil
}

func (p *HTTPProxy) TLSConfigFromCA(ca *tls.Certificate) func(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
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
			p.Info("creating spoofed certificate for %s:%d", tui.Yellow(hostname), port)
			cert, err = btls.SignCertificateForHost(ca, hostname, port)
			if err != nil {
				p.Warning("cannot sign host certificate with provided CA: %s", err)
				return nil, err
			}

			setCachedCert(hostname, port, cert)
		} else {
			p.Debug("serving spoofed certificate for %s:%d", tui.Yellow(hostname), port)
		}

		config := tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{*cert},
		}

		return &config, nil
	}
}

func (p *HTTPProxy) ConfigureTLS(address string, proxyPort int, httpPort int, doRedirect bool, scriptPath string,
	certFile string,
	keyFile string, jsToInject string, stripSSL bool) (err error) {
	if err = p.Configure(address, proxyPort, httpPort, doRedirect, scriptPath, jsToInject, stripSSL); err != nil {
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
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: p.TLSConfigFromCA(&ourCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: p.TLSConfigFromCA(&ourCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: p.TLSConfigFromCA(&ourCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: p.TLSConfigFromCA(&ourCa)}

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
			p.Warning("error accepting connection: %s.", err)
			continue
		}

		go func(c net.Conn) {
			now := time.Now()
			c.SetReadDeadline(now.Add(httpReadTimeout))
			c.SetWriteDeadline(now.Add(httpWriteTimeout))

			tlsConn, err := vhost.TLS(c)
			if err != nil {
				p.Warning("error reading SNI: %s.", err)
				return
			}

			hostname := tlsConn.Host()
			if hostname == "" {
				p.Warning("client does not support SNI.")
				return
			}

			p.Debug("proxying connection from %s to %s", tui.Bold(stripPort(c.RemoteAddr().String())), tui.Yellow(hostname))

			req := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: hostname,
					Host:   net.JoinHostPort(hostname, "443"),
				},
				Host:       hostname,
				Header:     make(http.Header),
				RemoteAddr: c.RemoteAddr().String(),
			}
			p.Proxy.ServeHTTP(dumbResponseWriter{tlsConn}, req)
		}(c)
	}

	return nil
}

func (p *HTTPProxy) Start() {
	go func() {
		var err error

		strip := tui.Yellow("enabled")
		if !p.Stripper.Enabled() {
			strip = tui.Dim("disabled")
		}

		p.Info("started on %s (sslstrip %s)", p.Server.Addr, strip)

		if p.isTLS {
			err = p.httpsWorker()
		} else {
			err = p.httpWorker()
		}

		if err != nil && err.Error() != "http: Server closed" {
			p.Fatal("%s", err)
		}
	}()
}

func (p *HTTPProxy) Stop() error {
	if p.doRedirect && p.Redirection != nil {
		p.Debug("disabling redirection %s", p.Redirection.String())
		if err := p.Sess.Firewall.EnableRedirection(p.Redirection, false); err != nil {
			return err
		}
		p.Redirection = nil
	}

	p.Sess.UnkCmdCallback = nil

	if p.isTLS {
		p.isRunning = false
		p.sniListener.Close()
		return nil
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.Server.Shutdown(ctx)
	}
}
