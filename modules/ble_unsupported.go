// +build windows darwin

package modules

import (
	"github.com/bettercap/bettercap/session"
)

type BLERecon struct {
	session.SessionModule
}

/*
// darwin

var defaultBLEClientOptions = []gatt.Option{
	gatt.MacDeviceRole(gatt.CentralManager),
}

var defaultBLEServerOptions = []gatt.Option{
	gatt.MacDeviceRole(gatt.PeripheralManager),
}
*/
func NewBLERecon(s *session.Session) *BLERecon {
	d := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
	}

	d.AddHandler(session.NewModuleHandler("ble.recon on", "",
		"Start Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	d.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	return d
}

func (d BLERecon) Name() string {
	return "ble.recon"
}

func (d BLERecon) Description() string {
	return "Bluetooth Low Energy devices discovery."
}

func (d BLERecon) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (d *BLERecon) Configure() (err error) {
	return session.ErrNotSupported
}

func (d *BLERecon) Start() error {
	return session.ErrNotSupported
}

func (d *BLERecon) Stop() error {
	return session.ErrNotSupported
}
