// +build windows darwin

package modules

import (
	"errors"

	"github.com/bettercap/bettercap/session"
)

var (
	notSupported = errors.New("ble.recon is not supported on this OS")
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
			return notSupported
		}))

	d.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return notSupported
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
	return notSupported
}

func (d *BLERecon) Start() error {
	return notSupported
}

func (d *BLERecon) Stop() error {
	return notSupported
}
