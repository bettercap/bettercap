// +build !windows

package ble

import (
	"github.com/bettercap/gatt"
)

func (mod *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	mod.Debug("state changed to %v", s)

	switch s {
	case gatt.StatePoweredOn:
		if mod.currDevice == nil {
			mod.Debug("starting discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		} else {
			mod.Debug("current device was not cleaned: %v", mod.currDevice)
		}
	case gatt.StatePoweredOff:
		mod.Debug("resetting device instance")
		mod.gattDevice.StopScanning()
		mod.setCurrentDevice(nil)
		mod.gattDevice = nil

	default:
		mod.Warning("unexpected state: %v", s)
	}
}

func (mod *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	mod.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
}

func (mod *BLERecon) onPeriphDisconnected(p gatt.Peripheral, err error) {
	mod.Session.Events.Add("ble.device.disconnected", mod.currDevice)
	mod.setCurrentDevice(nil)
	if mod.Running() {
		mod.Debug("device disconnected, restoring discovery.")
		mod.gattDevice.Scan([]gatt.UUID{}, true)
	}
}

func (mod *BLERecon) onPeriphConnected(p gatt.Peripheral, err error) {
	if err != nil {
		mod.Warning("connected to %s but with error: %s", p.ID(), err)
		return
	} else if mod.currDevice == nil {
		mod.Warning("connected to %s but after the timeout :(", p.ID())
		return
	}

	mod.connected = true

	defer func(per gatt.Peripheral) {
		mod.Debug("disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
		mod.setCurrentDevice(nil)
	}(p)

	mod.Session.Events.Add("ble.device.connected", mod.currDevice)

	if err := p.SetMTU(500); err != nil {
		mod.Warning("failed to set MTU: %s", err)
	}

	mod.Debug("connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	// https://github.com/bettercap/bettercap/issues/498
	if err != nil && err.Error() != "success" {
		mod.Error("error discovering services: %s", err)
		return
	}

	mod.showServices(p, services)
}
