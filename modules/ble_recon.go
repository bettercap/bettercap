// +build !windows

package modules

import (
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/bettercap/gatt"

	"github.com/olekukonko/tablewriter"
)

var (
	bleAliveInterval   = time.Duration(5) * time.Second
	blePresentInterval = time.Duration(30) * time.Second
)

type BLERecon struct {
	session.SessionModule
	gattDevice gatt.Device
	quit       chan bool
}

func NewBLERecon(s *session.Session) *BLERecon {
	d := &BLERecon{
		SessionModule: session.NewSessionModule("ble.recon", s),
		gattDevice:    nil,
		quit:          make(chan bool),
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

func (d *BLERecon) Configure() (err error) {
	if d.gattDevice == nil {
		// hey Paypal GATT library, could you please just STFU?!
		golog.SetOutput(ioutil.Discard)
		if d.gattDevice, err = gatt.NewDevice(defaultBLEClientOptions...); err != nil {
			return err
		}

		d.gattDevice.Handle(gatt.PeripheralDiscovered(d.onPeriphDiscovered))
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

	isConnectable := core.Red("✕")
	if dev.Advertisement.Connectable == true {
		isConnectable = core.Green("✓")
	}

	return []string{
		fmt.Sprintf("%d dBm", dev.RSSI),
		address,
		dev.Device.Name(),
		vendor,
		strings.Join(dev.Advertisement.Flags, ", "),
		isConnectable,
		lastSeen,
	}
}

func (d *BLERecon) showTable(header []string, rows [][]string) {
	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetColWidth(80)
	table.AppendBulk(rows)
	table.Render()
}

func (d *BLERecon) Show() error {
	devices := d.Session.BLE.Devices()

	sort.Sort(ByBLERSSISorter(devices))

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, d.getRow(dev))
	}
	nrows := len(rows)

	columns := []string{"RSSI", "Address", "Name", "Vendor", "Flags", "Connectable", "Last Seen"}

	if nrows > 0 {
		d.showTable(columns, rows)
	}

	d.Session.Refresh()
	return nil
}

func (d *BLERecon) Stop() error {
	return d.SetRunning(false, func() {
		d.quit <- true
	})
}
