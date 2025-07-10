package ble

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bettercap/bettercap/v2/modules/utils"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"tinygo.org/x/bluetooth"
)

type CurrentDevice struct {
	sync.RWMutex

	device      string
	connected   bool
	writeData   []byte
	writeTarget string
}

func (dev *CurrentDevice) IsEnumerating() bool {
	dev.RLock()
	defer dev.RUnlock()
	return dev.device != ""
}

func (dev *CurrentDevice) Reset() {
	dev.Lock()
	defer dev.Unlock()
	dev.device = ""
	dev.connected = false
	dev.writeData = nil
	dev.writeTarget = ""
}

func (dev *CurrentDevice) ResetWrite() {
	dev.Lock()
	defer dev.Unlock()
	dev.writeData = nil
	dev.writeTarget = ""
}

type BLERecon struct {
	session.SessionModule
	current *CurrentDevice
	// deviceId    int
	adapter     *bluetooth.Adapter
	connTimeout int
	devTTL      int
	quit        chan bool
	done        chan bool
	selector    *utils.ViewSelector
}

func NewBLERecon(s *session.Session) *BLERecon {
	mod := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
		// deviceId:      -1,
		quit:        make(chan bool),
		done:        make(chan bool),
		connTimeout: 15,
		devTTL:      30,
		current:     &CurrentDevice{},
	}

	mod.InitState("scanning")

	mod.selector = utils.ViewSelectorFor(&mod.SessionModule,
		"ble.show",
		[]string{"rssi", "mac", "seen"}, "rssi asc")

	mod.AddHandler(session.NewModuleHandler("ble.recon on", "",
		"Start Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("ble.clear", "",
		"Clear all devices collected by the BLE discovery module.",
		func(args []string) error {
			mod.Session.BLE.Clear()
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("ble.show", "",
		"Show discovered Bluetooth Low Energy devices.",
		func(args []string) error {
			return mod.Show()
		}))

	enum := session.NewModuleHandler("ble.enum MAC", "ble.enum "+network.BLEMacValidator,
		"Enumerate services and characteristics for the given BLE device.",
		func(args []string) error {
			if !mod.Running() {
				return errors.New("module is not running")
			} else if mod.current.IsEnumerating() {
				return fmt.Errorf("an enumeration for %s is already running, please wait", mod.current.device)
			}
			return mod.startEnumeration(mod.normalizeAddress(args[0]))
		})

	enum.Complete("ble.enum", s.BLECompleter)

	mod.AddHandler(enum)

	/*
			write := session.NewModuleHandler("ble.write MAC UUID HEX_DATA", "ble.write "+network.BLEMacValidator+" ([a-fA-F0-9]+) ([a-fA-F0-9]+)",
				"Write the HEX_DATA buffer to the BLE device with the specified MAC address, to the characteristics with the given UUID.",
				func(args []string) error {
					mac := mod.normalizeAddress(args[0])
					uuid := args[1]
					data, err := hex.DecodeString(args[2])
					if err != nil {
						return fmt.Errorf("Error parsing %s: %s", args[2], err)
					}
					return mod.writeBuffer(mac, uuid, data)
				})

			write.Complete("ble.write", s.BLECompleter)


		mod.AddHandler(write)
	*/
	/*
		mod.AddParam(session.NewIntParameter("ble.device",
			fmt.Sprintf("%d", mod.deviceId),
			"Index of the HCI device to use, -1 to autodetect."))
	*/

	mod.AddParam(session.NewIntParameter("ble.timeout",
		fmt.Sprintf("%d", mod.connTimeout),
		"Connection timeout in seconds."))

	mod.AddParam(session.NewIntParameter("ble.ttl",
		fmt.Sprintf("%d", mod.devTTL),
		"Seconds of inactivity for a device to be pruned."))

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

func (mod *BLERecon) normalizeAddress(address string) string {
	// macOS does not provide real MACs
	if !strings.ContainsRune(address, '-') {
		address = network.NormalizeMac(address)
	}
	return address
}

func (mod *BLERecon) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if mod.adapter == nil {
		/*
			if err, mod.deviceId = mod.IntParam("ble.device"); err != nil {
				return err
			}

			mod.Debug("initializing device (id:%d) ...", mod.deviceId)
		*/

		mod.Debug("initializing device ...")

		mod.adapter = bluetooth.DefaultAdapter
		if err := mod.adapter.Enable(); err != nil {
			return err
		}
	}

	if err, mod.connTimeout = mod.IntParam("ble.timeout"); err != nil {
		return err
	} else if err, mod.devTTL = mod.IntParam("ble.ttl"); err != nil {
		return err
	}

	return nil
}

func (mod *BLERecon) onDevice(adapter *bluetooth.Adapter, scanResult bluetooth.ScanResult) {
	mod.Session.BLE.AddIfNew(mod.normalizeAddress(scanResult.Address.String()), scanResult)
}

func (mod *BLERecon) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		go mod.pruner()

		if err := mod.adapter.Scan(mod.onDevice); err != nil {
			mod.Error("could not perform scan: %v", err)
		}

		mod.done <- true
	})
}

func (mod *BLERecon) Stop() error {
	return mod.SetRunning(false, func() {
		if mod.adapter != nil {
			mod.Debug("stopping scan")
			if err := mod.adapter.StopScan(); err != nil {
				mod.Warning("error stopping scan: %v", err)
			}
			<-mod.done
		}

		mod.Debug("module stopped, cleaning state")
		// mod.adapter = nil
		mod.current.Reset()
		mod.ResetState()
	})
}

func (mod *BLERecon) pruner() {
	blePresentInterval := time.Duration(mod.devTTL) * time.Second
	mod.Debug("started devices pruner with ttl %s", blePresentInterval)

	for mod.Running() {
		for _, dev := range mod.Session.BLE.Devices() {
			if time.Since(dev.LastSeen) > blePresentInterval {
				mod.Session.BLE.Remove(mod.normalizeAddress(dev.Address))
			}
		}
		time.Sleep(5 * time.Second)
	}
}
