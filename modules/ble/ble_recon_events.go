// +build !windows
// +build !darwin

package ble

import (
	"github.com/bettercap/bettercap/log"

	"github.com/bettercap/gatt"
)

func (d *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	log.Info("BLE state changed to %v", s)

	switch s {
	case gatt.StatePoweredOn:
		if d.currDevice == nil {
			log.Info("Starting BLE discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		}
	case gatt.StatePoweredOff:
		d.gattDevice = nil

	default:
		log.Warning("Unexpected BLE state: %v", s)
	}
}

func (d *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	d.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
}

func (d *BLERecon) onPeriphDisconnected(p gatt.Peripheral, err error) {
	if d.Running() {
		// restore scanning
		log.Info("Device disconnected, restoring BLE discovery.")
		d.setCurrentDevice(nil)
		d.gattDevice.Scan([]gatt.UUID{}, true)
	}
}

func (d *BLERecon) onPeriphConnected(p gatt.Peripheral, err error) {
	if err != nil {
		log.Warning("Connected to %s but with error: %s", p.ID(), err)
		return
	} else if d.currDevice == nil {
		// timed out
		log.Warning("Connected to %s but after the timeout :(", p.ID())
		return
	}

	d.connected = true

	defer func(per gatt.Peripheral) {
		log.Info("Disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
	}(p)

	d.Session.Events.Add("ble.device.connected", d.currDevice)

	if err := p.SetMTU(500); err != nil {
		log.Warning("Failed to set MTU: %s", err)
	}

	log.Info("Connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Error("Error discovering services: %s", err)
		return
	}

	d.showServices(p, services)
}
