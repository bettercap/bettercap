// +build !windows

package modules

import (
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/currantlabs/gatt"
)

var (
	bleAliveInterval   = time.Duration(5) * time.Second
	blePresentInterval = time.Duration(30) * time.Second
)

type BLERecon struct {
	session.SessionModule
	gattDevice  gatt.Device
	currDevice  *network.BLEDevice
	connected   bool
	connTimeout time.Duration
	quit        chan bool
}

func NewBLERecon(s *session.Session) *BLERecon {
	d := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
		gattDevice:    nil,
		quit:          make(chan bool),
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

	d.AddHandler(session.NewModuleHandler("ble.enum MAC", "ble.enum ([a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2}:[a-fA-F0-9]{1,2})",
		"Enumerate services and characteristics for the given BLE device.",
		func(args []string) error {
			if d.isEnumerating() == true {
				return fmt.Errorf("An enumeration for %s is already running, please wait.", d.currDevice.Device.ID())
			}

			return d.enumAll(network.NormalizeMac(args[0]))
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
	if d.gattDevice == nil {
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

func (d *BLERecon) onStateChanged(dev gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Info("Starting BLE discovery ...")
		dev.Scan([]gatt.UUID{}, true)
		return
	default:
		log.Warning("Unexpected BLE state: %v", s)
	}
}

func (d *BLERecon) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	d.Session.BLE.AddIfNew(p.ID(), p, a, rssi)
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
	if d.Running() {
		return session.ErrAlreadyStarted
	} else if err := d.Configure(); err != nil {
		return err
	}

	return d.SetRunning(true, func() {
		log.Debug("Initializing BLE device ...")

		go d.pruner()

		<-d.quit

		log.Info("Stopping BLE scan ...")

		d.gattDevice.StopScanning()
	})
}

func (d *BLERecon) getRow(dev *network.BLEDevice) []string {
	address := network.NormalizeMac(dev.Device.ID())
	vendor := dev.Vendor
	sinceSeen := time.Since(dev.LastSeen)
	lastSeen := dev.LastSeen.Format("15:04:05")

	if sinceSeen <= bleAliveInterval {
		lastSeen = core.Bold(lastSeen)
	} else if sinceSeen > blePresentInterval {
		lastSeen = core.Dim(lastSeen)
		address = core.Dim(address)
	}

	isConnectable := core.Red("no")
	if dev.Advertisement.Connectable == true {
		isConnectable = core.Green("yes")
	}

	return []string{
		fmt.Sprintf("%d dBm", dev.RSSI),
		address,
		dev.Device.Name(),
		vendor,
		isConnectable,
		lastSeen,
	}
}

func (d *BLERecon) Show() error {
	devices := d.Session.BLE.Devices()

	sort.Sort(ByBLERSSISorter(devices))

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, d.getRow(dev))
	}
	nrows := len(rows)

	columns := []string{"RSSI", "Address", "Name", "Vendor", "Connectable", "Last Seen"}

	if nrows > 0 {
		core.AsTable(os.Stdout, columns, rows)
	}

	d.Session.Refresh()
	return nil
}

func (d *BLERecon) setCurrentDevice(dev *network.BLEDevice) {
	d.connected = false
	d.currDevice = dev
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
	defer func() {
		log.Info("Disconnecting from %s ...", p.ID())
		p.Device().CancelConnection(p)
	}()

	// timed out
	if d.currDevice == nil {
		log.Debug("Connected to %s but after the timeout :(", p.ID())
		return
	}

	d.connected = true

	d.Session.Events.Add("ble.device.connected", d.currDevice)

	/*
		if err := p.SetMTU(500); err != nil {
			log.Warning("Failed to set MTU: %s", err)
		} */

	log.Info("Enumerating all the things for %s!", p.ID())
	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Error("Error discovering services: %s", err)
		return
	}

	d.showServices(p, services)
}

func parseProperties(ch *gatt.Characteristic) (props []string, isReadable bool) {
	isReadable = false
	props = make([]string, 0)
	mask := ch.Properties()

	if (mask & gatt.CharBroadcast) != 0 {
		props = append(props, "bcast")
	}
	if (mask & gatt.CharRead) != 0 {
		isReadable = true
		props = append(props, "read")
	}
	if (mask&gatt.CharWriteNR) != 0 || (mask&gatt.CharWrite) != 0 {
		props = append(props, core.Bold("write"))
	}
	if (mask & gatt.CharNotify) != 0 {
		props = append(props, "notify")
	}
	if (mask & gatt.CharIndicate) != 0 {
		props = append(props, "indicate")
	}
	if (mask & gatt.CharSignedWrite) != 0 {
		props = append(props, core.Yellow("*write"))
	}
	if (mask & gatt.CharExtended) != 0 {
		props = append(props, "x")
	}

	return
}

func parseRawData(raw []byte) string {
	s := ""
	for _, b := range raw {
		if b != 00 && strconv.IsPrint(rune(b)) == false {
			return fmt.Sprintf("%x", raw)
		} else if b == 0 {
			break
		} else {
			s += fmt.Sprintf("%c", b)
		}
	}

	return core.Yellow(s)
}

func (d *BLERecon) showServices(p gatt.Peripheral, services []*gatt.Service) {
	columns := []string{"Handles", "Service > Characteristics", "Properties", "Data"}
	rows := make([][]string, 0)

	for _, svc := range services {
		d.Session.Events.Add("ble.device.service.discovered", svc)

		name := svc.Name()
		if name == "" {
			name = svc.UUID().String()
		} else {
			name = fmt.Sprintf("%s (%s)", core.Green(name), core.Dim(svc.UUID().String()))
		}

		row := []string{
			fmt.Sprintf("%04x -> %04x", svc.Handle(), svc.EndHandle()),
			name,
			"",
			"",
		}

		rows = append(rows, row)

		chars, err := p.DiscoverCharacteristics(nil, svc)
		if err != nil {
			log.Error("Error while enumerating chars for service %s: %s", svc.UUID(), err)
			continue
		}

		for _, ch := range chars {
			d.Session.Events.Add("ble.device.characteristic.discovered", ch)

			name = ch.Name()
			if name == "" {
				name = "    " + ch.UUID().String()
			} else {
				name = fmt.Sprintf("    %s (%s)", core.Green(name), core.Dim(ch.UUID().String()))
			}

			props, isReadable := parseProperties(ch)

			data := ""
			if isReadable {
				raw, err := p.ReadCharacteristic(ch)
				if err != nil {
					data = core.Red(err.Error())
				} else {
					data = parseRawData(raw)
				}
			}

			row := []string{
				fmt.Sprintf("%04x", ch.Handle()),
				name,
				strings.Join(props, ", "),
				data,
			}

			rows = append(rows, row)
		}
	}

	core.AsTable(os.Stdout, columns, rows)
	d.Session.Refresh()
}

func (d *BLERecon) enumAll(mac string) error {
	dev, found := d.Session.BLE.Get(mac)
	if found == false {
		return fmt.Errorf("BLE device with address %s not found.", mac)
	}

	services := dev.Device.Services()
	if len(services) > 0 {
		d.showServices(dev.Device, services)
	} else {
		d.setCurrentDevice(dev)

		log.Info("Connecting to %s ...", mac)

		if d.Running() {
			d.gattDevice.StopScanning()
		}

		go func() {
			time.Sleep(d.connTimeout)
			if d.currDevice != nil && d.connected == false {
				d.Session.Events.Add("ble.connection.timeout", d.currDevice)
				d.onPeriphDisconnected(nil, nil)
			}
		}()

		d.gattDevice.Connect(dev.Device)
	}

	return nil
}

func (d *BLERecon) Stop() error {
	return d.SetRunning(false, func() {
		d.quit <- true
	})
}
