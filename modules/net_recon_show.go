package modules

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/dustin/go-humanize"
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

func rankByProtoHits(protos map[string]uint64) (ProtoPairList, uint64) {
	pl := make(ProtoPairList, len(protos))
	max := uint64(0)
	i := 0
	for k, v := range protos {
		pl[i] = ProtoPair{k, v}
		if v > max {
			max = v
		}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl, max
}

func (d *Discovery) getRow(e *network.Endpoint, withMeta bool) []string {
	sinceStarted := time.Since(d.Session.StartedAt)
	sinceFirstSeen := time.Since(e.FirstSeen)

	addr := e.IpAddress
	mac := e.HwAddress
	if d.Session.Lan.WasMissed(e.HwAddress) == true {
		// if endpoint was not found in ARP at least once
		addr = core.Dim(addr)
		mac = core.Dim(mac)
	} else if sinceStarted > (justJoinedTimeInterval*2) && sinceFirstSeen <= justJoinedTimeInterval {
		// if endpoint was first seen in the last 10 seconds
		addr = core.Bold(addr)
		mac = core.Bold(mac)
	}

	name := ""
	if e == d.Session.Interface {
		name = e.Name()
	} else if e == d.Session.Gateway {
		name = "gateway"
	} else if e.Alias != "" {
		name = core.Green(e.Alias)
	} else if e.Hostname != "" {
		name = core.Yellow(e.Hostname)
	}

	var traffic *packets.Traffic
	var found bool
	if traffic, found = d.Session.Queue.Traffic[e.IpAddress]; found == false {
		traffic = &packets.Traffic{}
	}

	seen := e.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(e.LastSeen)
	if sinceStarted > aliveTimeInterval && sinceLastSeen <= aliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = core.Bold(seen)
	} else if sinceLastSeen <= presentTimeInterval {
		// if endpoint seen in the last 60 seconds
	} else {
		// not seen in a while
		seen = core.Dim(seen)
	}

	row := []string{
		addr,
		mac,
		name,
		e.Vendor,
		humanize.Bytes(traffic.Sent),
		humanize.Bytes(traffic.Received),
		seen,
	}

	if withMeta {
		metas := []string{}
		e.Meta.Each(func(name string, value interface{}) {
			metas = append(metas, fmt.Sprintf("%s: %s", name, value.(string)))
		})

		row = append(row, strings.Join(metas, "\n"))
	}

	return row
}

func (d *Discovery) Show(by string) error {
	targets := d.Session.Lan.List()
	if by == "seen" {
		sort.Sort(BySeenSorter(targets))
	} else if by == "sent" {
		sort.Sort(BySentSorter(targets))
	} else if by == "rcvd" {
		sort.Sort(ByRcvdSorter(targets))
	} else {
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
	for _, t := range targets {
		if t.Meta.Empty() == false {
			hasMeta = true
			break
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
		rows = append(rows, d.getRow(t, hasMeta))
		if i == pad {
			rows = append(rows, padCols)
		}
	}

	core.AsTable(os.Stdout, colNames, rows)

	d.Session.Queue.Stats.RLock()
	fmt.Printf("\n%s %s / %s %s / %d pkts / %d errs\n\n",
		core.Red("↑"),
		humanize.Bytes(d.Session.Queue.Stats.Sent),
		core.Green("↓"),
		humanize.Bytes(d.Session.Queue.Stats.Received),
		d.Session.Queue.Stats.PktReceived,
		d.Session.Queue.Stats.Errors)
	d.Session.Queue.Stats.RUnlock()

	d.Session.Refresh()

	return nil
}
