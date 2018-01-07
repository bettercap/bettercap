package session_modules

import (
	"fmt"
	"github.com/op/go-logging"
	"net/http"
	"regexp"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/html"

	"github.com/evilsocket/bettercap-ng/firewall"
	"github.com/evilsocket/bettercap-ng/session"
)

var log = logging.MustGetLogger("mitm")

type ProxyFilter struct {
	Type       string
	Expression string
	Replace    string
	Compiled   *regexp.Regexp
}

func tokenize(s string, sep byte, n int) (error, []string) {
	filtered := make([]string, 0)
	tokens := strings.Split(s, string(sep))

	for _, t := range tokens {
		if t != "" {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) != n {
		return fmt.Errorf("Could not split '%s' by '%s'.", s, string(sep)), filtered
	} else {
		return nil, filtered
	}
}

func NewProxyFilter(type_, expression string) (error, *ProxyFilter) {
	err, tokens := tokenize(expression, expression[0], 2)
	if err != nil {
		return err, nil
	}

	filter := &ProxyFilter{
		Type:       type_,
		Expression: tokens[0],
		Replace:    tokens[1],
		Compiled:   nil,
	}

	if filter.Compiled, err = regexp.Compile(filter.Expression); err != nil {
		return err, nil
	}

	return nil, filter
}

func (f *ProxyFilter) Process(req *http.Request, response_body string) string {
	orig := response_body
	filtered := f.Compiled.ReplaceAllString(orig, f.Replace)

	// TODO: this sucks
	if orig != filtered {
		log.Infof("%s > Applied %s-filtering to %d of response body.", req.RemoteAddr, f.Type, len(filtered))
	}

	return filtered
}

type HttpProxy struct {
	session.SessionModule

	address     string
	redirection *firewall.Redirection
	server      http.Server
	proxy       *goproxy.ProxyHttpServer

	preFilter  *ProxyFilter
	postFilter *ProxyFilter
}

func NewHttpProxy(s *session.Session) *HttpProxy {
	p := &HttpProxy{
		SessionModule: session.NewSessionModule(s),
		proxy:         goproxy.NewProxyHttpServer(),
		address:       "",
		redirection:   nil,
		preFilter:     nil,
		postFilter:    nil,
	}

	p.AddParam(session.NewIntParameter("http.port", "80", "", "HTTP port to redirect when the proxy is activated."))
	p.AddParam(session.NewIntParameter("http.proxy.port", "8080", "", "Port to bind the HTTP proxy to."))
	p.AddParam(session.NewStringParameter("http.proxy.address", "<interface address>", `^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "Address to bind the HTTP proxy to."))
	p.AddParam(session.NewStringParameter("http.proxy.post.filter", "", "", "SED like syntax to replace things in the response ( example |</head>|<script src='...'></script></head>| )."))

	p.AddHandler(session.NewModuleHandler("http.proxy on", "",
		"Start HTTP proxy.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("http.proxy off", "",
		"Stop HTTP proxy.",
		func(args []string) error {
			return p.Stop()
		}))

	p.proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if p.doProxy(req) == true {
			req.URL.Scheme = "http"
			req.URL.Host = req.Host

			// TODO: p.preFilter.Process?

			p.proxy.ServeHTTP(w, req)
		} else {
			log.Infof("Skipping %s\n", req.Host)
		}
	})

	p.proxy.OnResponse(goproxy_html.IsHtml).Do(goproxy_html.HandleString(func(body string, ctx *goproxy.ProxyCtx) string {
		if p.postFilter != nil {
			body = p.postFilter.Process(ctx.Req, body)
		}
		return body
	}))

	return p
}

func (p HttpProxy) Name() string {
	return "HTTP Proxy"
}

func (p HttpProxy) Description() string {
	return "A full featured HTTP proxy that can be used to inject malicious contents into webpages, all HTTP traffic will be redirected to it."
}

func (p HttpProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p HttpProxy) OnSessionStarted(s *session.Session) {
	// refresh the address after session has been created
	s.Env.Set("http.proxy.address", s.Interface.IpAddress)
}

func (p HttpProxy) OnSessionEnded(s *session.Session) {
	if p.Running() {
		p.Stop()
	}
}

func (p *HttpProxy) Start() error {
	var http_port int
	var proxy_port int

	if p.Running() == true {
		return fmt.Errorf("HTTP proxy already started.")
	}

	if err, v := p.Param("http.proxy.address").Get(p.Session); err != nil {
		return err
	} else {
		p.address = v.(string)
	}

	if err, v := p.Param("http.proxy.port").Get(p.Session); err != nil {
		return err
	} else {
		proxy_port = v.(int)
	}

	if err, v := p.Param("http.port").Get(p.Session); err != nil {
		return err
	} else {
		http_port = v.(int)
	}

	p.postFilter = nil
	if err, v := p.Param("http.proxy.post.filter").Get(p.Session); err != nil {
		return err
	} else {
		expression := v.(string)
		if expression != "" {
			if err, p.postFilter = NewProxyFilter("post", expression); err != nil {
				return err
			} else {
				log.Debug("Proxy POST filter set to '%s'.", expression)
			}
		}
	}

	if p.Session.Firewall.IsForwardingEnabled() == false {
		p.Session.Firewall.EnableForwarding(true)
	}

	p.redirection = firewall.NewRedirection(p.Session.Interface.Name(),
		"TCP",
		http_port,
		p.address,
		proxy_port)

	if err := p.Session.Firewall.EnableRedirection(p.redirection, true); err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", p.address, proxy_port)
	log.Infof("Starting proxy on %s.\n", address)

	p.server = http.Server{Addr: address, Handler: p.proxy}
	go func() {
		p.SetRunning(true)
		if err := p.server.ListenAndServe(); err != nil {
			p.SetRunning(false)
			log.Warning(err)
		}
	}()

	return nil
}

func (p *HttpProxy) Stop() error {
	if p.Running() == true {
		p.SetRunning(false)
		p.server.Shutdown(nil)
		log.Info("HTTP proxy stopped.\n")
		if p.redirection != nil {
			if err := p.Session.Firewall.EnableRedirection(p.redirection, false); err != nil {
				return err
			}
			p.redirection = nil
		}
		return nil
	} else {
		return fmt.Errorf("HTTP proxy stopped.")
	}
}

func (p *HttpProxy) doProxy(req *http.Request) bool {
	blacklist := []string{
		"localhost",
		"127.0.0.1",
		p.address,
	}

	if req.Host == "" {
		log.Errorf("Got request with empty host: %v\n", req)
		return false
	}

	for _, blacklisted := range blacklist {
		if strings.HasPrefix(req.Host, blacklisted) {
			log.Errorf("Got request with blacklisted host: %s\n", req.Host)
			return false
		}
	}

	return true
}
