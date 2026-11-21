package zerogod

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	tls_utils "github.com/bettercap/bettercap/v2/tls"
	yaml "gopkg.in/yaml.v3"
)

type Advertiser struct {
	Filename  string
	Services  []*ServiceData
	Acceptors []*Acceptor
}

func isPortAvailable(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	if conn, err := net.DialTimeout("tcp", address, 10*time.Millisecond); err != nil {
		return true
	} else if conn == nil {
		return true
	} else {
		conn.Close()
		return false
	}
}

func isPortRequested(svc *ServiceData, services []*ServiceData) bool {
	for _, other := range services {
		if svc != other && svc.Port == other.Port {
			return true
		}
	}
	return false
}

func (mod *ZeroGod) loadTLSConfig(host string) (*tls.Config, error) {
	caCfg, err := tls_utils.CAConfigFromModule("zerogod.advertise", mod.SessionModule)
	if err != nil {
		return nil, err
	}

	if caCfg.IsEmudhra() {
		ca, err := caCfg.Resolve()
		if err != nil {
			return nil, err
		}
		mod.Warning("advertising with revoked eMudhra CA %q (chain %s, recovered private key %d)", ca.Name, ca.Chain, ca.KeyValue())
	} else if caCfg.FilesExist() {
		mod.Info("loading server TLS key from %s", caCfg.KeyFile)
		mod.Info("loading server TLS certificate from %s", caCfg.CertFile)
	} else {
		mod.Debug("%+v", caCfg.Config)
		mod.Info("generating server TLS key to %s", caCfg.KeyFile)
		mod.Info("generating server TLS certificate to %s", caCfg.CertFile)
	}

	cert, err := caCfg.LoadForHost(host, caCfg.Config.Bits)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
	}, nil
}

func (mod *ZeroGod) startAdvertiser(fileName string) error {
	if mod.advertiser != nil {
		return fmt.Errorf("advertiser already started for %s", mod.advertiser.Filename)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %v", err)
	}
	hostName = strings.ReplaceAll(hostName, ".local", "")

	// the mDNS A record advertised by zeroconf points at <hostname>.local
	tlsConfig, err := mod.loadTLSConfig(hostName + ".local")
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", fileName, err)
	}

	var services []*ServiceData
	if err = yaml.Unmarshal(data, &services); err != nil {
		return fmt.Errorf("could not deserialize %s: %v", fileName, err)
	}

	numServices := len(services)

	mod.Info("loaded %d services from %s", numServices, fileName)

	advertiser := &Advertiser{
		Filename:  fileName,
		Services:  services,
		Acceptors: make([]*Acceptor, 0),
	}

	// fix ports
	for _, svc := range advertiser.Services {
		// if no external responder has been specified, check if port is available
		if svc.Responder == "" {
			// if the port was not set or is not available or is requesdted by another service
			for svc.Port == 0 || !isPortAvailable(svc.Port) || isPortRequested(svc, services) {
				// set a new one and try again
				newPort := (rand.Intn(65535-1024) + 1024)
				mod.Warning("port %d for service %s is not avaialble, trying %d ...",
					svc.Port,
					svc.FullName(),
					newPort)
				svc.Port = newPort
			}
		}
	}

	// paralleize initialization
	svcChan := make(chan error, numServices)
	for _, svc := range advertiser.Services {
		go func(svc *ServiceData) {
			// deregister the service from the network first
			if err := svc.Unregister(mod); err != nil {
				svcChan <- fmt.Errorf("could not unregister service %s: %v", svc.FullName(), err)
			} else {
				// give some time to the network to adjust
				time.Sleep(time.Duration(1) * time.Second)
				// register it
				if err := svc.Register(mod, hostName); err != nil {
					svcChan <- err
				} else {
					svcChan <- nil
				}
			}
		}(svc)
	}

	for i := 0; i < numServices; i++ {
		if err := <-svcChan; err != nil {
			return err
		}
	}

	// now create the tcp acceptors for entries without an explicit responder address
	for _, svc := range advertiser.Services {
		// if no external responder has been specified
		if svc.Responder == "" {
			acceptor := NewAcceptor(mod, svc.FullName(), hostName, uint16(svc.Port), tlsConfig, svc.IPP, svc.HTTP)
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

	for _, service := range mod.advertiser.Services {
		service.Unregister(mod)
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
