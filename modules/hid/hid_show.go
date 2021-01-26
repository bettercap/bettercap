package hid

import (
	"sort"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
)

var (
	AliveTimeInterval      = time.Duration(5) * time.Minute
	PresentTimeInterval    = time.Duration(1) * time.Minute
	JustJoinedTimeInterval = time.Duration(10) * time.Second
)

func (mod *HIDRecon) getRow(dev *network.HIDDevice) []string {
	sinceLastSeen := time.Since(dev.LastSeen)
	seen := dev.LastSeen.Format("15:04:05")

	if sinceLastSeen <= JustJoinedTimeInterval {
		seen = tui.Bold(seen)
	} else if sinceLastSeen > PresentTimeInterval {
		seen = tui.Dim(seen)
	}

	return []string{
		dev.Address,
		dev.Type.String(),
		dev.Channels(),
		humanize.Bytes(dev.PayloadsSize()),
		seen,
	}
}

func (mod *HIDRecon) doFilter(dev *network.HIDDevice) bool {
	if mod.selector.Expression == nil {
		return true
	}
	return mod.selector.Expression.MatchString(dev.Address)
}

func (mod *HIDRecon) doSelection() (err error, devices []*network.HIDDevice) {
	if err = mod.selector.Update(); err != nil {
		return
	}

	devices = mod.Session.HID.Devices()
	filtered := []*network.HIDDevice{}
	for _, dev := range devices {
		if mod.doFilter(dev) {
			filtered = append(filtered, dev)
		}
	}
	devices = filtered

	switch mod.selector.SortField {
	case "mac":
		sort.Sort(ByHIDMacSorter(devices))
	case "seen":
		sort.Sort(ByHIDSeenSorter(devices))
	}

	// default is asc
	if mod.selector.Sort == "desc" {
		// from https://github.com/golang/go/wiki/SliceTricks
		for i := len(devices)/2 - 1; i >= 0; i-- {
			opp := len(devices) - 1 - i
			devices[i], devices[opp] = devices[opp], devices[i]
		}
	}

	if mod.selector.Limit > 0 {
		limit := mod.selector.Limit
		max := len(devices)
		if limit > max {
			limit = max
		}
		devices = devices[0:limit]
	}

	return
}

func (mod *HIDRecon) colNames() []string {
	colNames := []string{"MAC", "Type", "Channels", "Data", "Seen"}
	switch mod.selector.SortField {
	case "mac":
		colNames[0] += " " + mod.selector.SortSymbol
	case "seen":
		colNames[4] += " " + mod.selector.SortSymbol
	}
	return colNames
}

func (mod *HIDRecon) Show() (err error) {
	var devices []*network.HIDDevice
	if err, devices = mod.doSelection(); err != nil {
		return
	}

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, mod.getRow(dev))
	}

	tui.Table(mod.Session.Events.Stdout, mod.colNames(), rows)

	if mod.sniffAddrRaw == nil {
		mod.Printf("\nchannel:%d\n\n", mod.channel)
	} else {
		mod.Printf("\nchannel:%d sniffing:%s\n\n", mod.channel, tui.Red(mod.sniffAddr))
	}

	if len(rows) > 0 {
		mod.Session.Refresh()
	}

	return nil
}
