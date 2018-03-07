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

const (
	macRegexp = "([a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2})"
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

	d.AddHandler(session.NewModuleHandler("ble.enum MAC", "ble.enum "+macRegexp,
		"Enumerate services and characteristics for the given BLE device.",
		func(args []string) error {
			if d.isEnumerating() == true {
				return fmt.Errorf("An enumeration for %s is already running, please wait.", d.currDevice.Device.ID())
			}

			d.writeData = nil
			d.writeUUID = nil

			return d.enumAllTheThings(network.NormalizeMac(args[0]))
		}))

	d.AddHandler(session.NewModuleHandler("ble.write MAC UUID HEX_DATA", "ble.write "+macRegexp+" ([a-fA-F0-9]+) ([a-fA-F0-9]+)",
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

func (d *BLERecon) writeBuffer(mac string, uuid gatt.UUID, data []byte) error {
	d.writeUUID = &uuid
	d.writeData = data
	return d.enumAllTheThings(mac)
}

func (d *BLERecon) enumAllTheThings(mac string) error {
	dev, found := d.Session.BLE.Get(mac)
	if found == false || dev == nil {
		return fmt.Errorf("BLE device with address %s not found.", mac)
	} else if d.Running() {
		d.gattDevice.StopScanning()
	}

	d.setCurrentDevice(dev)
	if err := d.Configure(); err != nil && err != session.ErrAlreadyStarted {
		return err
	}

	log.Info("Connecting to %s ...", mac)

	go func() {
		time.Sleep(d.connTimeout)
		if d.isEnumerating() && d.connected == false {
			d.Session.Events.Add("ble.connection.timeout", d.currDevice)
			d.onPeriphDisconnected(nil, nil)
		}
	}()

	d.gattDevice.Connect(dev.Device)

	return nil
}

func (d *BLERecon) Stop() error {
	return d.SetRunning(false, func() {
		d.quit <- true
		<-d.done
	})
}

func (d *BLERecon) setCurrentDevice(dev *network.BLEDevice) {
	d.connected = false
	d.currDevice = dev
}

func (d *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		if d.currDevice == nil {
			log.Info("Starting BLE discovery ...")
			dev.Scan([]gatt.UUID{}, true)
		}
	case gatt.StatePoweredOff:
		d.gattDevice = nil

	default:
		log.Warning("Unexpected BLE state: %v", s)
	}
}

func (d *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	d.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
}

func (d *BLERecon) onPeriphDisconnected(p gatt.Peripheral, err error) {
	if d.Running() {
		// restore scanning
		log.Info("Device disconnected, restoring BLE discovery.")
		d.setCurrentDevice(nil)
		d.gattDevice.Scan([]gatt.UUID{}, true)
	}
}

func (d *BLERecon) onPeriphConnected(p gatt.Peripheral, err error) {
	// timed out
	if d.currDevice == nil {
		log.Warning("Connected to %s but after the timeout :(", p.ID())
		return
	}

	d.connected = true

	defer func(per gatt.Peripheral) {
		log.Info("Disconnecting from %s ...", per.ID())
		per.Device().CancelConnection(per)
	}(p)

	d.Session.Events.Add("ble.device.connected", d.currDevice)

	if err := p.SetMTU(500); err != nil {
		log.Warning("Failed to set MTU: %s", err)
	}

	log.Info("Connected, enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Error("Error discovering services: %s", err)
		return
	}

	d.showServices(p, services)
}
