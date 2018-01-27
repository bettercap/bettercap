package modules

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

var (
	aliveTimeInterval   = time.Duration(10) * time.Second
	presentTimeInterval = time.Duration(1) * time.Minute
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

func (d *Discovery) Show(by string) error {
	d.Session.Targets.Lock()
	d.Session.Queue.Lock()
	defer d.Session.Targets.Unlock()
	defer d.Session.Queue.Unlock()

	iface := d.Session.Interface
	gw := d.Session.Gateway

	data := [][]string{
		[]string{core.Green("interface"), core.Bold(iface.Name()), iface.IpAddress, iface.HwAddress, core.Dim(iface.Vendor)},
		[]string{core.Green("gateway"), core.Bold(gw.Hostname), gw.IpAddress, gw.HwAddress, core.Dim(gw.Vendor)},
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.SetColWidth(80)
	table.AppendBulk(data)
	table.Render()

	fmt.Println()

	nTargets := len(d.Session.Targets.Targets)
	if nTargets == 0 {
		fmt.Println(core.Dim("No endpoints discovered so far.\n"))
	} else {
		targets := make([]*net.Endpoint, 0, nTargets)
		for _, t := range d.Session.Targets.Targets {
			targets = append(targets, t)
		}

		if by == "seen" {
			sort.Sort(BySeenSorter(targets))
		} else if by == "sent" {
			sort.Sort(BySentSorter(targets))
		} else if by == "rcvd" {
			sort.Sort(ByRcvdSorter(targets))
		} else {
			sort.Sort(ByAddressSorter(targets))
		}

		data = make([][]string, nTargets)
		for i, t := range targets {
			var traffic *packets.Traffic
			var found bool

			if traffic, found = d.Session.Queue.Traffic[t.IpAddress]; found == false {
				traffic = &packets.Traffic{}
			}

			seen := t.LastSeen.Format("15:04:05")
			sinceLastSeen := time.Since(t.LastSeen)
			if sinceLastSeen <= aliveTimeInterval {
				seen = core.Bold(seen)
			} else if sinceLastSeen <= presentTimeInterval {

			} else {
				seen = core.Dim(seen)
			}

			data[i] = []string{
				t.IpAddress,
				t.HwAddress,
				core.Yellow(t.Hostname),
				t.Vendor,
				humanize.Bytes(traffic.Sent),
				humanize.Bytes(traffic.Received),
				seen,
			}
		}

		table = tablewriter.NewWriter(os.Stdout)

		table.SetHeader([]string{"IP", "MAC", "Hostname", "Vendor", "Sent", "Recvd", "Last Seen"})
		table.SetColWidth(80)
		table.AppendBulk(data)
		table.Render()

		fmt.Println()
	}

	row := []string{
		humanize.Bytes(d.Session.Queue.Sent),
		humanize.Bytes(d.Session.Queue.Received),
		fmt.Sprintf("%d", d.Session.Queue.PktReceived),
		fmt.Sprintf("%d", d.Session.Queue.Errors),
	}

	table = tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Sent", "Sniffed", "# Packets", "Errors"})
	table.SetColWidth(80)
	table.Append(row)
	table.Render()

	fmt.Println()

	table = tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(80)

	protos, maxPackets := rankByProtoHits(d.Session.Queue.Protos)
	maxBarWidth := 70

	for _, p := range protos {
		width := int(float32(maxBarWidth) * (float32(p.Hits) / float32(maxPackets)))
		bar := ""
		for i := 0; i < width; i++ {
			bar += "▇"
		}

		table.Append([]string{p.Protocol, fmt.Sprintf("%s %d", bar, p.Hits)})
	}

	table.SetHeader([]string{"Proto", "# Packets"})
	table.Render()

	return nil
}
