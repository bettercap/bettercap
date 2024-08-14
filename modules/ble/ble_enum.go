package ble

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/evilsocket/islazy/tui"
	"tinygo.org/x/bluetooth"
)

func parseRawData(raw []byte) string {
	s := ""
	for _, b := range raw {
		if strconv.IsPrint(rune(b)) {
			s += tui.Yellow(string(b))
		} else {
			s += tui.Dim(fmt.Sprintf("%02x", b))
		}
	}
	return s
}

func (mod *BLERecon) startEnumeration(address string) error {
	knownDev, found := mod.Session.BLE.Get(address)
	if !found || knownDev == nil {
		return fmt.Errorf("device with address %s not found", address)
	} else if err := mod.adapter.StopScan(); err != nil {
		mod.Warning("error stopping previous scan: %v", err)
	}

	mod.current.ResetWrite()

	defer func() {
		// make sure to restart scan when we're done
		mod.Info("restoring ble scan")
		if err := mod.adapter.Scan(mod.onDevice); err != nil {
			mod.Error("error restoring ble scan: %v", err)
		}
		mod.Info("ble scan restored")
	}()

	mod.Info("connecting to %s ...", address)

	addr := bluetooth.Address{}
	addr.Set(address)
	tm := bluetooth.NewDuration(time.Second * time.Duration(mod.connTimeout))

	device, err := mod.adapter.Connect(addr, bluetooth.ConnectionParams{ConnectionTimeout: tm})
	if err != nil {
		mod.Error("error connecting to %s: %v", address, err)
		return err
	}

	mod.Session.Events.Add("ble.device.connected", knownDev)

	mod.Info("connected to %s", address)

	srvcs, err := device.DiscoverServices(nil)
	if err != nil {
		return fmt.Errorf("could not discover services for %s: %v", address, err)
	}

	columns := []string{"Service > Characteristics", "Properties", "Data"}
	rows := make([][]string, 0)

	knownDev.ResetServices()

	for _, svc := range srvcs {
		service := network.NewBLEService(strings.ToLower(svc.UUID().String()))

		mod.Session.Events.Add("ble.device.service.discovered", svc)

		name := ""
		if service.Name == "" {
			name = svc.UUID().String()
		} else {
			name = fmt.Sprintf("%s (%s)", tui.Green(service.Name), tui.Dim(svc.UUID().String()))
		}

		row := []string{
			name,
			"",
			"",
		}

		rows = append(rows, row)

		chars, err := svc.DiscoverCharacteristics(nil)
		if err != nil {
			mod.Error("error while enumerating chars for service %s: %s", svc.UUID(), err)
		} else {
			for _, ch := range chars {
				char := network.NewBLECharacteristic(strings.ToLower(ch.UUID().String()))
				if mtu, err := ch.GetMTU(); err != nil {
					mod.Warning("can't read %v mtu: %v", ch.UUID(), err)
				} else {
					char.MTU = mtu
				}

				if props, err := ch.Properties(); err != nil {

				} else {

				}

				mod.Session.Events.Add("ble.device.characteristic.discovered", ch)

				name := char.Name
				if name == "" {
					name = "    " + ch.UUID().String()
				} else {
					name = fmt.Sprintf("    %s (%s)", tui.Green(name), tui.Dim(ch.UUID().String()))
				}

				data := ""
				multi := ([]string)(nil)
				raw := make([]byte, 255)
				if n, err := ch.Read(raw); err == nil {
					raw = raw[0:n]
					char.Properties = []string{"READ"}
					data = parseRawData(raw)
				}

				if multi == nil {
					char.Data = data
					rows = append(rows, []string{
						name,
						strings.Join(char.Properties, ", "),
						data,
					})
				} else {
					char.Data = multi
					for i, m := range multi {
						if i == 0 {
							rows = append(rows, []string{
								name,
								strings.Join(char.Properties, ", "),
								m,
							})
						} else {
							rows = append(rows, []string{"", "", "", m})
						}
					}
				}

				service.Characteristics = append(service.Characteristics, char)
			}
			// blank row after every service, bleah style
			rows = append(rows, []string{"", "", ""})
		}

		knownDev.AddService(service)
	}

	tui.Table(mod.Session.Events.Stdout, columns, rows)

	if err := device.Disconnect(); err != nil {
		mod.Warning("error disconnecting from %s: %v", address, err)
	} else {
		mod.Info("disconnected from %s", address)
	}

	return nil
}
