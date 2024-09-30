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
	"github.com/evilsocket/islazy/fs"
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
	hostName = strings.ReplaceAll(hostName, ".local", "")

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
