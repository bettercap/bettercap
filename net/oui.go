package net

import (
	"strings"
)

var (
	oui = make(map[string]string)
)

func OuiInit() {
	bytes, err := Asset("net/oui.dat")
	if err != nil {
		panic(err)
	}

	data := string(bytes)
	lines := strings.Split(data, "\n")

	for lineno, line := range lines {
		line = strings.Trim(line, " \n\r\t")
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			log.Warningf("Skipping line %d '%s'\n", lineno+1, line)
			continue
		}

		prefix := strings.ToLower(strings.Trim(parts[0], " \n\r\t"))
		vendor := strings.Trim(parts[1], " \n\r\t")

		oui[prefix] = vendor
	}

	log.Debugf("Loaded %d vendors signatures.\n", len(oui))
}

func OuiLookup(mac string) string {
	octects := strings.Split(mac, ":")
	if len(octects) > 3 {
		prefix := octects[0] + octects[1] + octects[2]

		if vendor, found := oui[prefix]; found == true {
			return vendor
		}
	} else {
		log.Warningf("Unexpected mac '%s' in net.OuiLookup\n", mac)
	}

	return "???"
}
