// +build !windows

package ble

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/gatt"

	"github.com/evilsocket/islazy/tui"
)

var appearances = map[uint16]string{
	0:    "Unknown",
	64:   "Generic Phone",
	128:  "Generic Computer",
	192:  "Generic Watch",
	193:  "Watch: Sports Watch",
	256:  "Generic Clock",
	320:  "Generic Display",
	384:  "Generic Remote Control",
	448:  "Generic Eye-glasses",
	512:  "Generic Tag",
	576:  "Generic Keyring",
	640:  "Generic Media Player",
	704:  "Generic Barcode Scanner",
	768:  "Generic Thermometer",
	769:  "Thermometer: Ear",
	832:  "Generic Heart rate Sensor",
	833:  "Heart Rate Sensor: Heart Rate Belt",
	896:  "Generic Blood Pressure",
	897:  "Blood Pressure: Arm",
	898:  "Blood Pressure: Wrist",
	960:  "Human Interface Device (HID)",
	961:  "Keyboard",
	962:  "Mouse",
	963:  "Joystick",
	964:  "Gamepad",
	965:  "Digitizer Tablet",
	966:  "Card Reader",
	967:  "Digital Pen",
	968:  "Barcode Scanner",
	1024: "Generic Glucose Meter",
	1088: "Generic: Running Walking Sensor",
	1089: "Running Walking Sensor: In-Shoe",
	1090: "Running Walking Sensor: On-Shoe",
	1091: "Running Walking Sensor: On-Hip",
	1152: "Generic: Cycling",
	1153: "Cycling: Cycling Computer",
	1154: "Cycling: Speed Sensor",
	1155: "Cycling: Cadence Sensor",
	1156: "Cycling: Power Sensor",
	1157: "Cycling: Speed and Cadence Sensor",
	1216: "Generic Control Device",
	1217: "Switch",
	1218: "Multi-switch",
	1219: "Button",
	1220: "Slider",
	1221: "Rotary",
	1222: "Touch-panel",
	1280: "Generic Network Device",
	1281: "Access Point",
	1344: "Generic Sensor",
	1345: "Motion Sensor",
	1346: "Air Quality Sensor",
	1347: "Temperature Sensor",
	1348: "Humidity Sensor",
	1349: "Leak Sensor",
	1350: "Smoke Sensor",
	1351: "Occupancy Sensor",
	1352: "Contact Sensor",
	1353: "Carbon Monoxide Sensor",
	1354: "Carbon Dioxide Sensor",
	1355: "Ambient Light Sensor",
	1356: "Energy Sensor",
	1357: "Color Light Sensor",
	1358: "Rain Sensor",
	1359: "Fire Sensor",
	1360: "Wind Sensor",
	1361: "Proximity Sensor",
	1362: "Multi-Sensor",
	1408: "Generic Light Fixtures",
	1409: "Wall Light",
	1410: "Ceiling Light",
	1411: "Floor Light",
	1412: "Cabinet Light",
	1413: "Desk Light",
	1414: "Troffer Light",
	1415: "Pendant Light",
	1416: "In-ground Light",
	1417: "Flood Light",
	1418: "Underwater Light",
	1419: "Bollard with Light",
	1420: "Pathway Light",
	1421: "Garden Light",
	1422: "Pole-top Light",
	1423: "Spotlight",
	1424: "Linear Light",
	1425: "Street Light",
	1426: "Shelves Light",
	1427: "High-bay / Low-bay Light",
	1428: "Emergency Exit Light",
	1472: "Generic Fan",
	1473: "Ceiling Fan",
	1474: "Axial Fan",
	1475: "Exhaust Fan",
	1476: "Pedestal Fan",
	1477: "Desk Fan",
	1478: "Wall Fan",
	1536: "Generic HVAC",
	1537: "Thermostat",
	1600: "Generic Air Conditioning",
	1664: "Generic Humidifier",
	1728: "Generic Heating",
	1729: "Radiator",
	1730: "Boiler",
	1731: "Heat Pump",
	1732: "Infrared Heater",
	1733: "Radiant Panel Heater",
	1734: "Fan Heater",
	1735: "Air Curtain",
	1792: "Generic Access Control",
	1793: "Access Door",
	1794: "Garage Door",
	1795: "Emergency Exit Door",
	1796: "Access Lock",
	1797: "Elevator",
	1798: "Window",
	1799: "Entrance Gate",
	1856: "Generic Motorized Device",
	1857: "Motorized Gate",
	1858: "Awning",
	1859: "Blinds or Shades",
	1860: "Curtains",
	1861: "Screen",
	1920: "Generic Power Device",
	1921: "Power Outlet",
	1922: "Power Strip",
	1923: "Plug",
	1924: "Power Supply",
	1925: "LED Driver",
	1926: "Fluorescent Lamp Gear",
	1927: "HID Lamp Gear",
	1984: "Generic Light Source",
	1985: "Incandescent Light Bulb",
	1986: "LED Bulb",
	1987: "HID Lamp",
	1988: "Fluorescent Lamp",
	1989: "LED Array",
	1990: "Multi-Color LED Array",
	3136: "Generic: Pulse Oximeter",
	3137: "Fingertip",
	3138: "Wrist Worn",
	3200: "Generic: Weight Scale",
	3264: "Generic",
	3265: "Powered Wheelchair",
	3266: "Mobility Scooter",
	3328: "Generic",
	5184: "Generic: Outdoor Sports Activity",
	5185: "Location Display Device",
	5186: "Location and Navigation Display Device",
	5187: "Location Pod",
	5188: "Location and Navigation Pod",
}

func parseProperties(ch *gatt.Characteristic) (props []string, isReadable bool, isWritable bool, withResponse bool) {
	isReadable = false
	isWritable = false
	withResponse = false
	props = make([]string, 0)
	mask := ch.Properties()

	if (mask & gatt.CharBroadcast) != 0 {
		props = append(props, "BCAST")
	}
	if (mask & gatt.CharRead) != 0 {
		isReadable = true
		props = append(props, "READ")
	}
	if (mask&gatt.CharWriteNR) != 0 || (mask&gatt.CharWrite) != 0 {
		props = append(props, tui.Bold("WRITE"))
		isWritable = true
		withResponse = (mask & gatt.CharWriteNR) == 0
	}
	if (mask & gatt.CharNotify) != 0 {
		props = append(props, "NOTIFY")
	}
	if (mask & gatt.CharIndicate) != 0 {
		props = append(props, "INDICATE")
	}
	if (mask & gatt.CharSignedWrite) != 0 {
		props = append(props, tui.Yellow("SIGN WRITE"))
		isWritable = true
		withResponse = true
	}
	if (mask & gatt.CharExtended) != 0 {
		props = append(props, "X")
	}

	return
}

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

// org.bluetooth.characteristic.gap.appearance
func parseAppearance(raw []byte) string {
	app := binary.LittleEndian.Uint16(raw[0:2])
	if appName, found := appearances[app]; found {
		return tui.Green(appName)
	}
	return fmt.Sprintf("0x%x", app)
}

// org.bluetooth.characteristic.pnp_id
func parsePNPID(raw []byte) []string {
	vendorIdSrc := byte(raw[0])
	vendorId := binary.LittleEndian.Uint16(raw[1:3])
	prodId := binary.LittleEndian.Uint16(raw[3:5])
	prodVer := binary.LittleEndian.Uint16(raw[5:7])

	src := ""
	if vendorIdSrc == 1 {
		src = " (Bluetooth SIG assigned Company Identifier)"
	} else if vendorIdSrc == 2 {
		src = " (USB Implementer’s Forum assigned Vendor ID value)"
	}

	return []string{
		tui.Green("Vendor ID") + fmt.Sprintf(": 0x%04x%s", vendorId, tui.Dim(src)),
		tui.Green("Product ID") + fmt.Sprintf(": 0x%04x", prodId),
		tui.Green("Product Version") + fmt.Sprintf(": 0x%04x", prodVer),
	}
}

// org.bluetooth.characteristic.gap.peripheral_preferred_connection_parameters
func parseConnectionParams(raw []byte) []string {
	minConInt := binary.LittleEndian.Uint16(raw[0:2])
	maxConInt := binary.LittleEndian.Uint16(raw[2:4])
	slaveLat := binary.LittleEndian.Uint16(raw[4:6])
	conTimeMul := binary.LittleEndian.Uint16(raw[6:8])

	return []string{
		tui.Green("Connection Interval") + fmt.Sprintf(": %d -> %d", minConInt, maxConInt),
		tui.Green("Slave Latency") + fmt.Sprintf(": %d", slaveLat),
		tui.Green("Connection Supervision Timeout Multiplier") + fmt.Sprintf(": %d", conTimeMul),
	}
}

// org.bluetooth.characteristic.gap.peripheral_privacy_flag
func parsePrivacyFlag(raw []byte) string {
	if raw[0] == 0x0 {
		return tui.Green("Privacy Disabled")
	}
	return tui.Red("Privacy Enabled")
}

func (mod *BLERecon) showServices(p gatt.Peripheral, services []*gatt.Service) {
	columns := []string{"Handles", "Service > Characteristics", "Properties", "Data"}
	rows := make([][]string, 0)

	wantsToWrite := mod.writeUUID != nil
	foundToWrite := false

	mod.currDevice.Services = make([]network.BLEService, 0)

	for _, svc := range services {
		service := network.BLEService{
			UUID:            svc.UUID().String(),
			Name:            svc.Name(),
			Handle:          svc.Handle(),
			EndHandle:       svc.EndHandle(),
			Characteristics: make([]network.BLECharacteristic, 0),
		}

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
		} else {
			for _, ch := range chars {
				props, isReadable, isWritable, withResponse := parseProperties(ch)

				char := network.BLECharacteristic{
					UUID:       ch.UUID().String(),
					Name:       ch.Name(),
					Handle:     ch.VHandle(),
					Properties: props,
				}

				mod.Session.Events.Add("ble.device.characteristic.discovered", ch)

				name = ch.Name()
				if name == "" {
					name = "    " + ch.UUID().String()
				} else {
					name = fmt.Sprintf("    %s (%s)", tui.Green(name), tui.Dim(ch.UUID().String()))
				}

				if wantsToWrite && mod.writeUUID.Equal(ch.UUID()) {
					foundToWrite = true
					if isWritable {
						mod.Debug("writing %d bytes to characteristics %s ...", len(mod.writeData), mod.writeUUID)
					} else {
						mod.Warning("attempt to write %d bytes to non writable characteristics %s ...", len(mod.writeData), mod.writeUUID)
					}

					if err := p.WriteCharacteristic(ch, mod.writeData, !withResponse); err != nil {
						mod.Error("error while writing: %s", err)
					}
				}

				sz := 0
				raw := ([]byte)(nil)
				err := error(nil)
				if isReadable {
					if raw, err = p.ReadCharacteristic(ch); raw != nil {
						sz = len(raw)
					}
				}

				data := ""
				multi := ([]string)(nil)
				if err != nil {
					data = tui.Red(err.Error())
				} else if ch.Name() == "Appearance" && sz >= 2 {
					data = parseAppearance(raw)
				} else if ch.Name() == "PnP ID" && sz >= 7 {
					multi = parsePNPID(raw)
				} else if ch.Name() == "Peripheral Preferred Connection Parameters" && sz >= 8 {
					multi = parseConnectionParams(raw)
				} else if ch.Name() == "Peripheral Privacy Flag" && sz >= 1 {
					data = parsePrivacyFlag(raw)
				} else {
					data = parseRawData(raw)
				}

				if ch.Name() == "Device Name" && data != "" && mod.currDevice.DeviceName == "" {
					mod.currDevice.DeviceName = data
				}

				if multi == nil {
					char.Data = data
					rows = append(rows, []string{
						fmt.Sprintf("%04x", ch.VHandle()),
						name,
						strings.Join(props, ", "),
						data,
					})
				} else {
					char.Data = multi
					for i, m := range multi {
						if i == 0 {
							rows = append(rows, []string{
								fmt.Sprintf("%04x", ch.VHandle()),
								name,
								strings.Join(props, ", "),
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
			rows = append(rows, []string{"", "", "", ""})
		}

		mod.currDevice.Services = append(mod.currDevice.Services, service)
	}

	if wantsToWrite && !foundToWrite {
		mod.Error("writable characteristics %s not found.", mod.writeUUID)
	} else {
		tui.Table(mod.Session.Events.Stdout, columns, rows)
		mod.Session.Refresh()
	}
}
