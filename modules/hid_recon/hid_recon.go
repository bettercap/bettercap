package hid_recon

import (
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
	hopPeriod    time.Duration
	pingPeriod   time.Duration
	sniffPeriod  time.Duration
	lastHop      time.Time
	lastPing     time.Time
	useLNA       bool
	sniffLock    *sync.Mutex
	sniffAddrRaw []byte
	sniffAddr    string
	pingPayload  []byte
	inSniffMode  bool
	inPromMode   bool
	inInjectMode bool
	keyLayout    string
	selector     *utils.ViewSelector
}

func NewHIDRecon(s *session.Session) *HIDRecon {
	mod := &HIDRecon{
		SessionModule: session.NewSessionModule("hid.recon", s),
		waitGroup:     &sync.WaitGroup{},
		sniffLock:     &sync.Mutex{},
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
		pingPayload:   []byte{0x0f, 0x0f, 0x0f, 0x0f},
		keyLayout:     "US",
	}

	mod.AddHandler(session.NewModuleHandler("hid.recon on", "",
		"Start HID recon.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("hid.recon off", "",
		"Stop HID recon.",
		func(args []string) error {
			return mod.Stop()
		}))

	sniff := session.NewModuleHandler("hid.sniff ADDRESS", `(?i)^hid\.sniff ([a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}|clear)$`,
		"TODO TODO",
		func(args []string) error {
			return mod.setSniffMode(args[0])
		})

	sniff.Complete("hid.sniff", s.HIDCompleter)

	mod.AddHandler(sniff)

	mod.AddHandler(session.NewModuleHandler("hid.show", "",
		"TODO TODO",
		func(args []string) error {
			return mod.Show()
		}))

	inject := session.NewModuleHandler("hid.inject ADDRESS", `(?i)^hid\.inject ([a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2})$`,
		"TODO TODO",
		func(args []string) error {
			return mod.setInjectionMode(args[0])
		})

	inject.Complete("hid.inject", s.HIDCompleter)

	mod.AddHandler(inject)

	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "hid.show", []string{"mac", "seen"}, "mac desc")

	return mod
}

func (mod HIDRecon) Name() string {
	return "hid.recon"
}

func (mod HIDRecon) Description() string {
	return "TODO TODO"
}

func (mod HIDRecon) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *HIDRecon) Configure() error {
	var err error

	if mod.dongle, err = nrf24.Open(); err != nil {
		return err
	}

	mod.Debug("using device %s", mod.dongle.String())

	if mod.useLNA {
		if err = mod.dongle.EnableLNA(); err != nil {
			return err
		}
		mod.Debug("LNA enabled")
	}

	return nil
}

func (mod *HIDRecon) doHopping() {
	if mod.inPromMode == false {
		if err := mod.dongle.EnterPromiscMode(); err != nil {
			mod.Error("error entering promiscuous mode: %v", err)
		} else {
			mod.inSniffMode = false
			mod.inPromMode = true
			mod.Debug("device entered promiscuous mode")
		}
	}

	if time.Since(mod.lastHop) >= mod.hopPeriod {
		mod.channel++
		if mod.channel > nrf24.TopChannel {
			mod.channel = 1
		}
		if err := mod.dongle.SetChannel(mod.channel); err != nil {
			mod.Warning("error hopping on channel %d: %v", mod.channel, err)
		} else {
			mod.lastHop = time.Now()
		}
	}
}

func (mod *HIDRecon) onDeviceDetected(buf []byte) {
	if sz := len(buf); sz >= 5 {
		addr, payload := buf[0:5], buf[5:]
		mod.Debug("detected device %x on channel %d (payload:%x)\n", addr, mod.channel, payload)
		if isNew, dev := mod.Session.HID.AddIfNew(addr, mod.channel, payload); isNew {
			// sniff for a while in order to detect the device type
			go func() {
				if err := mod.setSniffMode(dev.Address); err == nil {
					mod.Debug("detecting device type ...")
					defer func() {
						mod.sniffLock.Unlock()
						mod.setSniffMode("clear")
					}()
					// make sure nobody can sniff to another
					// address until we're not done here...
					mod.sniffLock.Lock()

					time.Sleep(mod.sniffPeriod)
				} else {
					mod.Warning("error while sniffing %s: %v", dev.Address, err)
				}
			}()
		}
	}
}

func (mod *HIDRecon) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		mod.Info("hopping on %d channels every %s", nrf24.TopChannel, mod.hopPeriod)
		for mod.Running() {
			if mod.isSniffing() {
				mod.doPing()
			} else {
				mod.doHopping()
			}

			if mod.isInjecting() {
				mod.doInjection()
				mod.setInjectionMode("clear")
				continue
			}

			buf, err := mod.dongle.ReceivePayload()
			if err != nil {
				mod.Warning("error receiving payload from channel %d: %v", mod.channel, err)
				continue
			}

			if mod.isSniffing() {
				mod.onSniffedBuffer(buf)
			} else {
				mod.onDeviceDetected(buf)
			}
		}

		mod.Debug("stopped")
	})
}

func (mod *HIDRecon) Stop() error {
	return mod.SetRunning(false, func() {
		mod.waitGroup.Wait()
		mod.dongle.Close()
		mod.Debug("device closed")
	})
}
