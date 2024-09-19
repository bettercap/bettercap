package zerogod

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	tls_utils "github.com/bettercap/bettercap/v2/tls"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
	"github.com/grandcat/zeroconf"
	yaml "gopkg.in/yaml.v3"
)

type Advertiser struct {
	Filename  string
	Mapping   map[string]zeroconf.ServiceEntry
	Servers   map[string]*zeroconf.Server
	Acceptors map[string]*Acceptor
}

type setupResult struct {
	err    error
	key    string
	server *zeroconf.Server
}

func (mod *ZeroGod) startAdvertiser(fileName string) error {
	if mod.advertiser != nil {
		return fmt.Errorf("advertiser already started for %s", mod.advertiser.Filename)
	}

	var certFile string
	var keyFile string
	var err error

	// read tls configuration
	if err, certFile = mod.StringParam("zerogod.advertise.certificate"); err != nil {
		return err
	} else if certFile, err = fs.Expand(certFile); err != nil {
		return err
	}

	if err, keyFile = mod.StringParam("zerogod.advertise.key"); err != nil {
		return err
	} else if keyFile, err = fs.Expand(keyFile); err != nil {
		return err
	}

	if !fs.Exists(certFile) || !fs.Exists(keyFile) {
		cfg, err := tls_utils.CertConfigFromModule("zerogod.advertise", mod.SessionModule)
		if err != nil {
			return err
		}

		mod.Debug("%+v", cfg)
		mod.Info("generating server TLS key to %s", keyFile)
		mod.Info("generating server TLS certificate to %s", certFile)
		if err := tls_utils.Generate(cfg, certFile, keyFile, false); err != nil {
			return err
		}
	} else {
		mod.Info("loading server TLS key from %s", keyFile)
		mod.Info("loading server TLS certificate from %s", certFile)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	tlsConfig := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", fileName, err)
	}

	mapping := make(map[string]zeroconf.ServiceEntry)
	if err = yaml.Unmarshal(data, &mapping); err != nil {
		return fmt.Errorf("could not deserialize %s: %v", fileName, err)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %v", err)
	}
	if !strings.HasSuffix(hostName, ".") {
		hostName += "."
	}

	mod.Info("loaded %d services from %s, advertising with host=%s iface=%s ipv4=%s ipv6=%s",
		len(mapping),
		fileName,
		hostName,
		mod.Session.Interface.Name(),
		mod.Session.Interface.IpAddress,
		mod.Session.Interface.Ip6Address)

	advertiser := &Advertiser{
		Filename:  fileName,
		Mapping:   mapping,
		Servers:   make(map[string]*zeroconf.Server),
		Acceptors: make(map[string]*Acceptor),
	}

	svcChan := make(chan setupResult)

	// TODO: support external responders

	// paralleize initialization
	for key, svc := range mapping {
		go func(key string, svc zeroconf.ServiceEntry) {
			mod.Info("unregistering instance %s ...", tui.Yellow(fmt.Sprintf("%s.%s.%s", svc.Instance, svc.Service, svc.Domain)))

			// create a first instance just to deregister it from the network
			server, err := zeroconf.Register(svc.Instance, svc.Service, svc.Domain, svc.Port, svc.Text, nil)
			if err != nil {
				svcChan <- setupResult{err: fmt.Errorf("could not create service %s: %v", svc.Instance, err)}
				return
			}
			server.Shutdown()

			// give some time to the network to adjust
			time.Sleep(time.Duration(1) * time.Second)

			// now create it again to actually advertise
			if server, err = zeroconf.Register(svc.Instance, svc.Service, svc.Domain, svc.Port, svc.Text, nil); err != nil {
				svcChan <- setupResult{err: fmt.Errorf("could not create service %s: %v", svc.Instance, err)}
				return
			}

			mod.Info("advertising service %s", tui.Yellow(svc.Service))

			svcChan <- setupResult{
				key:    key,
				server: server,
			}
		}(key, svc)
	}

	for res := range svcChan {
		if res.err != nil {
			return res.err
		}
		advertiser.Servers[res.key] = res.server
		if len(advertiser.Servers) == len(mapping) {
			break
		}
	}

	// now create the tcp acceptors
	for key, svc := range mapping {
		acceptor := NewAcceptor(mod, key, hostName, uint16(svc.Port), &tlsConfig)
		if err := acceptor.Start(); err != nil {
			return err
		}
		advertiser.Acceptors[key] = acceptor
	}

	mod.advertiser = advertiser

	mod.Debug("%+v", *mod.advertiser)

	return nil
}

func (mod *ZeroGod) stopAdvertiser() error {
	if mod.advertiser == nil {
		return errors.New("advertiser not started")
	}

	mod.Info("stopping %d services ...", len(mod.advertiser.Mapping))

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
