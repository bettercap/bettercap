package zerogod

import (
	"fmt"
	"sort"

	"github.com/bettercap/bettercap/v2/zeroconf"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

type entry struct {
	ip       string
	services map[string]*zeroconf.ServiceEntry
}

func (mod *ZeroGod) show(filter string, withData bool) error {
	fmt.Fprintf(mod.Session.Events.Stdout, "\n")

	// convert to list for sorting
	entries := make([]entry, 0)
	for ip, services := range mod.mapping {
		if filter == "" || ip == filter {
			entries = append(entries, entry{ip, services})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ip < entries[j].ip
	})

	for _, entry := range entries {
		if endpoint := mod.Session.Lan.GetByIp(entry.ip); endpoint != nil {
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s (%s)\n", tui.Bold(endpoint.IpAddress), tui.Dim(endpoint.Vendor))
		} else {
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s\n", tui.Bold(entry.ip))
		}

		for name, svc := range entry.services {
			fmt.Fprintf(mod.Session.Events.Stdout, "  %s (%s) [%v / %v]:%s\n",
				tui.Green(name),
				tui.Dim(svc.HostName),
				svc.AddrIPv4,
				svc.AddrIPv6,
				tui.Red(fmt.Sprintf("%d", svc.Port)),
			)

			numFields := len(svc.Text)
			if withData {
				if numFields > 0 {
					for _, field := range svc.Text {
						if field = str.Trim(field); len(field) > 0 {
							fmt.Fprintf(mod.Session.Events.Stdout, "    %s\n", field)
						}
					}
				} else {
					fmt.Fprintf(mod.Session.Events.Stdout, "    %s\n", tui.Dim("no data"))
				}
			} else {
				if numFields > 0 {
					fmt.Fprintf(mod.Session.Events.Stdout, "    <%d records>\n", numFields)
				} else {
					fmt.Fprintf(mod.Session.Events.Stdout, "    %s\n", tui.Dim("<no records>"))
				}
			}
		}

		fmt.Fprintf(mod.Session.Events.Stdout, "\n")
	}

	if len(entries) > 0 {
		mod.Session.Refresh()
	}

	return nil
}
