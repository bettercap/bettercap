package zerogod

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	tls_utils "github.com/bettercap/bettercap/v2/tls"
	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
	yaml "gopkg.in/yaml.v3"
)

type Advertiser struct {
	Filename string

	Services  []ServiceData
	Servers   []*zeroconf.Server
	Acceptors []*Acceptor
}

type setupResult struct {
	err    error
	server *zeroconf.Server
}

func (mod *ZeroGod) loadTLSConfig() (*tls.Config, error) {
	var certFile string
	var keyFile string
	var err error

	// read tls configuration
	if err, certFile = mod.StringParam("zerogod.advertise.certificate"); err != nil {
		return nil, err
	} else if certFile, err = fs.Expand(certFile); err != nil {
		return nil, err
	}

	if err, keyFile = mod.StringParam("zerogod.advertise.key"); err != nil {
		return nil, err
	} else if keyFile, err = fs.Expand(keyFile); err != nil {
		return nil, err
	}

	if !fs.Exists(certFile) || !fs.Exists(keyFile) {
		cfg, err := tls_utils.CertConfigFromModule("zerogod.advertise", mod.SessionModule)
		if err != nil {
			return nil, err
		}

		mod.Debug("%+v", cfg)
		mod.Info("generating server TLS key to %s", keyFile)
		mod.Info("generating server TLS certificate to %s", certFile)
		if err := tls_utils.Generate(cfg, certFile, keyFile, false); err != nil {
			return nil, err
		}
	} else {
		mod.Info("loading server TLS key from %s", keyFile)
		mod.Info("loading server TLS certificate from %s", certFile)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}, nil
}

func (mod *ZeroGod) startAdvertiser(fileName string) error {
	if mod.advertiser != nil {
		return fmt.Errorf("advertiser already started for %s", mod.advertiser.Filename)
	}

	tlsConfig, err := mod.loadTLSConfig()
	if err != nil {
		return err
	}

	hostName, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %v", err)
	}
	if !strings.HasSuffix(hostName, ".") {
		hostName += "."
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", fileName, err)
	}

	var services []ServiceData
	if err = yaml.Unmarshal(data, &services); err != nil {
		return fmt.Errorf("could not deserialize %s: %v", fileName, err)
	}

	mod.Info("loaded %d services from %s", len(services), fileName)

	advertiser := &Advertiser{
		Filename:  fileName,
		Services:  services,
		Servers:   make([]*zeroconf.Server, 0),
		Acceptors: make([]*Acceptor, 0),
	}

	svcChan := make(chan setupResult)

	// paralleize initialization
	for _, svc := range services {
		go func(svc ServiceData) {
			mod.Info("unregistering instance %s ...", tui.Yellow(svc.FullName()))

			// deregister the service from the network first
			if err := svc.Unregister(); err != nil {
				svcChan <- setupResult{err: fmt.Errorf("could not unregister service %s: %v", svc.FullName(), err)}
				return
			}

			// give some time to the network to adjust
			time.Sleep(time.Duration(1) * time.Second)

			var server *zeroconf.Server

			// now create it again to actually advertise
			if svc.Responder == "" {
				// use our own IP
				if server, err = zeroconf.Register(
					svc.Name,
					svc.Service,
					svc.Domain,
					svc.Port,
					svc.Records,
					nil); err != nil {
					svcChan <- setupResult{err: fmt.Errorf("could not create service %s: %v", svc.FullName(), err)}
					return
				}
				mod.Info("advertising %s with responder=%s port=%d",
					tui.Yellow(svc.FullName()),
					tui.Red(svc.Responder),
					svc.Port)
			} else {
				responderHostName := ""

				// try first to do a reverse DNS of the ip
				if addr, err := net.LookupAddr(svc.Responder); err == nil && len(addr) > 0 {
					responderHostName = addr[0]
				} else {
					mod.Debug("could not get responder %s reverse dns entry: %v", svc.Responder, err)
				}

				// if we don't have a host, create a .nip.io representation
				if responderHostName == "" {
					responderHostName = fmt.Sprintf("%s.nip.io.", strings.ReplaceAll(svc.Responder, ".", "-"))
				}

				// use external responder
				if server, err = zeroconf.RegisterExternalResponder(
					svc.Name,
					svc.Service,
					svc.Domain,
					svc.Port,
					responderHostName,
					[]string{svc.Responder},
					svc.Records,
					nil); err != nil {
					svcChan <- setupResult{err: fmt.Errorf("could not create service %s: %v", svc.FullName(), err)}
					return
				}

				mod.Info("advertising %s with responder=%s hostname=%s port=%d",
					tui.Yellow(svc.FullName()),
					tui.Red(svc.Responder),
					tui.Yellow(responderHostName),
					svc.Port)
			}

			svcChan <- setupResult{
				server: server,
			}
		}(svc)
	}

	for res := range svcChan {
		if res.err != nil {
			return res.err
		}
		advertiser.Servers = append(advertiser.Servers, res.server)
		if len(advertiser.Servers) == len(advertiser.Services) {
			break
		}
	}

	// now create the tcp acceptors for entries without an explicit responder address
	for _, svc := range advertiser.Services {
		if svc.Responder == "" {
			acceptor := NewAcceptor(mod, svc.FullName(), hostName, uint16(svc.Port), tlsConfig)
			if err := acceptor.Start(); err != nil {
				return err
			}
			advertiser.Acceptors = append(advertiser.Acceptors, acceptor)
		}
	}

	mod.advertiser = advertiser

	mod.Debug("%+v", *mod.advertiser)

	return nil
}

func (mod *ZeroGod) stopAdvertiser() error {
	if mod.advertiser == nil {
		return errors.New("advertiser not started")
	}

	mod.Info("stopping %d services ...", len(mod.advertiser.Services))

	for key, server := range mod.advertiser.Servers {
		mod.Info("stopping %s ...", key)
		server.Shutdown()
	}

	mod.Info("all services stopped")

	mod.Info("stopping %d acceptors ...", len(mod.advertiser.Acceptors))

	for _, acceptor := range mod.advertiser.Acceptors {
		acceptor.Stop()
	}

	mod.Info("all acceptors stopped")

	mod.advertiser = nil
	return nil
}
