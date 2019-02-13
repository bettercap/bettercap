// +build !windows
// +build !darwin

package ble

import (
	"github.com/bettercap/gatt"
)

func (mod *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	mod.Info("BLE state changed to %v", s)

	switch s {
	case gatt.StatePoweredOn:
		if mod.currDevice == nil {
			mod.Info("Starting BLE discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		}
	case gatt.StatePoweredOff:
		mod.gattDevice = nil

	default:
		mod.Warning("Unexpected BLE state: %v", s)
	}
}

func (mod *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	mod.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
}

func (mod *BLERecon) onPeriphDisconnected(p gatt.Peripheral, err error) {
	if mod.Running() {
		// restore scanning
		mod.Info("Device disconnected, restoring BLE discovery.")
		mod.setCurrentDevice(nil)
		mod.gattDevice.Scan([]gatt.UUID{}, true)
	}
}

func (mod *BLERecon) onPeriphConnected(p gatt.Peripheral, err error) {
	if err != nil {
		mod.Warning("Connected to %s but with error: %s", p.ID(), err)
		return
	} else if mod.currDevice == nil {
		// timed out
		mod.Warning("Connected to %s but after the timeout :(", p.ID())
		return
	}

	mod.connected = true

	defer func(per gatt.Peripheral) {
		mod.Info("Disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
	}(p)

	mod.Session.Events.Add("ble.device.connected", mod.currDevice)

	if err := p.SetMTU(500); err != nil {
		mod.Warning("Failed to set MTU: %s", err)
	}

	mod.Info("Connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		mod.Error("Error discovering services: %s", err)
		return
	}

	mod.showServices(p, services)
}
