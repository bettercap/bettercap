package routing

import (
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/evilsocket/islazy/str"
)

var (
	routeHeadings    []string
	whitespaceParser = regexp.MustCompile(`\s+`)
)

func update() ([]Route, error) {
	table = make([]Route, 0)

	output, err := core.Exec("netstat", []string{"-r", "-n"})
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(output, "\n") {
		if line = str.Trim(line); len(line) != 0 {
			parts := whitespaceParser.Split(line, -1)
			if parts[0] == "Kernel" {
				continue
			}

			if parts[0] == "Destination" {
				routeHeadings = parts
				continue
			}

			route := Route{}
			for i, s := range parts {
				switch routeHeadings[i] {
				case "Destination":
					route.Destination = s
					break
				case "Flags":
					route.Flags = s
					break
				case "Gateway":
					route.Gateway = s
					break
				case "Iface":
					route.Device = s
					break
				}
			}

			route.Default = strings.Contains(route.Flags, "G")

			if route.Destination != "" {
				if strings.ContainsRune(route.Destination, '.') || strings.ContainsRune(route.Gateway, '.') {
					route.Type = "IPv4"
				} else {
					route.Type = "IPv6"
				}
			}

			table = append(table, route)
		}
	}

	return table, nil
}
