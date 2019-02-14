// +build !windows
// +build !darwin

package ble

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/bettercap/gatt"

	"github.com/evilsocket/islazy/tui"
)

var (
	bleAliveInterval   = time.Duration(5) * time.Second
	blePresentInterval = time.Duration(30) * time.Second
)

func (mod *BLERecon) getRow(dev *network.BLEDevice) []string {
	// ref. https://www.metageek.com/training/resources/understanding-rssi-2.html
	rssi := fmt.Sprintf("%d dBm", dev.RSSI)
	if dev.RSSI >= -67 {
		rssi = tui.Green(rssi)
	} else if dev.RSSI >= -70 {
		rssi = tui.Dim(tui.Green(rssi))
	} else if dev.RSSI >= -80 {
		rssi = tui.Yellow(rssi)
	} else {
		rssi = tui.Dim(tui.Red(rssi))
	}

	address := network.NormalizeMac(dev.Device.ID())
	vendor := tui.Dim(dev.Vendor)
	sinceSeen := time.Since(dev.LastSeen)
	lastSeen := dev.LastSeen.Format("15:04:05")

	if sinceSeen <= bleAliveInterval {
		lastSeen = tui.Bold(lastSeen)
	} else if sinceSeen > blePresentInterval {
		lastSeen = tui.Dim(lastSeen)
		address = tui.Dim(address)
	}

	isConnectable := tui.Red("✖")
	if dev.Advertisement.Connectable {
		isConnectable = tui.Green("✔")
	}

	return []string{
		rssi,
		address,
		dev.Device.Name(),
		vendor,
		isConnectable,
		lastSeen,
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

func (mod *BLERecon) doSelection() (err error, devices []*network.BLEDevice) {
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

func (mod *BLERecon) colNames() []string {
	colNames := []string{"RSSI", "MAC", "Name", "Vendor", "Connectable", "Seen"}
	switch mod.selector.SortField {
	case "rssi":
		colNames[0] += " " + mod.selector.SortSymbol
	case "mac":
		colNames[1] += " " + mod.selector.SortSymbol
	case "seen":
		colNames[5] += " " + mod.selector.SortSymbol
	}
	return colNames
}

func (mod *BLERecon) Show() error {
	err, devices := mod.doSelection()
	if err != nil {
		return err
	}

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, mod.getRow(dev))
	}

	if len(rows) > 0 {
		tui.Table(os.Stdout, mod.colNames(), rows)
		mod.Session.Refresh()
	}

	return nil
}

func parseProperties(ch *gatt.Characteristic) (props []string, isReadable bool, isWritable bool, withResponse bool) {
	isReadable = false
	isWritable = false
	withResponse = false
	props = make([]string, 0)
	mask := ch.Properties()

	if (mask & gatt.CharBroadcast) != 0 {
		props = append(props, "bcast")
	}
	if (mask & gatt.CharRead) != 0 {
		isReadable = true
		props = append(props, "read")
	}
	if (mask&gatt.CharWriteNR) != 0 || (mask&gatt.CharWrite) != 0 {
		props = append(props, tui.Bold("write"))
		isWritable = true
		withResponse = (mask & gatt.CharWriteNR) == 0
	}
	if (mask & gatt.CharNotify) != 0 {
		props = append(props, "notify")
	}
	if (mask & gatt.CharIndicate) != 0 {
		props = append(props, "indicate")
	}
	if (mask & gatt.CharSignedWrite) != 0 {
		props = append(props, tui.Yellow("*write"))
		isWritable = true
		withResponse = true
	}
	if (mask & gatt.CharExtended) != 0 {
		props = append(props, "x")
	}

	return
}

func parseRawData(raw []byte) string {
	s := ""
	for _, b := range raw {
		if b != 00 && !strconv.IsPrint(rune(b)) {
			return fmt.Sprintf("%x", raw)
		} else if b == 0 {
			break
		} else {
			s += fmt.Sprintf("%c", b)
		}
	}

	return tui.Yellow(s)
}

func (mod *BLERecon) showServices(p gatt.Peripheral, services []*gatt.Service) {
	columns := []string{"Handles", "Service > Characteristics", "Properties", "Data"}
	rows := make([][]string, 0)

	wantsToWrite := mod.writeUUID != nil
	foundToWrite := false

	for _, svc := range services {
		mod.Session.Events.Add("ble.device.service.discovered", svc)

		name := svc.Name()
		if name == "" {
			name = svc.UUID().String()
		} else {
			name = fmt.Sprintf("%s (%s)", tui.Green(name), tui.Dim(svc.UUID().String()))
		}

		row := []string{
			fmt.Sprintf("%04x -> %04x", svc.Handle(), svc.EndHandle()),
			name,
			"",
			"",
		}

		rows = append(rows, row)

		chars, err := p.DiscoverCharacteristics(nil, svc)
		if err != nil {
			mod.Error("error while enumerating chars for service %s: %s", svc.UUID(), err)
			continue
		}

		for _, ch := range chars {
			mod.Session.Events.Add("ble.device.characteristic.discovered", ch)

			name = ch.Name()
			if name == "" {
				name = "    " + ch.UUID().String()
			} else {
				name = fmt.Sprintf("    %s (%s)", tui.Green(name), tui.Dim(ch.UUID().String()))
			}

			props, isReadable, isWritable, withResponse := parseProperties(ch)

			if wantsToWrite && mod.writeUUID.Equal(ch.UUID()) {
				foundToWrite = true
				if isWritable {
					mod.Info("writing %d bytes to characteristics %s ...", len(mod.writeData), mod.writeUUID)
				} else {
					mod.Warning("attempt to write %d bytes to non writable characteristics %s ...", len(mod.writeData), mod.writeUUID)
				}

				err := p.WriteCharacteristic(ch, mod.writeData, !withResponse)
				if err != nil {
					mod.Error("error while writing: %s", err)
				}
			}

			data := ""
			if isReadable {
				raw, err := p.ReadCharacteristic(ch)
				if err != nil {
					data = tui.Red(err.Error())
				} else {
					data = parseRawData(raw)
				}
			}

			row := []string{
				fmt.Sprintf("%04x", ch.Handle()),
				name,
				strings.Join(props, ", "),
				data,
			}

			rows = append(rows, row)
		}
	}

	if wantsToWrite && !foundToWrite {
		mod.Error("writable characteristics %s not found.", mod.writeUUID)
	} else {
		tui.Table(os.Stdout, columns, rows)
		mod.Session.Refresh()
	}
}
