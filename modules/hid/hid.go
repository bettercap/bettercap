package hid

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bettercap/bettercap/modules/utils"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/nrf24"
)

type HIDRecon struct {
	session.SessionModule
	dongle       *nrf24.Dongle
	waitGroup    *sync.WaitGroup
	channel      int
	devTTL       int
	hopPeriod    time.Duration
	pingPeriod   time.Duration
	sniffPeriod  time.Duration
	lastHop      time.Time
	lastPing     time.Time
	useLNA       bool
	sniffLock    *sync.Mutex
	writeLock    *sync.Mutex
	sniffAddrRaw []byte
	sniffAddr    string
	sniffType    string
	pingPayload  []byte
	inSniffMode  bool
	sniffSilent  bool
	inPromMode   bool
	inInjectMode bool
	keyLayout    string
	scriptPath   string
	parser       DuckyParser
	selector     *utils.ViewSelector
}

func NewHIDRecon(s *session.Session) *HIDRecon {
	mod := &HIDRecon{
		SessionModule: session.NewSessionModule("hid", s),
		waitGroup:     &sync.WaitGroup{},
		sniffLock:     &sync.Mutex{},
		writeLock:     &sync.Mutex{},
		devTTL:        1200,
		hopPeriod:     100 * time.Millisecond,
		pingPeriod:    100 * time.Millisecond,
		sniffPeriod:   500 * time.Millisecond,
		lastHop:       time.Now(),
		lastPing:      time.Now(),
		useLNA:        true,
		channel:       1,
		sniffAddrRaw:  nil,
		sniffAddr:     "",
		inSniffMode:   false,
		inPromMode:    false,
		inInjectMode:  false,
		sniffSilent:   true,
		pingPayload:   []byte{0x0f, 0x0f, 0x0f, 0x0f},
		keyLayout:     "US",
		scriptPath:    "",
	}

	mod.State.Store("sniffing", &mod.sniffAddr)
	mod.State.Store("injecting", &mod.inInjectMode)
	mod.State.Store("layouts", SupportedLayouts())

	mod.AddHandler(session.NewModuleHandler("hid.recon on", "",
		"Start scanning for HID devices on the 2.4Ghz spectrum.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("hid.recon off", "",
		"Stop scanning for HID devices on the 2.4Ghz spectrum.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("hid.clear", "",
		"Clear all devices collected by the HID discovery module.",
		func(args []string) error {
			mod.Session.HID.Clear()
			return nil
		}))

	sniff := session.NewModuleHandler("hid.sniff ADDRESS", `(?i)^hid\.sniff ([a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}|clear)$`,
		"Start sniffing a specific ADDRESS in order to collect payloads, use 'clear' to stop collecting.",
		func(args []string) error {
			return mod.setSniffMode(args[0], false)
		})

	sniff.Complete("hid.sniff", s.HIDCompleter)

	mod.AddHandler(sniff)

	mod.AddHandler(session.NewModuleHandler("hid.show", "",
		"Show a list of detected HID devices on the 2.4Ghz spectrum.",
		func(args []string) error {
			return mod.Show()
		}))

	inject := session.NewModuleHandler("hid.inject ADDRESS LAYOUT FILENAME", `(?i)^hid\.inject ([a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2})\s+(.+)\s+(.+)$`,
		"Parse the duckyscript FILENAME and inject it as HID frames spoofing the device ADDRESS, using the LAYOUT keyboard mapping.",
		func(args []string) error {
			if err := mod.setInjectionMode(args[0]); err != nil {
				return err
			}
			mod.keyLayout = args[1]
			mod.scriptPath = args[2]
			return nil
		})

	inject.Complete("hid.inject", s.HIDCompleter)

	mod.AddHandler(inject)

	mod.AddParam(session.NewIntParameter("hid.ttl",
		fmt.Sprintf("%d", mod.devTTL),
		"Seconds of inactivity to consider a device as not in range."))

	mod.AddParam(session.NewBoolParameter("hid.lna",
		"true",
		"If true, enable the LNA power amplifier for CrazyRadio devices."))

	mod.AddParam(session.NewIntParameter("hid.hop.period",
		"100",
		"Time in milliseconds to stay on each channel before hopping to the next one."))

	mod.AddParam(session.NewIntParameter("hid.ping.period",
		"100",
		"Time in milliseconds to attempt to ping a device on a given channel while in sniffer mode."))

	mod.AddParam(session.NewIntParameter("hid.sniff.period",
		"500",
		"Time in milliseconds to automatically sniff payloads from a device, once it's detected, in order to determine its type."))

	builders := availBuilders()

	mod.AddParam(session.NewStringParameter("hid.force.type",
		"logitech",
		fmt.Sprintf("(%s)", strings.Join(builders, "|")),
		fmt.Sprintf("If the device is not visible or its type has not being detected, force the device type to this value. Accepted values: %s", strings.Join(builders, ", "))))

	mod.parser = DuckyParser{mod}
	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "hid.show", []string{"mac", "seen"}, "mac desc")

	return mod
}

func (mod HIDRecon) Name() string {
	return "hid"
}

func (mod HIDRecon) Description() string {
	return "A scanner and frames injection module for HID devices on the 2.4Ghz spectrum, using Nordic Semiconductor nRF24LU1+ based USB dongles and Bastille Research RFStorm firmware."
}

func (mod HIDRecon) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com> (this module and the nrf24 client library), Bastille Research (the rfstorm firmware and original research), phikshun and infamy for JackIt."
}

func (mod *HIDRecon) Configure() error {
	var err error
	var n int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if err, mod.useLNA = mod.BoolParam("hid.lna"); err != nil {
		return err
	}

	if err, mod.devTTL = mod.IntParam("hid.ttl"); err != nil {
		return err
	}

	if err, n = mod.IntParam("hid.hop.period"); err != nil {
		return err
	} else {
		mod.hopPeriod = time.Duration(n) * time.Millisecond
	}

	if err, n = mod.IntParam("hid.ping.period"); err != nil {
		return err
	} else {
		mod.pingPeriod = time.Duration(n) * time.Millisecond
	}

	if err, n = mod.IntParam("hid.sniff.period"); err != nil {
		return err
	} else {
		mod.sniffPeriod = time.Duration(n) * time.Millisecond
	}

	if mod.dongle, err = nrf24.Open(); err != nil {
		return fmt.Errorf("make sure that a nRF24LU1+ based USB dongle is connected and running the rfstorm firmware: %s", err)
	}

	mod.Debug("using device %s", mod.dongle.String())

	if mod.useLNA {
		if err = mod.dongle.EnableLNA(); err != nil {
			return fmt.Errorf("make sure your device supports LNA, otherwise set hid.lna to false and retry: %s", err)
		}
		mod.Debug("LNA enabled")
	}

	return nil
}

func (mod *HIDRecon) forceStop() error {
	return mod.SetRunning(false, func() {
		if mod.dongle != nil {
			mod.dongle.Close()
			mod.Debug("device closed")
		}
	})
}
func (mod *HIDRecon) Stop() error {
	return mod.SetRunning(false, func() {
		mod.waitGroup.Wait()
		if mod.dongle != nil {
			mod.dongle.Close()
			mod.Debug("device closed")
		}
	})
}
