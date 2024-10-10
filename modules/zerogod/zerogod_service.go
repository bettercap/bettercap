package zerogod

import (
	"fmt"
	"net"
	"strings"

	"github.com/bettercap/bettercap/v2/modules/zerogod/zeroconf"
	"github.com/evilsocket/islazy/tui"
)

type ServiceData struct {
	Name      string            `yaml:"name"`                // Instance name (e.g. "My web page")
	Service   string            `yaml:"service"`             // Service name (e.g. _http._tcp.)
	Domain    string            `yaml:"domain"`              // If blank, assumes "local"
	Port      int               `yaml:"port"`                // Service port
	Records   []string          `yaml:"records,omitempty"`   // Service DNS text records
	Responder string            `yaml:"responder,omitempty"` // Optional IP to use instead of our tcp acceptor
	IPP       map[string]string `yaml:"ipp,omitempty"`       // Optional IPP attributes map
	HTTP      map[string]string `yaml:"http,omitempty"`      // Optional HTTP URIs map

	server *zeroconf.Server
}

func (svc ServiceData) FullName() string {
	return fmt.Sprintf("%s.%s.%s",
		strings.Trim(svc.Name, "."),
		strings.Trim(svc.Service, "."),
		strings.Trim(svc.Domain, "."))
}

func (svc *ServiceData) Register(mod *ZeroGod, localHostName string) (err error) {
	// now create it again to actually advertise
	if svc.Responder == "" {
		// use our own IP
		if svc.server, err = zeroconf.Register(
			svc.Name,
			svc.Service,
			svc.Domain,
			svc.Port,
			svc.Records,
			nil); err != nil {
			return fmt.Errorf("could not create service %s: %v", svc.FullName(), err)
		}

		mod.Info("advertising %s with hostname=%s ipv4=%s ipv6=%s port=%d",
			tui.Yellow(svc.FullName()),
			tui.Red(localHostName),
			tui.Red(mod.Session.Interface.IpAddress),
			tui.Red(mod.Session.Interface.Ip6Address),
			svc.Port)
	} else {
		responderHostName := ""
		// try first to do a reverse DNS of the ip
		if addr, err := net.LookupAddr(svc.Responder); err == nil && len(addr) > 0 {
			responderHostName = addr[0]
		} else {
			mod.Debug("could not get responder %s hostname (%v)", svc.Responder, err)
		}

		// if we don't have a host, create a .nip.io representation
		if responderHostName == "" {
			responderHostName = fmt.Sprintf("%s.nip.io.", strings.ReplaceAll(svc.Responder, ".", "-"))
		}

		// use external responder
		if svc.server, err = zeroconf.RegisterExternalResponder(
			svc.Name,
			svc.Service,
			svc.Domain,
			svc.Port,
			responderHostName,
			[]string{svc.Responder},
			svc.Records,
			nil); err != nil {
			return fmt.Errorf("could not create service %s: %v", svc.FullName(), err)
		}

		mod.Info("advertising %s with responder=%s hostname=%s port=%d",
			tui.Yellow(svc.FullName()),
			tui.Red(svc.Responder),
			tui.Yellow(responderHostName),
			svc.Port)
	}

	return
}

func (svc *ServiceData) Unregister(mod *ZeroGod) error {
	mod.Info("unregistering instance %s ...", tui.Yellow(svc.FullName()))

	err := (error)(nil)
	if svc.server == nil {
		// if we haven't been registered yet, create the server
		if svc.server, err = zeroconf.Register(svc.Name, svc.Service, svc.Domain, svc.Port, svc.Records, nil); err != nil {
			return err
		}
	}

	svc.server.Shutdown()

	return nil
}
