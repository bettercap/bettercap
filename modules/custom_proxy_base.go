package modules

import (
	"net/http"
	"github.com/bettercap/bettercap/firewall"
	"net"
	"github.com/bettercap/bettercap/session"
	"strings"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/core"
)

type CustomProxy struct {
	Name        string
	Address     string
	Redirection *firewall.Redirection

	isTLS       bool
	isRunning   bool
	stripper    *SSLStripper
	sniListener net.Listener
	sess        *session.Session
}

func NewCustomProxy(s *session.Session) *CustomProxy {
	p := &CustomProxy{
		Name:     "custom.proxy",
		sess:     s,
		stripper: NewSSLStripper(s, false),
	}
	return p
}

func (p *CustomProxy) doProxy(req *http.Request) bool {
	blacklist := []string{
		"localhost",
		"127.0.0.1",
	}

	if req.Host == "" {
		log.Error("Got request with empty host: %v", req)
		return false
	}

	for _, blacklisted := range blacklist {
		if strings.Split(req.Host, ":")[0] == blacklisted {
			log.Error("Got request with blacklisted host: %s", req.Host)
			return false
		}
	}

	return true
}

func (p *CustomProxy) stripPort(s string) string {
	ix := strings.IndexRune(s, ':')
	if ix == -1 {
		return s
	}
	return s[:ix]
}

func (p *CustomProxy) Configure(proxyAddress string, proxyPort int, srcPort string, stripSSL bool) error {

	p.stripper.Enable(stripSSL)
	p.Address = proxyAddress

	if !p.sess.Firewall.IsForwardingEnabled() {
		log.Info("Enabling forwarding.")
		p.sess.Firewall.EnableForwarding(true)
	}

	p.Redirection = firewall.NewRedirection(p.sess.Interface.Name(),
		"TCP",
		srcPort,
		p.Address,
		proxyPort)

	if err := p.sess.Firewall.EnableRedirection(p.Redirection, true); err != nil {
		return err
	}
	log.Debug("Applied redirection %s", p.Redirection.String())


	return nil
}

func (p *CustomProxy) Start() {
	go func() {
		var err error

		strip := core.Yellow("enabled")
		if !p.stripper.Enabled() {
			strip = core.Dim("disabled")
		}

		log.Info("%s started on %s (sslstrip %s)", core.Green(p.Name), p.Address, strip)

		if err != nil && err.Error() != "http: Server closed" {
			log.Fatal("%s", err)
		}
	}()
}

func (p *CustomProxy) Stop() error {
	if p.Redirection != nil {
		log.Debug("Disabling redirection %s", p.Redirection.String())
		if err := p.sess.Firewall.EnableRedirection(p.Redirection, false); err != nil {
			return err
		}
		p.Redirection = nil
	}
	return nil
}