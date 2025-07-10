package zerogod

import (
	"errors"
	"fmt"
	"strings"

	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func (mod *ZeroGod) show(filter string, withData bool) error {
	if mod.browser == nil {
		return errors.New("use 'zerogod.discovery on' to start the discovery first")
	}

	fmt.Fprintf(mod.Session.Events.Stdout, "\n")

	entries := mod.browser.ServicesByAddress(filter)

	for _, entry := range entries {
		if endpoint := mod.Session.Lan.GetByIp(entry.Address); endpoint != nil {
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s (%s)%s\n",
				tui.Bold(endpoint.IpAddress),
				tui.Dim(endpoint.Vendor),
				ops.Ternary(endpoint.Hostname == "", "", " "+tui.Bold(endpoint.Hostname)))
		} else {
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s\n", tui.Bold(entry.Address))
		}

		for _, svc := range entry.Services {
			ip := ""
			if len(svc.AddrIPv4) > 0 {
				ip = svc.AddrIPv4[0].String()
			} else if len(svc.AddrIPv6) > 0 {
				ip = svc.AddrIPv6[0].String()
			} else {
				ip = svc.HostName
			}

			svcDesc := ""
			svcName := strings.SplitN(svc.Service, ".", 2)[0]
			if desc, found := KNOWN_SERVICES[svcName]; found {
				svcDesc = tui.Dim(fmt.Sprintf(" %s", desc))
			}

			fmt.Fprintf(mod.Session.Events.Stdout, "  %s%s %s:%s\n",
				tui.Green(svc.ServiceInstanceName()),
				svcDesc,
				ip,
				tui.Red(fmt.Sprintf("%d", svc.Port)),
			)

			numFields := len(svc.Text)
			if withData {
				if numFields > 0 {
					columns := []string{"key", "value"}
					rows := make([][]string, 0)

					for _, field := range svc.Text {
						if field = str.Trim(field); len(field) > 0 {
							keyval := strings.SplitN(field, "=", 2)
							key := str.Trim(keyval[0])
							val := str.Trim(keyval[1])

							if key != "" || val != "" {
								rows = append(rows, []string{
									key,
									val,
								})
							}
						}
					}

					if len(rows) == 0 {
						fmt.Fprintf(mod.Session.Events.Stdout, "    %s\n", tui.Dim("no data"))
					} else {
						tui.Table(mod.Session.Events.Stdout, columns, rows)
						fmt.Fprintf(mod.Session.Events.Stdout, "\n")
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
