// +build !windows
// +build !darwin

package modules

import (
	"fmt"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
)

func (s *EventsStream) viewBLEEvent(e session.Event) {
	if e.Tag == "ble.device.new" {
		dev := e.Data.(*network.BLEDevice)
		name := dev.Device.Name()
		if name != "" {
			name = " " + core.Bold(name)
		}
		vend := dev.Vendor
		if vend != "" {
			vend = fmt.Sprintf(" (%s)", core.Yellow(vend))
		}

		fmt.Fprintf(s.output, "[%s] [%s] New BLE device%s detected as %s%s %s.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			name,
			dev.Device.ID(),
			vend,
			core.Dim(fmt.Sprintf("%d dBm", dev.RSSI)))
	} else if e.Tag == "ble.device.lost" {
		dev := e.Data.(*network.BLEDevice)
		name := dev.Device.Name()
		if name != "" {
			name = " " + core.Bold(name)
		}
		vend := dev.Vendor
		if vend != "" {
			vend = fmt.Sprintf(" (%s)", core.Yellow(vend))
		}

		fmt.Fprintf(s.output, "[%s] [%s] BLE device%s %s%s lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			name,
			dev.Device.ID(),
			vend)
	} /* else {
		fmt.Fprintf(s.output,"[%s] [%s]\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag))
	} */
}
