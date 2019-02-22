// +build !windows
// +build !darwin

package ble

import (
	"github.com/bettercap/gatt"
)

func (mod *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	mod.Debug("state changed to %v", s)

	switch s {
	case gatt.StatePoweredOn:
		if mod.currDevice == nil {
			mod.Info("starting discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		}
	case gatt.StatePoweredOff:
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
	mod.setCurrentDevice(nil)
	if mod.Running() {
		mod.Info("device disconnected, restoring discovery.")
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
		mod.Info("disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
	}(p)

	mod.Session.Events.Add("ble.device.connected", mod.currDevice)

	if err := p.SetMTU(500); err != nil {
		mod.Warning("failed to set MTU: %s", err)
	}

	mod.Info("connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		mod.Error("error discovering services: %s", err)
		return
	}

	mod.showServices(p, services)
}
