// +build windows

package ble

import (
	"github.com/bettercap/bettercap/session"
)

type BLERecon struct {
	session.SessionModule
}

func NewBLERecon(s *session.Session) *BLERecon {
	mod := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
	}

	mod.AddHandler(session.NewModuleHandler("ble.recon on", "",
		"Start Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	mod.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	return mod
}

func (mod BLERecon) Name() string {
	return "ble.recon"
}

func (mod BLERecon) Description() string {
	return "Bluetooth Low Energy devices discovery."
}

func (mod BLERecon) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *BLERecon) Configure() (err error) {
	return session.ErrNotSupported
}

func (mod *BLERecon) Start() error {
	return session.ErrNotSupported
}

func (mod *BLERecon) Stop() error {
	return session.ErrNotSupported
}
