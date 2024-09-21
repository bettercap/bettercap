package zerogod

import (
	"errors"
	"fmt"
	"strings"

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
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s (%s)\n", tui.Bold(endpoint.IpAddress), tui.Dim(endpoint.Vendor))
		} else {
			fmt.Fprintf(mod.Session.Events.Stdout, "* %s\n", tui.Bold(entry.Address))
		}

		for _, svc := range entry.Services {
			fmt.Fprintf(mod.Session.Events.Stdout, "  %s (%s) [%v / %v]:%s\n",
				tui.Green(svc.ServiceInstanceName()),
				tui.Dim(svc.HostName),
				svc.AddrIPv4,
				svc.AddrIPv6,
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
							rows = append(rows, []string{
								keyval[0],
								keyval[1],
							})
						}
					}

					tui.Table(mod.Session.Events.Stdout, columns, rows)
					fmt.Fprintf(mod.Session.Events.Stdout, "\n")

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
