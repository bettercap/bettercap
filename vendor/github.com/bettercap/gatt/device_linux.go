package gatt

import (
	"encoding/binary"
	"net"

	"github.com/bettercap/gatt/linux"
	"github.com/bettercap/gatt/linux/cmd"
)

type device struct {
	deviceHandler

	hci   *linux.HCI
	state State

	// All the following fields are only used peripheralManager (server) implementation.
	svcs  []*Service
	attrs *attrRange

	devID   int
	chkLE   bool
	maxConn int

	advData   *cmd.LESetAdvertisingData
	scanResp  *cmd.LESetScanResponseData
	advParam  *cmd.LESetAdvertisingParameters
	scanParam *cmd.LESetScanParameters
}

func NewDevice(opts ...Option) (Device, error) {
	d := &device{
		maxConn: 1,    // Support 1 connection at a time.
		devID:   -1,   // Find an available HCI device.
		chkLE:   true, // Check if the device supports LE.

		advParam: &cmd.LESetAdvertisingParameters{
			AdvertisingIntervalMin:  0x800,     // [0x0800]: 0.625 ms * 0x0800 = 1280.0 ms
			AdvertisingIntervalMax:  0x800,     // [0x0800]: 0.625 ms * 0x0800 = 1280.0 ms
			AdvertisingType:         0x00,      // [0x00]: ADV_IND, 0x01: DIRECT(HIGH), 0x02: SCAN, 0x03: NONCONN, 0x04: DIRECT(LOW)
			OwnAddressType:          0x00,      // [0x00]: public, 0x01: random
			DirectAddressType:       0x00,      // [0x00]: public, 0x01: random
			DirectAddress:           [6]byte{}, // Public or Random Address of the device to be connected
			AdvertisingChannelMap:   0x7,       // [0x07] 0x01: ch37, 0x2: ch38, 0x4: ch39
			AdvertisingFilterPolicy: 0x00,
		},
		scanParam: &cmd.LESetScanParameters{
			LEScanType:           0x01,   // [0x00]: passive, 0x01: active
			LEScanInterval:       0x0010, // [0x10]: 0.625ms * 16
			LEScanWindow:         0x0010, // [0x10]: 0.625ms * 16
			OwnAddressType:       0x00,   // [0x00]: public, 0x01: random
			ScanningFilterPolicy: 0x00,   // [0x00]: accept all, 0x01: ignore non-white-listed.
		},
	}

	d.Option(opts...)
	h, err := linux.NewHCI(d.devID, d.chkLE, d.maxConn)
	if err != nil {
		return nil, err
	}

	d.hci = h
	return d, nil
}

func (d *device) Init(f func(Device, State)) error {
	d.hci.AcceptMasterHandler = func(pd *linux.PlatData) {
		a := pd.Address
		c := newCentral(d.attrs, net.HardwareAddr([]byte{a[5], a[4], a[3], a[2], a[1], a[0]}), pd.Conn)
		if d.centralConnected != nil {
			d.centralConnected(c)
		}
		c.loop()
		if d.centralDisconnected != nil {
			d.centralDisconnected(c)
		}
	}
	d.hci.AcceptSlaveHandler = func(pd *linux.PlatData) {
		p := &peripheral{
			d:     d,
			pd:    pd,
			l2c:   pd.Conn,
			reqc:  make(chan message),
			quitc: make(chan struct{}),
			sub:   newSubscriber(),
		}
		if d.peripheralConnected != nil {
			go d.peripheralConnected(p, nil)
		}
		p.loop()
		if d.peripheralDisconnected != nil {
			d.peripheralDisconnected(p, nil)
		}
	}
	d.hci.AdvertisementHandler = func(pd *linux.PlatData) {
		a := &Advertisement{}
		a.unmarshall(pd.Data)
		a.Connectable = pd.Connectable
		p := &peripheral{pd: pd, d: d}
		if d.peripheralDiscovered != nil {
			pd.Name = a.LocalName
			d.peripheralDiscovered(p, a, int(pd.RSSI))
		}
	}
	d.state = StatePoweredOn
	d.stateChanged = f
	go d.stateChanged(d, d.state)
	return nil
}

func (d *device) Stop() error {
	d.state = StatePoweredOff
	defer d.stateChanged(d, d.state)
	return d.hci.Close()
}

func (d *device) AddService(s *Service) error {
	d.svcs = append(d.svcs, s)
	d.attrs = generateAttributes(d.svcs, uint16(1)) // ble attrs start at 1
	return nil
}

func (d *device) RemoveAllServices() error {
	d.svcs = nil
	d.attrs = nil
	return nil
}

func (d *device) SetServices(s []*Service) error {
	d.RemoveAllServices()
	d.svcs = append(d.svcs, s...)
	d.attrs = generateAttributes(d.svcs, uint16(1)) // ble attrs start at 1
	return nil
}

func (d *device) Advertise(a *AdvPacket) error {
	d.advData = &cmd.LESetAdvertisingData{
		AdvertisingDataLength: uint8(a.Len()),
		AdvertisingData:       a.Bytes(),
	}

	if err := d.update(); err != nil {
		return err
	}

	return d.hci.SetAdvertiseEnable(true)
}

func (d *device) AdvertiseNameAndServices(name string, uu []UUID) error {
	a := &AdvPacket{}
	a.AppendFlags(flagGeneralDiscoverable | flagLEOnly)
	a.AppendUUIDFit(uu)

	if len(a.b)+len(name)+2 < MaxEIRPacketLength {
		a.AppendName(name)
		d.scanResp = nil
	} else {
		a := &AdvPacket{}
		a.AppendName(name)
		d.scanResp = &cmd.LESetScanResponseData{
			ScanResponseDataLength: uint8(a.Len()),
			ScanResponseData:       a.Bytes(),
		}
	}

	return d.Advertise(a)
}

func (d *device) AdvertiseIBeaconData(b []byte) error {
	a := &AdvPacket{}
	a.AppendFlags(flagGeneralDiscoverable | flagLEOnly)
	a.AppendManufacturerData(0x004C, b)

	return d.Advertise(a)
}

func (d *device) AdvertiseIBeacon(u UUID, major, minor uint16, pwr int8) error {
	b := make([]byte, 23)
	b[0] = 0x02                               // Data type: iBeacon
	b[1] = 0x15                               // Data length: 21 bytes
	copy(b[2:], reverse(u.b))                 // Big endian
	binary.BigEndian.PutUint16(b[18:], major) // Big endian
	binary.BigEndian.PutUint16(b[20:], minor) // Big endian
	b[22] = uint8(pwr)                        // Measured Tx Power
	return d.AdvertiseIBeaconData(b)
}

func (d *device) StopAdvertising() error {
	return d.hci.SetAdvertiseEnable(false)
}

func (d *device) Scan(ss []UUID, dup bool) {
	// TODO: filter
	d.hci.SetScanEnable(true, dup)
}

func (d *device) StopScanning() {
	d.hci.SetScanEnable(false, true)
}

func (d *device) Connect(p Peripheral) {
	d.hci.Connect(p.(*peripheral).pd)
}

func (d *device) CancelConnection(p Peripheral) {
	d.hci.CancelConnection(p.(*peripheral).pd)
}

func (d *device) SendHCIRawCommand(c cmd.CmdParam) ([]byte, error) {
	return d.hci.SendRawCommand(c)
}

// Flush pending advertising settings to the device.
func (d *device) update() error {
	if d.advParam != nil {
		if err := d.hci.SendCmdWithAdvOff(d.advParam); err != nil {
			return err
		}
		d.advParam = nil
	}
	if d.scanResp != nil {
		if err := d.hci.SendCmdWithAdvOff(d.scanResp); err != nil {
			return err
		}
		d.scanResp = nil
	}
	if d.advData != nil {
		if err := d.hci.SendCmdWithAdvOff(d.advData); err != nil {
			return err
		}
		d.advData = nil
	}
	return nil
}
