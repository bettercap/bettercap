package mdns

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	yaml "gopkg.in/yaml.v3"
)

type multiService struct {
	services []*MDNSService
}

func (m multiService) Records(q dns.Question) []dns.RR {
	records := make([]dns.RR, 0)

	for _, svc := range m.services {
		records = append(records, svc.Records(q)...)
	}

	return records
}

type Advertiser struct {
	Filename string
	Mapping  map[string]ServiceEntry

	Service multiService
	Server  *Server
}

func (mod *MDNSModule) startAdvertiser(fileName string) error {
	if mod.advertiser != nil {
		return fmt.Errorf("advertiser already started for %s", mod.advertiser.Filename)
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", fileName, err)
	}

	mapping := make(map[string]ServiceEntry)
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

	mod.Info("loaded %d services from %s, advertising with: host=%s ipv4=%s ipv6=%s",
		len(mapping),
		fileName,
		hostName,
		mod.Session.Interface.IpAddress,
		mod.Session.Interface.Ip6Address)

	advertiser := &Advertiser{
		Filename: fileName,
		Mapping:  mapping,
		Service: multiService{
			services: make([]*MDNSService, 0),
		},
	}

	for _, svcData := range mapping {
		svcParts := strings.SplitN(svcData.Name, ".", 2)
		svcInstance := svcParts[0]
		svcService := strings.Replace(svcParts[1], ".local.", "", 1)

		// TODO: patch UUID

		service, err := NewMDNSService(
			mod,
			svcInstance,
			svcService,
			"local.",
			hostName,
			svcData.Port,
			[]net.IP{
				mod.Session.Interface.IP,
				mod.Session.Interface.IPv6,
			},
			svcData.InfoFields)
		if err != nil {
			return fmt.Errorf("could not create service %s: %v", svcData.Name, err)
		}

		advertiser.Service.services = append(advertiser.Service.services, service)
	}

	if advertiser.Server, err = NewServer(mod, &Config{Zone: advertiser.Service}); err != nil {
		return fmt.Errorf("could not create server: %v", err)
	}

	mod.advertiser = advertiser

	mod.Debug("%+v", *mod.advertiser)

	return nil
}

func (mod *MDNSModule) stopAdvertiser() error {
	if mod.advertiser == nil {
		return errors.New("advertiser not started")
	}

	mod.Info("stopping %d services ...", len(mod.advertiser.Mapping))

	mod.advertiser.Server.Shutdown()

	mod.advertiser = nil
	return nil
}
