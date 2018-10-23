// +build !windows
// +build !darwin

package modules

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	golog "log"
	"time"

	"github.com/bettercap/bettercap/log"
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
}

func NewBLERecon(s *session.Session) *BLERecon {
	d := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
		gattDevice:    nil,
		quit:          make(chan bool),
		done:          make(chan bool),
		connTimeout:   time.Duration(10) * time.Second,
		currDevice:    nil,
		connected:     false,
	}

	d.AddHandler(session.NewModuleHandler("ble.recon on", "",
		"Start Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return d.Start()
		}))

	d.AddHandler(session.NewModuleHandler("ble.recon off", "",
		"Stop Bluetooth Low Energy devices discovery.",
		func(args []string) error {
			return d.Stop()
		}))

	d.AddHandler(session.NewModuleHandler("ble.show", "",
		"Show discovered Bluetooth Low Energy devices.",
		func(args []string) error {
			return d.Show()
		}))

	d.AddHandler(session.NewModuleHandler("ble.enum MAC", "ble.enum "+network.BLEMacValidator,
		"Enumerate services and characteristics for the given BLE device.",
		func(args []string) error {
			if d.isEnumerating() {
				return fmt.Errorf("An enumeration for %s is already running, please wait.", d.currDevice.Device.ID())
			}

			d.writeData = nil
			d.writeUUID = nil

			return d.enumAllTheThings(network.NormalizeMac(args[0]))
		}))

	d.AddHandler(session.NewModuleHandler("ble.write MAC UUID HEX_DATA", "ble.write "+network.BLEMacValidator+" ([a-fA-F0-9]+) ([a-fA-F0-9]+)",
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

			return d.writeBuffer(mac, uuid, data)
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

func (d *BLERecon) isEnumerating() bool {
	return d.currDevice != nil
}

func (d *BLERecon) Configure() (err error) {
	if d.Running() {
		return session.ErrAlreadyStarted
	} else if d.gattDevice == nil {
		log.Info("Initializing BLE device ...")

		// hey Paypal GATT library, could you please just STFU?!
		golog.SetOutput(ioutil.Discard)
		if d.gattDevice, err = gatt.NewDevice(defaultBLEClientOptions...); err != nil {
			return err
		}

		d.gattDevice.Handle(
			gatt.PeripheralDiscovered(d.onPeriphDiscovered),
			gatt.PeripheralConnected(d.onPeriphConnected),
			gatt.PeripheralDisconnected(d.onPeriphDisconnected),
		)

		d.gattDevice.Init(d.onStateChanged)
	}

	return nil
}

func (d *BLERecon) Start() error {
	if err := d.Configure(); err != nil {
		return err
	}

	return d.SetRunning(true, func() {
		go d.pruner()

		<-d.quit

		log.Info("Stopping BLE scan ...")

		d.gattDevice.StopScanning()

		d.done <- true
	})
}

func (d *BLERecon) Stop() error {
	return d.SetRunning(false, func() {
		d.quit <- true
		<-d.done
	})
}

func (d *BLERecon) pruner() {
	log.Debug("Started BLE devices pruner ...")

	for d.Running() {
		for _, dev := range d.Session.BLE.Devices() {
			if time.Since(dev.LastSeen) > blePresentInterval {
				d.Session.BLE.Remove(dev.Device.ID())
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (d *BLERecon) setCurrentDevice(dev *network.BLEDevice) {
	d.connected = false
	d.currDevice = dev
}

func (d *BLERecon) writeBuffer(mac string, uuid gatt.UUID, data []byte) error {
	d.writeUUID = &uuid
	d.writeData = data
	return d.enumAllTheThings(mac)
}

func (d *BLERecon) enumAllTheThings(mac string) error {
	dev, found := d.Session.BLE.Get(mac)
	if !found || dev == nil {
		return fmt.Errorf("BLE device with address %s not found.", mac)
	}
	if d.Running() {
		d.gattDevice.StopScanning()
	}

	d.setCurrentDevice(dev)
	if err := d.Configure(); err != nil && err != session.ErrAlreadyStarted {
		return err
	}

	log.Info("Connecting to %s ...", mac)

	go func() {
		time.Sleep(d.connTimeout)
		if d.isEnumerating() && !d.connected {
			d.Session.Events.Add("ble.connection.timeout", d.currDevice)
			d.onPeriphDisconnected(nil, nil)
		}
	}()

	d.gattDevice.Connect(dev.Device)

	return nil
}
