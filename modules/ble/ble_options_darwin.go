package ble

import (
	"github.com/bettercap/gatt"
)

func getClientOptions(deviceID int) []gatt.Option {
	return []gatt.Option{
		gatt.MacDeviceRole(gatt.CentralManager),
	}
}

/*

var defaultBLEServerOptions = []gatt.Option{
	gatt.MacDeviceRole(gatt.PeripheralManager),
}

*/
