package dns_proxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/firewall"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/log"

	"github.com/miekg/dns"

	"github.com/robertkrimen/otto"
)

const (
	dialTimeout  = 2 * time.Second
	readTimeout  = 2 * time.Second
	writeTimeout = 2 * time.Second
)

type DNSProxy struct {
	Name        string
	Address     string
	Server      *dns.Server
	Redirection *firewall.Redirection
	Nameserver  string
	NetProtocol string
	Script      *DnsProxyScript
	CertFile    string
	KeyFile     string
	Blacklist   []string
	Whitelist   []string
	Sess        *session.Session

	doRedirect bool
	isRunning  bool
	tag        string
}

func (p *DNSProxy) shouldProxy(clientIP string) bool {
	// check if this client is in the whitelist
	for _, ip := range p.Whitelist {
		if clientIP == ip {
			return true
		}
	}

	// check if this client is in the blacklist
	for _, ip := range p.Blacklist {
		if ip == "*" || clientIP == ip {
			return false
		}
	}

	return true
}

func (p *DNSProxy) Configure(address string, dnsPort int, doRedirect bool, nameserver string, netProtocol string, proxyPort int, scriptPath string, certFile string, keyFile string) error {
	var err error

	p.Address = address
	p.doRedirect = doRedirect
	p.CertFile = certFile
	p.KeyFile = keyFile

	if scriptPath != "" {
		if err, p.Script = LoadDnsProxyScript(scriptPath, p.Sess); err != nil {
			return err
		} else {
			p.Debug("proxy script %s loaded.", scriptPath)
		}
	}

	dnsClient := dns.Client{
		DialTimeout:  dialTimeout,
		Net:          netProtocol,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	resolverAddr := fmt.Sprintf("%s:%d", nameserver, dnsPort)

	handler := dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)

		clientIP := strings.Split(w.RemoteAddr().String(), ":")[0]

		req, res := p.onRequestFilter(req, clientIP)
		if res == nil {
			// unused var is time til res
			res, _, err := dnsClient.Exchange(req, resolverAddr)
			if err != nil {
				p.Debug("error while resolving DNS query: %s", err.Error())
				m.SetRcode(req, dns.RcodeServerFailure)
				w.WriteMsg(m)
				return
			}
			res = p.onResponseFilter(req, res, clientIP)
			if res == nil {
				p.Debug("response is nil")
				m.SetRcode(req, dns.RcodeServerFailure)
				w.WriteMsg(m)
				return
			} else {
				if err := w.WriteMsg(res); err != nil {
					p.Error("Error writing response: %s", err)
				}
			}
		} else {
			if err := w.WriteMsg(res); err != nil {
				p.Error("Error writing response: %s", err)
			}
		}
	})

	p.Server = &dns.Server{
		Addr:    fmt.Sprintf("%s:%d", address, proxyPort),
		Net:     netProtocol,
		Handler: handler,
	}

	if netProtocol == "tcp-tls" && p.CertFile != "" && p.KeyFile != "" {
		rawCert, _ := ioutil.ReadFile(p.CertFile)
		rawKey, _ := ioutil.ReadFile(p.KeyFile)
		ourCa, err := tls.X509KeyPair(rawCert, rawKey)
		if err != nil {
			return err
		}

		if ourCa.Leaf, err = x509.ParseCertificate(ourCa.Certificate[0]); err != nil {
			return err
		}

		p.Server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{ourCa},
		}
	}

	if p.doRedirect {
		if !p.Sess.Firewall.IsForwardingEnabled() {
			p.Info("enabling forwarding.")
			p.Sess.Firewall.EnableForwarding(true)
		}

		redirectProtocol := netProtocol
		if redirectProtocol == "tcp-tls" {
			redirectProtocol = "tcp"
		}
		p.Redirection = firewall.NewRedirection(p.Sess.Interface.Name(),
			redirectProtocol,
			dnsPort,
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

func (p *DNSProxy) dnsWorker() error {
	p.isRunning = true
	return p.Server.ListenAndServe()
}

func (p *DNSProxy) Debug(format string, args ...interface{}) {
	p.Sess.Events.Log(log.DEBUG, p.tag+format, args...)
}

func (p *DNSProxy) Info(format string, args ...interface{}) {
	p.Sess.Events.Log(log.INFO, p.tag+format, args...)
}

func (p *DNSProxy) Warning(format string, args ...interface{}) {
	p.Sess.Events.Log(log.WARNING, p.tag+format, args...)
}

func (p *DNSProxy) Error(format string, args ...interface{}) {
	p.Sess.Events.Log(log.ERROR, p.tag+format, args...)
}

func (p *DNSProxy) Fatal(format string, args ...interface{}) {
	p.Sess.Events.Log(log.FATAL, p.tag+format, args...)
}

func NewDNSProxy(s *session.Session, tag string) *DNSProxy {
	p := &DNSProxy{
		Name:       "dns.proxy",
		Sess:       s,
		Server:     nil,
		doRedirect: true,
		tag:        session.AsTag(tag),
	}

	return p
}

func (p *DNSProxy) Start() {
	go func() {
		p.Info("started on %s", p.Server.Addr)

		err := p.dnsWorker()
		// TODO: check the dns server closed error
		if err != nil && err.Error() != "dns: Server closed" {
			p.Fatal("%s", err)
		}
	}()
}

func (p *DNSProxy) Stop() error {
	if p.Script != nil {
		if p.Script.Plugin.HasFunc("onExit") {
			if _, err := p.Script.Call("onExit"); err != nil {
				log.Error("Error while executing onExit callback: %s", "\nTraceback:\n  "+err.(*otto.Error).String())
			}
		}
	}

	if p.doRedirect && p.Redirection != nil {
		p.Debug("disabling redirection %s", p.Redirection.String())
		if err := p.Sess.Firewall.EnableRedirection(p.Redirection, false); err != nil {
			return err
		}
		p.Redirection = nil
	}

	p.Sess.UnkCmdCallback = nil

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.Server.ShutdownContext(ctx)
}
