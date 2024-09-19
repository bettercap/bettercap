package mdns

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/evilsocket/islazy/tui"
	"github.com/grandcat/zeroconf"
	yaml "gopkg.in/yaml.v3"
)

/*
type multiService struct {
	mod      *MDNSModule
	services []*MDNSService
}

func (m multiService) Records(q dns.Question) []dns.RR {
	records := make([]dns.RR, 0)

	m.mod.Debug("QUESTION: %+v", q)

	if strings.HasPrefix(q.Name, "_services._dns-sd._udp.") {
		for _, svc := range m.services {
			records = append(records, svc.Records(q)...)
		}
	} else {
		for _, svc := range m.services {
			if svcRecords := svc.Records(q); len(svcRecords) > 0 {
				records = svcRecords
				break
			}
		}
	}

	if num := len(records); num == 0 {
		m.mod.Debug("unhandled service %+v", q)
	} else {
		m.mod.Info("responding to query %s with %d records", tui.Green(q.Name), num)
		if q.Name == "_services._dns-sd._udp.local." {
			for _, r := range records {
				m.mod.Info("  %+v", r)
			}
		}
	}

	return records
}
*/

type Advertiser struct {
	Filename string
	Mapping  map[string]zeroconf.ServiceEntry
	Servers  map[string]*zeroconf.Server
}

func (mod *MDNSModule) startAdvertiser(fileName string) error {
	if mod.advertiser != nil {
		return fmt.Errorf("advertiser already started for %s", mod.advertiser.Filename)
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

	ifName := mod.Session.Interface.Name()
	/*
		iface, err := net.InterfaceByName(ifName)
		if err != nil {
			return fmt.Errorf("error getting interface %s: %v", ifName, err)
		}
	*/

	mod.Info("loaded %d services from %s, advertising with host=%s iface=%s ipv4=%s ipv6=%s",
		len(mapping),
		fileName,
		hostName,
		ifName,
		mod.Session.Interface.IpAddress,
		mod.Session.Interface.Ip6Address)

	advertiser := &Advertiser{
		Filename: fileName,
		Mapping:  mapping,
		Servers:  make(map[string]*zeroconf.Server),
	}

	for key, svc := range mapping {
		server, err := zeroconf.Register(svc.Instance, svc.Service, svc.Domain, svc.Port, svc.Text, nil)
		if err != nil {
			return fmt.Errorf("could not create service %s: %v", svc.Instance, err)
		}

		mod.Info("advertising service %s", tui.Yellow(svc.Service))

		advertiser.Servers[key] = server
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

	for key, server := range mod.advertiser.Servers {
		mod.Info("stopping %s ...", key)
		server.Shutdown()
	}

	mod.Info("all services stopped")

	mod.advertiser = nil
	return nil
}
