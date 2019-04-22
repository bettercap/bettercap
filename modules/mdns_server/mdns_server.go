package mdns_server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"

	"github.com/hashicorp/mdns"
)

type MDNSServer struct {
	session.SessionModule
	hostname string
	instance string
	service  *mdns.MDNSService
	server   *mdns.Server
}

func NewMDNSServer(s *session.Session) *MDNSServer {
	host, _ := os.Hostname()
	mod := &MDNSServer{
		SessionModule: session.NewSessionModule("mdns.server", s),
		hostname:      host,
	}

	mod.AddParam(session.NewStringParameter("mdns.server.host",
		mod.hostname+".",
		"",
		"mDNS hostname to advertise on the network."))

	mod.AddParam(session.NewStringParameter("mdns.server.service",
		"_companion-link._tcp.",
		"",
		"mDNS service name to advertise on the network."))

	mod.AddParam(session.NewStringParameter("mdns.server.domain",
		"local.",
		"",
		"mDNS domain."))

	mod.AddParam(session.NewStringParameter("mdns.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"IPv4 address of the mDNS service."))

	mod.AddParam(session.NewStringParameter("mdns.server.address6",
		session.ParamIfaceAddress6,
		session.IPv6Validator,
		"IPv6 address of the mDNS service."))

	mod.AddParam(session.NewIntParameter("mdns.server.port",
		"52377",
		"Port of the mDNS service."))

	mod.AddParam(session.NewStringParameter("mdns.server.info",
		"rpBA=DE:AD:BE:EF:CA:FE, rpAD=abf99d4ff73f, rpHI=ec5fb3caf528, rpHN=20f8fb46e2eb, rpVr=164.16, rpHA=7406bd0eff69",
		"",
		"Comma separated list of informative TXT records for the mDNS server."))

	mod.AddHandler(session.NewModuleHandler("mdns.server on", "",
		"Start mDNS server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("mdns.server off", "",
		"Stop mDNS server.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *MDNSServer) Name() string {
	return "mdns.server"
}

func (mod *MDNSServer) Description() string {
	return "A mDNS server module to create multicast services or spoof existing ones."
}

func (mod *MDNSServer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *MDNSServer) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	var host string
	var service string
	var domain string
	var ip4 string
	var ip6 string
	var port int
	var info string

	if err, host = mod.StringParam("mdns.server.host"); err != nil {
		return err
	} else if err, service = mod.StringParam("mdns.server.service"); err != nil {
		return err
	} else if err, domain = mod.StringParam("mdns.server.domain"); err != nil {
		return err
	} else if err, ip4 = mod.StringParam("mdns.server.address"); err != nil {
		return err
	} else if err, ip6 = mod.StringParam("mdns.server.address6"); err != nil {
		return err
	} else if err, port = mod.IntParam("mdns.server.port"); err != nil {
		return err
	} else if err, info = mod.StringParam("mdns.server.info"); err != nil {
		return err
	}

	log.SetOutput(ioutil.Discard)

	mod.instance = fmt.Sprintf("%s%s%s", host, service, domain)
	mod.service, err = mdns.NewMDNSService(
		mod.instance,
		service,
		domain,
		host,
		port,
		[]net.IP{
			net.ParseIP(ip4),
			net.ParseIP(ip6),
		},
		str.Comma(info))

	return err
}

func (mod *MDNSServer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		var err error
		mod.Info("advertising service %s -> %s:%d", tui.Bold(mod.instance), mod.service.IPs, mod.service.Port)
		if mod.server, err = mdns.NewServer(&mdns.Config{Zone: mod.service}); err != nil {
			mod.Error("%v", err)
			mod.Stop()
		}
	})
}

func (mod *MDNSServer) Stop() error {
	return mod.SetRunning(false, func() {
		mod.server.Shutdown()
	})
}
