// +build windows

package hid

import (
	"github.com/bettercap/bettercap/session"
)

type HIDRecon struct {
	session.SessionModule
}

func NewHIDRecon(s *session.Session) *HIDRecon {
	mod := &HIDRecon{
		SessionModule: session.NewSessionModule("hid.recon", s),
	}

	mod.AddHandler(session.NewModuleHandler("hid.recon on", "",
		"Start scanning for HID devices on the 2.4Ghz spectrum.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	mod.AddHandler(session.NewModuleHandler("hid.recon off", "",
		"Stop scanning for HID devices on the 2.4Ghz spectrum.",
		func(args []string) error {
			return session.ErrNotSupported
		}))

	return mod
}

func (mod HIDRecon) Name() string {
	return "hid"
}

func (mod HIDRecon) Description() string {
	return "A scanner and frames injection module for HID devices on the 2.4Ghz spectrum, using Nordic Semiconductor nRF24LU1+ based USB dongles and Bastille Research RFStorm firmware."
}

func (mod HIDRecon) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *HIDRecon) Configure() (err error) {
	return session.ErrNotSupported
}

func (mod *HIDRecon) Start() error {
	return session.ErrNotSupported
}

func (mod *HIDRecon) Stop() error {
	return session.ErrNotSupported
}
