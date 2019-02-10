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

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"

	"github.com/bettercap/gatt"

	"github.com/evilsocket/islazy/tui"
)

var (
	bleAliveInterval   = time.Duration(5) * time.Second
	blePresentInterval = time.Duration(30) * time.Second
)

func (d *BLERecon) getRow(dev *network.BLEDevice) []string {
	address := network.NormalizeMac(dev.Device.ID())
	vendor := dev.Vendor
	sinceSeen := time.Since(dev.LastSeen)
	lastSeen := dev.LastSeen.Format("15:04:05")

	if sinceSeen <= bleAliveInterval {
		lastSeen = tui.Bold(lastSeen)
	} else if sinceSeen > blePresentInterval {
		lastSeen = tui.Dim(lastSeen)
		address = tui.Dim(address)
	}

	isConnectable := tui.Red("no")
	if dev.Advertisement.Connectable {
		isConnectable = tui.Green("yes")
	}

	return []string{
		fmt.Sprintf("%d dBm", dev.RSSI),
		address,
		dev.Device.Name(),
		vendor,
		isConnectable,
		lastSeen,
	}
}

func (d *BLERecon) Show() error {
	devices := d.Session.BLE.Devices()

	sort.Sort(ByBLERSSISorter(devices))

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, d.getRow(dev))
	}
	nrows := len(rows)

	columns := []string{"RSSI", "Address", "Name", "Vendor", "Connectable", "Last Seen"}

	if nrows > 0 {
		tui.Table(os.Stdout, columns, rows)
	}

	d.Session.Refresh()
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

func (d *BLERecon) showServices(p gatt.Peripheral, services []*gatt.Service) {
	columns := []string{"Handles", "Service > Characteristics", "Properties", "Data"}
	rows := make([][]string, 0)

	wantsToWrite := d.writeUUID != nil
	foundToWrite := false

	for _, svc := range services {
		d.Session.Events.Add("ble.device.service.discovered", svc)

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
			log.Error("Error while enumerating chars for service %s: %s", svc.UUID(), err)
			continue
		}

		for _, ch := range chars {
			d.Session.Events.Add("ble.device.characteristic.discovered", ch)

			name = ch.Name()
			if name == "" {
				name = "    " + ch.UUID().String()
			} else {
				name = fmt.Sprintf("    %s (%s)", tui.Green(name), tui.Dim(ch.UUID().String()))
			}

			props, isReadable, isWritable, withResponse := parseProperties(ch)

			if wantsToWrite && d.writeUUID.Equal(ch.UUID()) {
				foundToWrite = true
				if isWritable {
					log.Info("Writing %d bytes to characteristics %s ...", len(d.writeData), d.writeUUID)
				} else {
					log.Warning("Attempt to write %d bytes to non writable characteristics %s ...", len(d.writeData), d.writeUUID)
				}

				err := p.WriteCharacteristic(ch, d.writeData, !withResponse)
				if err != nil {
					log.Error("Error while writing: %s", err)
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
		log.Error("Writable characteristics %s not found.", d.writeUUID)
	} else {
		tui.Table(os.Stdout, columns, rows)
		d.Session.Refresh()
	}
}
