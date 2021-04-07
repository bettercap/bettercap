package net_recon

import (
	"fmt"
	"github.com/bettercap/bettercap/modules/syn_scan"
	"sort"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/islazy/str"
)

const (
	AliveTimeInterval      = time.Duration(10) * time.Second
	PresentTimeInterval    = time.Duration(1) * time.Minute
	JustJoinedTimeInterval = time.Duration(10) * time.Second
)

type ProtoPair struct {
	Protocol string
	Hits     uint64
}

type ProtoPairList []ProtoPair

func (p ProtoPairList) Len() int           { return len(p) }
func (p ProtoPairList) Less(i, j int) bool { return p[i].Hits < p[j].Hits }
func (p ProtoPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (mod *Discovery) getRow(e *network.Endpoint, withMeta bool) [][]string {
	sinceStarted := time.Since(mod.Session.StartedAt)
	sinceFirstSeen := time.Since(e.FirstSeen)

	addr := e.IpAddress
	mac := e.HwAddress
	if mod.Session.Lan.WasMissed(e.HwAddress) {
		// if endpoint was not found in ARP at least once
		addr = tui.Dim(addr)
		mac = tui.Dim(mac)
	} else if sinceStarted > (JustJoinedTimeInterval*2) && sinceFirstSeen <= JustJoinedTimeInterval {
		// if endpoint was first seen in the last 10 seconds
		addr = tui.Bold(addr)
		mac = tui.Bold(mac)
	}

	name := ""
	if e == mod.Session.Interface {
		name = e.Name()
	} else if e == mod.Session.Gateway {
		name = "gateway"
	} else if e.Alias != "" {
		name = tui.Green(e.Alias)
	} else if e.Hostname != "" {
		name = tui.Yellow(e.Hostname)
	}

	var traffic *packets.Traffic
	var found bool
	var v interface{}
	if v, found = mod.Session.Queue.Traffic.Load(e.IpAddress); !found {
		traffic = &packets.Traffic{}
	} else {
		traffic = v.(*packets.Traffic)
	}

	seen := e.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(e.LastSeen)
	if sinceStarted > AliveTimeInterval && sinceLastSeen <= AliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = tui.Bold(seen)
	} else if sinceLastSeen <= PresentTimeInterval {
		// if endpoint seen in the last 60 seconds
	} else {
		// not seen in a while
		seen = tui.Dim(seen)
	}

	row := []string{
		addr,
		mac,
		name,
		tui.Dim(e.Vendor),
		humanize.Bytes(traffic.Sent),
		humanize.Bytes(traffic.Received),
		seen,
	}

	if !withMeta {
		return [][]string{row}
	} else if e.Meta.Empty() {
		return [][]string{append(row, tui.Dim("-"))}
	}

	metas := []string{}
	e.Meta.Each(func(name string, value interface{}) {
		s := ""
		if sv, ok := value.(string); ok {
			s = sv
		} else {
			s = fmt.Sprintf("%+v", value)
		}

		metas = append(metas, fmt.Sprintf("%s:%s", tui.Green(name), tui.Yellow(s)))
	})
	sort.Strings(metas)

	rows := make([][]string, 0, len(metas))
	for i, m := range metas {
		if i == 0 {
			rows = append(rows, append(row, m))
		} else {
			rows = append(rows, []string{"", "", "", "", "", "", "", m})
		}
	}

	return rows
}

func (mod *Discovery) doFilter(target *network.Endpoint) bool {
	if mod.selector.Expression == nil {
		return true
	}
	return mod.selector.Expression.MatchString(target.IpAddress) ||
		mod.selector.Expression.MatchString(target.Ip6Address) ||
		mod.selector.Expression.MatchString(target.HwAddress) ||
		mod.selector.Expression.MatchString(target.Hostname) ||
		mod.selector.Expression.MatchString(target.Alias) ||
		mod.selector.Expression.MatchString(target.Vendor)
}

func (mod *Discovery) doSelection(arg string) (err error, targets []*network.Endpoint) {
	if err = mod.selector.Update(); err != nil {
		return
	}

	if arg != "" {
		if targets, err = network.ParseEndpoints(arg, mod.Session.Lan); err != nil {
			return
		}
	} else {
		targets = mod.Session.Lan.List()
	}

	filtered := []*network.Endpoint{}
	for _, target := range targets {
		if mod.doFilter(target) {
			filtered = append(filtered, target)
		}
	}
	targets = filtered

	switch mod.selector.SortField {
	case "ip":
		sort.Sort(ByIpSorter(targets))
	case "mac":
		sort.Sort(ByMacSorter(targets))
	case "seen":
		sort.Sort(BySeenSorter(targets))
	case "sent":
		sort.Sort(BySentSorter(targets))
	case "rcvd":
		sort.Sort(ByRcvdSorter(targets))
	default:
		sort.Sort(ByAddressSorter(targets))
	}

	// default is asc
	if mod.selector.Sort == "desc" {
		// from https://github.com/golang/go/wiki/SliceTricks
		for i := len(targets)/2 - 1; i >= 0; i-- {
			opp := len(targets) - 1 - i
			targets[i], targets[opp] = targets[opp], targets[i]
		}
	}

	if mod.selector.Limit > 0 {
		limit := mod.selector.Limit
		max := len(targets)
		if limit > max {
			limit = max
		}
		targets = targets[0:limit]
	}

	return
}

func (mod *Discovery) colNames(hasMeta bool) []string {
	colNames := []string{"IP", "MAC", "Name", "Vendor", "Sent", "Recvd", "Seen"}
	if hasMeta {
		colNames = append(colNames, "Meta")
	}

	switch mod.selector.SortField {
	case "mac":
		colNames[1] += " " + mod.selector.SortSymbol
	case "sent":
		colNames[4] += " " + mod.selector.SortSymbol
	case "rcvd":
		colNames[5] += " " + mod.selector.SortSymbol
	case "seen":
		colNames[6] += " " + mod.selector.SortSymbol
	case "ip":
		colNames[0] += " " + mod.selector.SortSymbol
	}

	return colNames
}

func (mod *Discovery) showStatusBar() {
	parts := []string{
		fmt.Sprintf("%s %s", tui.Red("↑"), humanize.Bytes(mod.Session.Queue.Stats.Sent)),
		fmt.Sprintf("%s %s", tui.Green("↓"), humanize.Bytes(mod.Session.Queue.Stats.Received)),
		fmt.Sprintf("%d pkts", mod.Session.Queue.Stats.PktReceived),
	}

	if nErrors := mod.Session.Queue.Stats.Errors; nErrors > 0 {
		parts = append(parts, fmt.Sprintf("%d errs", nErrors))
	}

	mod.Printf("\n%s\n\n", strings.Join(parts, " / "))
}

func (mod *Discovery) Show(arg string) (err error) {
	var targets []*network.Endpoint
	if err, targets = mod.doSelection(arg); err != nil {
		return
	}

	pad := 1
	if mod.Session.Interface.HwAddress == mod.Session.Gateway.HwAddress {
		pad = 0
		targets = append([]*network.Endpoint{mod.Session.Interface}, targets...)
	} else {
		targets = append([]*network.Endpoint{mod.Session.Interface, mod.Session.Gateway}, targets...)
	}

	hasMeta := false
	if err, showMeta := mod.BoolParam("net.show.meta"); err != nil {
		return err
	} else if showMeta {
		for _, t := range targets {
			if !t.Meta.Empty() {
				hasMeta = true
				break
			}
		}
	}

	colNames := mod.colNames(hasMeta)
	padCols := make([]string, len(colNames))

	rows := make([][]string, 0)
	for i, t := range targets {
		rows = append(rows, mod.getRow(t, hasMeta)...)
		if i == pad {
			rows = append(rows, padCols)
		}
	}

	tui.Table(mod.Session.Events.Stdout, colNames, rows)

	mod.showStatusBar()

	mod.Session.Refresh()

	return nil
}

func (mod *Discovery) showMeta(arg string) (err error) {
	var targets []*network.Endpoint
	if err, targets = mod.doSelection(arg); err != nil {
		return
	}

	colNames := []string{"Name", "Value"}
	any := false

	for _, t := range targets {
		keys := []string{}

		t.Meta.Each(func(name string, value interface{}) {
			keys = append(keys, name)
		})

		if len(keys) > 0 {
			sort.Strings(keys)
			rows := [][]string{
				{
					tui.Green("address"),
					t.IP.String(),
				},
			}

			for _, k := range keys {
				meta := t.Meta.Get(k)
				val := ""
				if s, ok := meta.(string); ok {
					val = s
				} else if ports, ok := meta.(map[int]*syn_scan.OpenPort); ok {
					val = "ports: "
					for _, info := range ports {
						val += fmt.Sprintf("%s:%d", info.Proto, info.Port)
						if info.Service != "" {
							val += fmt.Sprintf("(%s)", info.Service)
						}
						if info.Banner != "" {
							val += fmt.Sprintf(" [%s]", info.Banner)
						}
						val += " "
					}
					val = str.Trim(val)
				} else {
					val = fmt.Sprintf("%#v", meta)
				}
				rows = append(rows, []string{
					tui.Green(k),
					tui.Yellow(val),
				})
			}

			any = true
			tui.Table(mod.Session.Events.Stdout, colNames, rows)
		}
	}

	if any {
		mod.Session.Refresh()
	}

	return nil
}
