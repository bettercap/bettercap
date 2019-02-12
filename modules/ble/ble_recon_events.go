// +build !windows
// +build !darwin

package ble

import (
	"github.com/bettercap/gatt"
)

func (d *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	d.Info("BLE state changed to %v", s)

	switch s {
	case gatt.StatePoweredOn:
		if d.currDevice == nil {
			d.Info("Starting BLE discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		}
	case gatt.StatePoweredOff:
		d.gattDevice = nil

	default:
		d.Warning("Unexpected BLE state: %v", s)
	}
}

func (d *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	d.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
}

func (d *BLERecon) onPeriphDisconnected(p gatt.Peripheral, err error) {
	if d.Running() {
		// restore scanning
		d.Info("Device disconnected, restoring BLE discovery.")
		d.setCurrentDevice(nil)
		d.gattDevice.Scan([]gatt.UUID{}, true)
	}
}

func (d *BLERecon) onPeriphConnected(p gatt.Peripheral, err error) {
	if err != nil {
		d.Warning("Connected to %s but with error: %s", p.ID(), err)
		return
	} else if d.currDevice == nil {
		// timed out
		d.Warning("Connected to %s but after the timeout :(", p.ID())
		return
	}

	d.connected = true

	defer func(per gatt.Peripheral) {
		d.Info("Disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
	}(p)

	d.Session.Events.Add("ble.device.connected", d.currDevice)

	if err := p.SetMTU(500); err != nil {
		d.Warning("Failed to set MTU: %s", err)
	}

	d.Info("Connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		d.Error("Error discovering services: %s", err)
		return
	}

	d.showServices(p, services)
}
