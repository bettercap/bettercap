// +build !windows
// +build !darwin

package ble

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	golog "log"
	"time"

	"github.com/bettercap/bettercap/modules/utils"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/gatt"
)

type BLERecon struct {
	session.SessionModule
	gattDevice  gatt.Device
	currDevice  *network.BLEDevice
	writeUUID   *gatt.UUID
	writeData   []byte
	connected   bool
	connTimeout time.Duration
	quit        chan bool
	done        chan bool
	selector    *utils.ViewSelector
}

func NewBLERecon(s *session.Session) *BLERecon {
	mod := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
		gattDevice:    nil,
		quit:          make(chan bool),
		done:          make(chan bool),
		connTimeout:   time.Duration(10) * time.Second,
		currDevice:    nil,
		connected:     false,
	}

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
			if mod.isEnumerating() {
				return fmt.Errorf("An enumeration for %s is already running, please wait.", mod.currDevice.Device.ID())
			}

			mod.writeData = nil
			mod.writeUUID = nil

			return mod.enumAllTheThings(network.NormalizeMac(args[0]))
		})

	enum.Complete("ble.enum", s.BLECompleter)

	mod.AddHandler(enum)

	write := session.NewModuleHandler("ble.write MAC UUID HEX_DATA", "ble.write "+network.BLEMacValidator+" ([a-fA-F0-9]+) ([a-fA-F0-9]+)",
		"Write the HEX_DATA buffer to the BLE device with the specified MAC address, to the characteristics with the given UUID.",
		func(args []string) error {
			mac := network.NormalizeMac(args[0])
			uuid, err := gatt.ParseUUID(args[1])
			if err != nil {
				return fmt.Errorf("Error parsing %s: %s", args[1], err)
			}
			data, err := hex.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("Error parsing %s: %s", args[2], err)
			}

			return mod.writeBuffer(mac, uuid, data)
		})

	write.Complete("ble.write", s.BLECompleter)

	mod.AddHandler(write)

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

func (mod *BLERecon) isEnumerating() bool {
	return mod.currDevice != nil
}

func (mod *BLERecon) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted
	} else if mod.gattDevice == nil {
		mod.Debug("initializing device ...")

		// hey Paypal GATT library, could you please just STFU?!
		golog.SetOutput(ioutil.Discard)
		if mod.gattDevice, err = gatt.NewDevice(defaultBLEClientOptions...); err != nil {
			return err
		}

		mod.gattDevice.Handle(
			gatt.PeripheralDiscovered(mod.onPeriphDiscovered),
			gatt.PeripheralConnected(mod.onPeriphConnected),
			gatt.PeripheralDisconnected(mod.onPeriphDisconnected),
		)

		mod.gattDevice.Init(mod.onStateChanged)
	}

	return nil
}

func (mod *BLERecon) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		go mod.pruner()

		<-mod.quit

		mod.Info("stopping scan ...")

		mod.gattDevice.StopScanning()

		mod.done <- true
	})
}

func (mod *BLERecon) Stop() error {
	return mod.SetRunning(false, func() {
		mod.quit <- true
		<-mod.done
	})
}

func (mod *BLERecon) pruner() {
	mod.Debug("started devices pruner ...")

	for mod.Running() {
		for _, dev := range mod.Session.BLE.Devices() {
			if time.Since(dev.LastSeen) > blePresentInterval {
				mod.Session.BLE.Remove(dev.Device.ID())
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (mod *BLERecon) setCurrentDevice(dev *network.BLEDevice) {
	mod.connected = false
	mod.currDevice = dev
}

func (mod *BLERecon) writeBuffer(mac string, uuid gatt.UUID, data []byte) error {
	mod.writeUUID = &uuid
	mod.writeData = data
	return mod.enumAllTheThings(mac)
}

func (mod *BLERecon) enumAllTheThings(mac string) error {
	dev, found := mod.Session.BLE.Get(mac)
	if !found || dev == nil {
		return fmt.Errorf("BLE device with address %s not found.", mac)
	} else if mod.Running() {
		mod.gattDevice.StopScanning()
	}

	mod.setCurrentDevice(dev)
	if err := mod.Configure(); err != nil && err != session.ErrAlreadyStarted {
		return err
	}

	mod.Info("connecting to %s ...", mac)

	go func() {
		time.Sleep(mod.connTimeout)
		if mod.isEnumerating() && !mod.connected {
			mod.Session.Events.Add("ble.connection.timeout", mod.currDevice)
			mod.onPeriphDisconnected(nil, nil)
		}
	}()

	mod.gattDevice.Connect(dev.Device)

	return nil
}
