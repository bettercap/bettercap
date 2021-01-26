// +build !windows

package ble

import (
	"sort"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
)

var (
	bleAliveInterval = time.Duration(5) * time.Second
)

func (mod *BLERecon) getRow(dev *network.BLEDevice, withName bool) []string {
	rssi := network.ColorRSSI(dev.RSSI)
	address := network.NormalizeMac(dev.Device.ID())
	vendor := tui.Dim(ops.Ternary(dev.Vendor == "", dev.Advertisement.Company, dev.Vendor).(string))
	isConnectable := ops.Ternary(dev.Advertisement.Connectable, tui.Green("✔"), tui.Red("✖")).(string)
	sinceSeen := time.Since(dev.LastSeen)
	lastSeen := dev.LastSeen.Format("15:04:05")

	blePresentInterval := time.Duration(mod.devTTL) * time.Second
	if sinceSeen <= bleAliveInterval {
		lastSeen = tui.Bold(lastSeen)
	} else if sinceSeen > blePresentInterval {
		lastSeen = tui.Dim(lastSeen)
		address = tui.Dim(address)
	}

	if withName {
		return []string{
			rssi,
			address,
			tui.Yellow(dev.Name()),
			vendor,
			dev.Advertisement.Flags.String(),
			isConnectable,
			lastSeen,
		}
	} else {
		return []string{
			rssi,
			address,
			vendor,
			dev.Advertisement.Flags.String(),
			isConnectable,
			lastSeen,
		}
	}
}

func (mod *BLERecon) doFilter(dev *network.BLEDevice) bool {
	if mod.selector.Expression == nil {
		return true
	}
	return mod.selector.Expression.MatchString(dev.Device.ID()) ||
		mod.selector.Expression.MatchString(dev.Device.Name()) ||
		mod.selector.Expression.MatchString(dev.Vendor)
}

func (mod *BLERecon) doSelection() (devices []*network.BLEDevice, err error) {
	if err = mod.selector.Update(); err != nil {
		return
	}

	devices = mod.Session.BLE.Devices()
	filtered := []*network.BLEDevice{}
	for _, dev := range devices {
		if mod.doFilter(dev) {
			filtered = append(filtered, dev)
		}
	}
	devices = filtered

	switch mod.selector.SortField {
	case "mac":
		sort.Sort(ByBLEMacSorter(devices))
	case "seen":
		sort.Sort(ByBLESeenSorter(devices))
	default:
		sort.Sort(ByBLERSSISorter(devices))
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

func (mod *BLERecon) colNames(withName bool) []string {
	colNames := []string{"RSSI", "MAC", "Vendor", "Flags", "Connect", "Seen"}
	seenIdx := 5
	if withName {
		colNames = []string{"RSSI", "MAC", "Name", "Vendor", "Flags", "Connect", "Seen"}
		seenIdx = 6
	}
	switch mod.selector.SortField {
	case "rssi":
		colNames[0] += " " + mod.selector.SortSymbol
	case "mac":
		colNames[1] += " " + mod.selector.SortSymbol
	case "seen":
		colNames[seenIdx] += " " + mod.selector.SortSymbol
	}
	return colNames
}

func (mod *BLERecon) Show() error {
	devices, err := mod.doSelection()
	if err != nil {
		return err
	}

	hasName := false
	for _, dev := range devices {
		if dev.Name() != "" {
			hasName = true
			break
		}
	}

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, mod.getRow(dev, hasName))
	}

	if len(rows) > 0 {
		tui.Table(mod.Session.Events.Stdout, mod.colNames(hasName), rows)
		mod.Session.Refresh()
	}

	return nil
}
