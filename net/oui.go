package net

import (
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
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

	for _, line := range lines {
		line = core.Trim(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		prefix := strings.ToLower(core.Trim(parts[0]))
		vendor := core.Trim(parts[1])

		oui[prefix] = vendor
	}
}

func OuiLookup(mac string) string {
	octects := strings.Split(mac, ":")
	if len(octects) > 3 {
		prefix := octects[0] + octects[1] + octects[2]

		if vendor, found := oui[prefix]; found == true {
			return vendor
		}
	}
	return ""
}
