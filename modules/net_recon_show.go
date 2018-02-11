package modules

import (
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
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

func (d *Discovery) getRow(e *net.Endpoint) []string {
	sinceStarted := time.Since(d.Session.StartedAt)
	sinceFirstSeen := time.Since(e.FirstSeen)

	addr := e.IpAddress
	mac := e.HwAddress
	if d.Session.Targets.WasMissed(e.HwAddress) == true {
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

	return []string{
		addr,
		mac,
		name,
		e.Vendor,
		humanize.Bytes(traffic.Sent),
		humanize.Bytes(traffic.Received),
		seen,
	}
}

func (d *Discovery) showTable(header []string, rows [][]string) {
	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetColWidth(80)
	table.AppendBulk(rows)
	table.Render()
}

func (d *Discovery) Show(by string) error {
	targets := d.Session.Targets.List()
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
		targets = append([]*net.Endpoint{d.Session.Interface}, targets...)
	} else {
		targets = append([]*net.Endpoint{d.Session.Interface, d.Session.Gateway}, targets...)
	}

	rows := make([][]string, 0)
	for i, t := range targets {
		rows = append(rows, d.getRow(t))
		if i == pad {
			rows = append(rows, []string{"", "", "", "", "", "", ""})
		}
	}

	d.showTable([]string{"IP", "MAC", "Name", "Vendor", "Sent", "Recvd", "Last Seen"}, rows)

	fmt.Printf("\n%s %s / %s %s / %d pkts / %d errs\n\n",
		core.Red("↑"),
		humanize.Bytes(atomic.LoadUint64(&d.Session.Queue.Stats.Sent)),
		core.Green("↓"),
		humanize.Bytes(atomic.LoadUint64(&d.Session.Queue.Stats.Received)),
		atomic.LoadUint64(&d.Session.Queue.Stats.PktReceived),
		atomic.LoadUint64(&d.Session.Queue.Stats.Errors))

	s := EventsStream{}
	events := d.Session.Events.Sorted()
	size := len(events)

	if size > 0 {
		max := 20
		if size > max {
			from := size - max
			size = max
			events = events[from:]
		}

		fmt.Printf("Last %d events:\n\n", size)

		for _, e := range events {
			s.View(e, false)
		}

		fmt.Println()
	}

	/*
		Last events are more useful than this histogram and vertical scroll
		isn't infinite :)

			rows = make([][]string, 0)
			protos, maxPackets := rankByProtoHits(d.Session.Queue.Protos)
			maxBarWidth := 70

			for _, p := range protos {
				width := int(float32(maxBarWidth) * (float32(p.Hits) / float32(maxPackets)))
				bar := ""
				for i := 0; i < width; i++ {
					bar += "▇"
				}

				rows = append(rows, []string{p.Protocol, fmt.Sprintf("%s %d", bar, p.Hits)})
			}

			d.showTable([]string{"Proto", "# Packets"}, rows)
	*/

	d.Session.Refresh()

	return nil
}
