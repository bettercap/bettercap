package modules

import (
	"fmt"

	"github.com/bettercap/bettercap/session"
)

type BLERecon struct {
	session.SessionModule
}

func NewBLERecon(s *session.Session) *BLERecon {
	d := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
	}

	d.AddHandler(session.NewModuleHandler("ble.recon on", "",
		"Start Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return fmt.Errorf("ble.recon is not supported on Windows")
		}))

	d.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return fmt.Errorf("ble.recon is not supported on Windows")
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
	return fmt.Errorf("ble.recon is not supported on Windows")
}

func (d *BLERecon) Start() error {
	return fmt.Errorf("ble.recon is not supported on Windows")
}

func (d *BLERecon) Stop() error {
	return fmt.Errorf("ble.recon is not supported on Windows")
}
