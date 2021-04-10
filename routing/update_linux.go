package routing

import (
	"github.com/bettercap/bettercap/core"
	"github.com/evilsocket/islazy/str"
	"regexp"
	"strings"
)

var parser = regexp.MustCompile(`^(.+)\sdev\s([^\s]+)\s(.+)$`)

func update() ([]Route, error) {
	table = make([]Route, 0)

	for ip, inet := range map[RouteType]string{IPv4: "inet", IPv6: "inet6"} {
		output, err := core.Exec("ip", []string{"-f", inet, "route"})
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(output, "\n") {
			if line = str.Trim(line); len(line) > 0 {
				matches := parser.FindStringSubmatch(line)
				if num := len(matches); num == 4 {
					route := Route{
						Type:        ip,
						Destination: matches[1],
						Device:      matches[2],
						Flags:       matches[3],
						Default:     strings.Index(matches[1], "default ") == 0,
					}

					if idx := strings.Index(route.Destination, " via "); idx >= 0 {
						route.Gateway = route.Destination[idx + len(" via "):]
						route.Destination = route.Destination[:idx]
					}

					table = append(table, route)
				}
			}
		}
	}

	return table, nil
}
