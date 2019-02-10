package ble

import (
	"github.com/bettercap/gatt"
	// "github.com/bettercap/gatt/linux/cmd"
)

var defaultBLEClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(255),
	gatt.LnxDeviceID(-1, true),
}

/*

var defaultBLEServerOptions = []gatt.Option{
	gatt.LnxMaxConnections(255),
	gatt.LnxDeviceID(-1, true),
	gatt.LnxSetAdvertisingParameters(&cmd.LESetAdvertisingParameters{
		AdvertisingIntervalMin: 0x00f4,
		AdvertisingIntervalMax: 0x00f4,
		AdvertisingChannelMap:  0x7,
	}),
}

*/
