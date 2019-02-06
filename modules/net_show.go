package modules

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
)

var (
	aliveTimeInterval      = time.Duration(10) * time.Second
	presentTimeInterval    = time.Duration(1) * time.Minute
	justJoinedTimeInterval = time.Duration(10) * time.Second
)

type ProtoPair struct {
	Protocol string
	Hits     uint64
}

type ProtoPairList []ProtoPair

func (p ProtoPairList) Len() int           { return len(p) }
func (p ProtoPairList) Less(i, j int) bool { return p[i].Hits < p[j].Hits }
func (p ProtoPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (d *Discovery) getRow(e *network.Endpoint, withMeta bool) [][]string {
	sinceStarted := time.Since(d.Session.StartedAt)
	sinceFirstSeen := time.Since(e.FirstSeen)

	addr := e.IpAddress
	mac := e.HwAddress
	if d.Session.Lan.WasMissed(e.HwAddress) {
		// if endpoint was not found in ARP at least once
		addr = tui.Dim(addr)
		mac = tui.Dim(mac)
	} else if sinceStarted > (justJoinedTimeInterval*2) && sinceFirstSeen <= justJoinedTimeInterval {
		// if endpoint was first seen in the last 10 seconds
		addr = tui.Bold(addr)
		mac = tui.Bold(mac)
	}

	name := ""
	if e == d.Session.Interface {
		name = e.Name()
	} else if e == d.Session.Gateway {
		name = "gateway"
	} else if e.Alias != "" {
		name = tui.Green(e.Alias)
	} else if e.Hostname != "" {
		name = tui.Yellow(e.Hostname)
	}

	var traffic *packets.Traffic
	var found bool
	if traffic, found = d.Session.Queue.Traffic[e.IpAddress]; !found {
		traffic = &packets.Traffic{}
	}

	seen := e.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(e.LastSeen)
	if sinceStarted > aliveTimeInterval && sinceLastSeen <= aliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = tui.Bold(seen)
	} else if sinceLastSeen <= presentTimeInterval {
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
		metas = append(metas, fmt.Sprintf("%s:%s", tui.Green(name), tui.Yellow(value.(string))))
	})
	sort.Strings(metas)

	rows := [][]string{}
	for i, m := range metas {
		if i == 0 {
			rows = append(rows, append(row, m))
		} else {
			rows = append(rows, []string{"", "", "", "", "", "", "", m})
		}
	}

	return rows
}

func (d *Discovery) doFilter(target *network.Endpoint) bool {
	if d.selector.Expression == nil {
		return true
	}
	return d.selector.Expression.MatchString(target.IpAddress) ||
		d.selector.Expression.MatchString(target.Ip6Address) ||
		d.selector.Expression.MatchString(target.HwAddress) ||
		d.selector.Expression.MatchString(target.Hostname) ||
		d.selector.Expression.MatchString(target.Alias) ||
		d.selector.Expression.MatchString(target.Vendor)
}

func (d *Discovery) doSelection(arg string) (err error, targets []*network.Endpoint) {
	if err = d.selector.Update(); err != nil {
		return
	}

	if arg != "" {
		if targets, err = network.ParseEndpoints(arg, d.Session.Lan); err != nil {
			return
		}
	} else {
		targets = d.Session.Lan.List()
	}

	filtered := []*network.Endpoint{}
	for _, target := range targets {
		if d.doFilter(target) {
			filtered = append(filtered, target)
		}
	}
	targets = filtered

	switch d.selector.SortField {
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
	if d.selector.Sort == "desc" {
		// from https://github.com/golang/go/wiki/SliceTricks
		for i := len(targets)/2 - 1; i >= 0; i-- {
			opp := len(targets) - 1 - i
			targets[i], targets[opp] = targets[opp], targets[i]
		}
	}

	if d.selector.Limit > 0 {
		limit := d.selector.Limit
		max := len(targets)
		if limit > max {
			limit = max
		}
		targets = targets[0:limit]
	}

	return
}

func (d *Discovery) colNames(hasMeta bool) []string {
	colNames := []string{"IP", "MAC", "Name", "Vendor", "Sent", "Recvd", "Last Seen"}
	if hasMeta {
		colNames = append(colNames, "Meta")
	}

	switch d.selector.SortField {
	case "mac":
		colNames[1] += " " + d.selector.SortSymbol
	case "sent":
		colNames[4] += " " + d.selector.SortSymbol
	case "rcvd":
		colNames[5] += " " + d.selector.SortSymbol
	case "seen":
		colNames[6] += " " + d.selector.SortSymbol
	case "ip":
		colNames[0] += " " + d.selector.SortSymbol
	}

	return colNames
}

func (d *Discovery) showStatusBar() {
	d.Session.Queue.Stats.RLock()
	defer d.Session.Queue.Stats.RUnlock()

	parts := []string{
		fmt.Sprintf("%s %s", tui.Red("↑"), humanize.Bytes(d.Session.Queue.Stats.Sent)),
		fmt.Sprintf("%s %s", tui.Green("↓"), humanize.Bytes(d.Session.Queue.Stats.Received)),
		fmt.Sprintf("%d pkts", d.Session.Queue.Stats.PktReceived),
	}

	if nErrors := d.Session.Queue.Stats.Errors; nErrors > 0 {
		parts = append(parts, fmt.Sprintf("%d errs", nErrors))
	}

	fmt.Printf("\n%s\n\n", strings.Join(parts, " / "))
}

func (d *Discovery) Show(arg string) (err error) {
	var targets []*network.Endpoint
	if err, targets = d.doSelection(arg); err != nil {
		return
	}

	pad := 1
	if d.Session.Interface.HwAddress == d.Session.Gateway.HwAddress {
		pad = 0
		targets = append([]*network.Endpoint{d.Session.Interface}, targets...)
	} else {
		targets = append([]*network.Endpoint{d.Session.Interface, d.Session.Gateway}, targets...)
	}

	hasMeta := false
	if err, showMeta := d.BoolParam("net.show.meta"); err != nil {
		return err
	} else if showMeta {
		for _, t := range targets {
			if !t.Meta.Empty() {
				hasMeta = true
				break
			}
		}
	}

	colNames := d.colNames(hasMeta)
	padCols := make([]string, len(colNames))

	rows := make([][]string, 0)
	for i, t := range targets {
		rows = append(rows, d.getRow(t, hasMeta)...)
		if i == pad {
			rows = append(rows, padCols)
		}
	}

	tui.Table(os.Stdout, colNames, rows)

	d.showStatusBar()

	d.Session.Refresh()

	return nil
}
