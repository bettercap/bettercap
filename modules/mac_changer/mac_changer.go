package mac_changer

import (
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

type MacChanger struct {
	session.SessionModule
	iface       string
	originalMac net.HardwareAddr
	fakeMac     net.HardwareAddr
}

func NewMacChanger(s *session.Session) *MacChanger {
	mod := &MacChanger{
		SessionModule: session.NewSessionModule("mac.changer", s),
	}

	mod.AddParam(session.NewStringParameter("mac.changer.iface",
		session.ParamIfaceName,
		"",
		"Name of the interface to use."))

	mod.AddParam(session.NewStringParameter("mac.changer.address",
		session.ParamRandomMAC,
		"[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}",
		"Hardware address to apply to the interface."))

	mod.AddHandler(session.NewModuleHandler("mac.changer on", "",
		"Start mac changer module.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("mac.changer off", "",
		"Stop mac changer module and restore original mac address.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *MacChanger) Name() string {
	return "mac.changer"
}

func (mod *MacChanger) Description() string {
	return "Change active interface mac address."
}

func (mod *MacChanger) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *MacChanger) Configure() (err error) {
	var changeTo string

	if err, mod.iface = mod.StringParam("mac.changer.iface"); err != nil {
		return err
	} else if err, changeTo = mod.StringParam("mac.changer.address"); err != nil {
		return err
	}

	changeTo = network.NormalizeMac(changeTo)
	if mod.fakeMac, err = net.ParseMAC(changeTo); err != nil {
		return err
	}

	mod.originalMac = mod.Session.Interface.HW

	return nil
}

func (mod *MacChanger) setMac(mac net.HardwareAddr) error {
	var args []string

	os := runtime.GOOS
	if strings.Contains(os, "bsd") || os == "darwin" {
		args = []string{mod.iface, "ether", mac.String()}
	} else if os == "linux" || os == "android" {
		args = []string{mod.iface, "hw", "ether", mac.String()}
	} else {
		return fmt.Errorf("OS %s is not supported by mac.changer module.", os)
	}

	_, err := core.Exec("ifconfig", args)
	if err == nil {
		mod.Session.Interface.HW = mac
	}

	return err
}

func (mod *MacChanger) Start() error {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err := mod.Configure(); err != nil {
		return err
	} else if err := mod.setMac(mod.fakeMac); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("interface mac address set to %s", tui.Bold(mod.fakeMac.String()))
	})
}

func (mod *MacChanger) Stop() error {
	return mod.SetRunning(false, func() {
		if err := mod.setMac(mod.originalMac); err == nil {
			mod.Info("interface mac address restored to %s", tui.Bold(mod.originalMac.String()))
		} else {
			mod.Error("error while restoring mac address: %s", err)
		}
	})
}
