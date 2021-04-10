package routing

import (
	"github.com/bettercap/bettercap/core"
	"github.com/evilsocket/islazy/str"
	"regexp"
	"strings"
)

var parser = regexp.MustCompile(`^.+\d+\s+([^\s]+)\s+\d+\s+(.+)$`)

func update() ([]Route, error) {
	table = make([]Route, 0)

	for ip, inet := range map[RouteType]string{IPv4: "ipv4", IPv6: "ipv6"} {
		output, err := core.Exec("netsh", []string{"interface", inet, "show", "route"})
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(output, "\n") {
			if line = str.Trim(line); len(line) > 0 {
				matches := parser.FindStringSubmatch(line)
				if num := len(matches); num == 3 {
					route := Route{
						Type:        ip,
						Destination: matches[1],
						Device:      matches[2],
					}

					if route.Destination == "0.0.0.0/0" || route.Destination == "::/0" {
						route.Default = true
						route.Gateway = route.Device
						route.Device = ""
					}

					table = append(table, route)
				}
			}
		}
	}

	return table, nil
}
