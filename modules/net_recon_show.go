package modules

import (
	"fmt"
	"os"
	"sort"
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

func (d *Discovery) Show(by string, expr string) (err error) {
	var targets []*network.Endpoint
	if expr != "" {
		if targets, err = network.ParseEndpoints(expr, d.Session.Lan); err != nil {
			return err
		}
	} else {
		targets = d.Session.Lan.List()
	}
	switch by {
	case "seen":
		sort.Sort(BySeenSorter(targets))
	case "sent":
		sort.Sort(BySentSorter(targets))
	case "rcvd":
		sort.Sort(ByRcvdSorter(targets))
	default:
		sort.Sort(ByAddressSorter(targets))
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

	padCols := []string{"", "", "", "", "", "", ""}
	colNames := []string{"IP", "MAC", "Name", "Vendor", "Sent", "Recvd", "Last Seen"}
	if hasMeta {
		padCols = append(padCols, "")
		colNames = append(colNames, "Meta")
	}

	rows := make([][]string, 0)
	for i, t := range targets {
		rows = append(rows, d.getRow(t, hasMeta)...)
		if i == pad {
			rows = append(rows, padCols)
		}
	}

	tui.Table(os.Stdout, colNames, rows)

	d.Session.Queue.Stats.RLock()
	fmt.Printf("\n%s %s / %s %s / %d pkts / %d errs\n\n",
		tui.Red("↑"),
		humanize.Bytes(d.Session.Queue.Stats.Sent),
		tui.Green("↓"),
		humanize.Bytes(d.Session.Queue.Stats.Received),
		d.Session.Queue.Stats.PktReceived,
		d.Session.Queue.Stats.Errors)
	d.Session.Queue.Stats.RUnlock()

	d.Session.Refresh()

	return nil
}
