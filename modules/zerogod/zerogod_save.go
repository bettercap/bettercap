package zerogod

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/str"
	yaml "gopkg.in/yaml.v3"
)

type ServiceData struct {
	Name      string   `yaml:"name"`                // Instance name (e.g. "My web page")
	Service   string   `yaml:"service"`             // Service name (e.g. _http._tcp.)
	Domain    string   `yaml:"domain"`              // If blank, assumes "local"
	Port      int      `yaml:"port"`                // Service port
	Records   []string `yaml:"records,omitempty"`   // Service DNS text records
	Responder string   `yaml:"responder,omitempty"` // Optional IP to use instead of our tcp acceptor
}

func (svc ServiceData) FullName() string {
	return fmt.Sprintf("%s.%s.%s",
		strings.Trim(svc.Name, "."),
		strings.Trim(svc.Service, "."),
		strings.Trim(svc.Domain, "."))
}

func (svc ServiceData) Unregister() error {
	if server, err := zeroconf.Register(svc.Name, svc.Service, svc.Domain, svc.Port, svc.Records, nil); err != nil {
		return err
	} else {
		server.Shutdown()
	}
	return nil
}

func svcEntriesToData(services map[string]*zeroconf.ServiceEntry) []ServiceData {
	data := make([]ServiceData, 0)
	for _, svc := range services {
		// filter out empty DNS records
		records := ([]string)(nil)
		for _, txt := range svc.Text {
			if txt = str.Trim(txt); len(txt) > 0 {
				records = append(records, txt)
			}
		}

		data = append(data, ServiceData{
			Name:    svc.Instance,
			Service: svc.Service,
			Domain:  svc.Domain,
			Port:    svc.Port,
			Records: records,
		})
	}
	return data
}

func (mod *ZeroGod) save(address, filename string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if ipServices, found := mod.mapping[address]; found {
		services := svcEntriesToData(ipServices)
		data, err := yaml.Marshal(services)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}

		mod.Info("zeroconf information saved to %s", filename)
	} else {
		return fmt.Errorf("no mDNS information found for address %s", address)
	}

	return nil
}
