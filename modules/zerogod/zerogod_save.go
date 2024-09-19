package zerogod

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v3"
)

func (mod *ZeroGod) save(address, filename string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if ipServices, found := mod.mapping[address]; found {
		data, err := yaml.Marshal(ipServices)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}

		mod.Info("mDNS information saved to %s", filename)
	} else {
		return fmt.Errorf("no mDNS information found for address %s", address)
	}

	return nil
}
