package zerogod

import (
	"fmt"
	"io/ioutil"

	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/str"
	yaml "gopkg.in/yaml.v3"
)

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
