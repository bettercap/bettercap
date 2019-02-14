package gatt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bettercap/gatt/xpc"
)

const (
	peripheralDiscovered   = 37
	peripheralConnected    = 38
	peripheralDisconnected = 40
	// below constants for Yosemite
	rssiRead                   = 55
	includedServicesDiscovered = 63
	serviceDiscovered          = 56
	characteristicsDiscovered  = 64
	characteristicRead         = 71
	characteristicWritten      = 72
	notificationValueSet       = 74
	descriptorsDiscovered      = 76
	descriptorRead             = 79
	descriptorWritten          = 80
)

type device struct {
	deviceHandler

	conn xpc.XPC

	role int // 1: peripheralManager (server), 0: centralManager (client)

	reqc chan message
	rspc chan message

	// Only used in client/centralManager implementation
	plist   map[string]*peripheral
	plistmu *sync.Mutex

	// Only used in server/peripheralManager implementation

	attrN int
	attrs map[int]*attr

	subscribers map[string]*central
}

func NewDevice(opts ...Option) (Device, error) {
	d := &device{
		reqc:    make(chan message),
		rspc:    make(chan message),
		plist:   map[string]*peripheral{},
		plistmu: &sync.Mutex{},

		attrN: 1,
		attrs: make(map[int]*attr),

		subscribers: make(map[string]*central),
	}
	d.Option(opts...)
	d.conn = xpc.XpcConnect("com.apple.blued", d)
	return d, nil
}

func (d *device) Init(f func(Device, State)) error {
	go d.loop()
	rsp := d.sendReq(1, xpc.Dict{
		"kCBMsgArgName":    fmt.Sprintf("gopher-%v", time.Now().Unix()),
		"kCBMsgArgOptions": xpc.Dict{"kCBInitOptionShowPowerAlert": 1},
		"kCBMsgArgType":    d.role,
	})
	d.stateChanged = f
	go d.stateChanged(d, State(rsp.MustGetInt("kCBMsgArgState")))
	return nil
}

func (d *device) Advertise(a *AdvPacket) error {
	rsp := d.sendReq(8, xpc.Dict{
		"kCBAdvDataAppleMfgData": a.b, // not a.Bytes(). should be slice
	})

	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return errors.New("FIXME: Advertise error")
	}
	return nil
}

func (d *device) AdvertiseNameAndServices(name string, ss []UUID) error {
	us := uuidSlice(ss)
	rsp := d.sendReq(8, xpc.Dict{
		"kCBAdvDataLocalName":    name,
		"kCBAdvDataServiceUUIDs": us},
	)
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return errors.New("FIXME: Advertise error")
	}
	return nil
}

func (d *device) AdvertiseIBeaconData(data []byte) error {
	var utsname xpc.Utsname
	xpc.Uname(&utsname)

	var rsp xpc.Dict

	if utsname.Release >= "14." {
		l := len(data)
		buf := bytes.NewBuffer([]byte{byte(l + 5), 0xFF, 0x4C, 0x00, 0x02, byte(l)})
		buf.Write(data)
		rsp = d.sendReq(8, xpc.Dict{"kCBAdvDataAppleMfgData": buf.Bytes()})
	} else {
		rsp = d.sendReq(8, xpc.Dict{"kCBAdvDataAppleBeaconKey": data})
	}

	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return errors.New("FIXME: Advertise error")
	}

	return nil
}

func (d *device) AdvertiseIBeacon(u UUID, major, minor uint16, pwr int8) error {
	b := make([]byte, 21)
	copy(b, reverse(u.b))                     // Big endian
	binary.BigEndian.PutUint16(b[16:], major) // Big endian
	binary.BigEndian.PutUint16(b[18:], minor) // Big endian
	b[20] = uint8(pwr)                        // Measured Tx Power
	return d.AdvertiseIBeaconData(b)
}

func (d *device) StopAdvertising() error {
	rsp := d.sendReq(9, nil)
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return errors.New("FIXME: Stop Advertise error")
	}
	return nil
}

func (d *device) Stop() error {
	// No Implementation
	defer d.stateChanged(d, StatePoweredOff)
	return errors.New("FIXME: Advertise error")
}

func (d *device) RemoveAllServices() error {
	d.sendCmd(12, nil)
	return nil
}

func (d *device) AddService(s *Service) error {
	if s.uuid.Equal(attrGAPUUID) || s.uuid.Equal(attrGATTUUID) {
		// skip GATT and GAP services
		return nil
	}

	xs := xpc.Dict{
		"kCBMsgArgAttributeID":     d.attrN,
		"kCBMsgArgAttributeIDs":    []int{},
		"kCBMsgArgCharacteristics": nil,
		"kCBMsgArgType":            1, // 1 => primary, 0 => excluded
		"kCBMsgArgUUID":            reverse(s.uuid.b),
	}
	d.attrN++

	xcs := xpc.Array{}
	for _, c := range s.Characteristics() {
		props := 0
		perm := 0
		if c.props&CharRead != 0 {
			props |= 0x02
			if CharRead&c.secure != 0 {
				perm |= 0x04
			} else {
				perm |= 0x01
			}
		}
		if c.props&CharWriteNR != 0 {
			props |= 0x04
			if c.secure&CharWriteNR != 0 {
				perm |= 0x08
			} else {
				perm |= 0x02
			}
		}
		if c.props&CharWrite != 0 {
			props |= 0x08
			if c.secure&CharWrite != 0 {
				perm |= 0x08
			} else {
				perm |= 0x02
			}
		}
		if c.props&CharNotify != 0 {
			if c.secure&CharNotify != 0 {
				props |= 0x100
			} else {
				props |= 0x10
			}
		}
		if c.props&CharIndicate != 0 {
			if c.secure&CharIndicate != 0 {
				props |= 0x200
			} else {
				props |= 0x20
			}
		}

		xc := xpc.Dict{
			"kCBMsgArgAttributeID":              d.attrN,
			"kCBMsgArgUUID":                     reverse(c.uuid.b),
			"kCBMsgArgAttributePermissions":     perm,
			"kCBMsgArgCharacteristicProperties": props,
			"kCBMsgArgData":                     c.value,
		}
		d.attrs[d.attrN] = &attr{h: uint16(d.attrN), value: c.value, pvt: c}
		d.attrN++

		xds := xpc.Array{}
		for _, d := range c.Descriptors() {
			if d.uuid.Equal(attrClientCharacteristicConfigUUID) {
				// skip CCCD
				continue
			}
			var v interface{}
			if len(d.valuestr) > 0 {
				v = d.valuestr
			} else {
				v = d.value
			}
			xd := xpc.Dict{
				"kCBMsgArgData": v,
				"kCBMsgArgUUID": reverse(d.uuid.b),
			}
			xds = append(xds, xd)
		}
		xc["kCBMsgArgDescriptors"] = xds
		xcs = append(xcs, xc)
	}
	xs["kCBMsgArgCharacteristics"] = xcs

	rsp := d.sendReq(10, xs)
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return errors.New("FIXME: Add Srvice error")
	}
	return nil
}

func (d *device) SetServices(ss []*Service) error {
	d.RemoveAllServices()
	for _, s := range ss {
		d.AddService(s)
	}
	return nil
}

func (d *device) Scan(ss []UUID, dup bool) {
	args := xpc.Dict{
		"kCBMsgArgUUIDs": uuidSlice(ss),
		"kCBMsgArgOptions": xpc.Dict{
			"kCBScanOptionAllowDuplicates": map[bool]int{true: 1, false: 0}[dup],
		},
	}
	d.sendCmd(29, args)
}

func (d *device) StopScanning() {
	d.sendCmd(30, nil)
}

func (d *device) Connect(p Peripheral) {
	pp := p.(*peripheral)
	d.plist[pp.id.String()] = pp
	d.sendCmd(31,
		xpc.Dict{
			"kCBMsgArgDeviceUUID": pp.id,
			"kCBMsgArgOptions": xpc.Dict{
				"kCBConnectOptionNotifyOnDisconnection": 1,
			},
		})
}

func (d *device) respondToRequest(id int, args xpc.Dict) {

	switch id {
	case 19: // ReadRequest
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		t := args.MustGetInt("kCBMsgArgTransactionID")
		a := args.MustGetInt("kCBMsgArgAttributeID")
		o := args.MustGetInt("kCBMsgArgOffset")

		attr := d.attrs[a]
		v := attr.value
		if v == nil {
			c := newCentral(d, u)
			req := &ReadRequest{
				Request: Request{Central: c},
				Cap:     int(c.mtu - 1),
				Offset:  o,
			}
			rsp := newResponseWriter(int(c.mtu - 1))
			if c, ok := attr.pvt.(*Characteristic); ok {
				c.rhandler.ServeRead(rsp, req)
				v = rsp.bytes()
			}
		}

		d.sendCmd(13, xpc.Dict{
			"kCBMsgArgAttributeID":   a,
			"kCBMsgArgData":          v,
			"kCBMsgArgTransactionID": t,
			"kCBMsgArgResult":        0,
		})

	case 20: // WriteRequest
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		t := args.MustGetInt("kCBMsgArgTransactionID")
		a := 0
		result := byte(0)
		noRsp := false
		xxws := args.MustGetArray("kCBMsgArgATTWrites")
		for _, xxw := range xxws {
			xw := xxw.(xpc.Dict)
			if a == 0 {
				a = xw.MustGetInt("kCBMsgArgAttributeID")
			}
			o := xw.MustGetInt("kCBMsgArgOffset")
			i := xw.MustGetInt("kCBMsgArgIgnoreResponse")
			b := xw.MustGetBytes("kCBMsgArgData")
			_ = o
			attr := d.attrs[a]
			c := newCentral(d, u)
			r := Request{Central: c}
			result = attr.pvt.(*Characteristic).whandler.ServeWrite(r, b)
			if i == 1 {
				noRsp = true
			}

		}
		if noRsp {
			break
		}
		d.sendCmd(13, xpc.Dict{
			"kCBMsgArgAttributeID":   a,
			"kCBMsgArgData":          nil,
			"kCBMsgArgTransactionID": t,
			"kCBMsgArgResult":        result,
		})

	case 21: // subscribed
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		a := args.MustGetInt("kCBMsgArgAttributeID")
		attr := d.attrs[a]
		c := newCentral(d, u)
		d.subscribers[u.String()] = c
		c.startNotify(attr, c.mtu)

	case 22: // unubscribed
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		a := args.MustGetInt("kCBMsgArgAttributeID")
		attr := d.attrs[a]
		if c := d.subscribers[u.String()]; c != nil {
			c.stopNotify(attr)
		}

	case 23: // notificationSent
	}
}

func (d *device) CancelConnection(p Peripheral) {
	d.sendCmd(32, xpc.Dict{"kCBMsgArgDeviceUUID": p.(*peripheral).id})
}

// process device events and asynchronous errors
// (implements XpcEventHandler)
func (d *device) HandleXpcEvent(event xpc.Dict, err error) {
	if err != nil {
		log.Println("error:", err)
		return
	}

	id := event.MustGetInt("kCBMsgId")
	args := event.MustGetDict("kCBMsgArgs")
	//log.Printf(">> %d, %v", id, args)

	switch id {
	case // device event
		6,  // StateChanged
		16, // AdvertisingStarted
		17, // AdvertisingStopped
		18: // ServiceAdded
		d.rspc <- message{id: id, args: args}

	case
		19, // ReadRequest
		20, // WriteRequest
		21, // Subscribe
		22, // Unubscribe
		23: // Confirmation
		d.respondToRequest(id, args)

	case peripheralDiscovered:
		xa := args.MustGetDict("kCBMsgArgAdvertisementData")
		if len(xa) == 0 {
			return
		}
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		a := &Advertisement{
			LocalName:        xa.GetString("kCBAdvDataLocalName", args.GetString("kCBMsgArgName", "")),
			TxPowerLevel:     xa.GetInt("kCBAdvDataTxPowerLevel", 0),
			ManufacturerData: xa.GetBytes("kCBAdvDataManufacturerData", nil),
		}

		rssi := args.MustGetInt("kCBMsgArgRssi")

		if xu, ok := xa["kCBAdvDataServiceUUIDs"]; ok {
			for _, xs := range xu.(xpc.Array) {
				s := UUID{reverse(xs.([]byte))}
				a.Services = append(a.Services, s)
			}
		}
		if xsds, ok := xa["kCBAdvDataServiceData"]; ok {
			xsd := xsds.(xpc.Array)
			for i := 0; i < len(xsd); i += 2 {
				sd := ServiceData{
					UUID: UUID{xsd[i].([]byte)},
					Data: xsd[i+1].([]byte),
				}
				a.ServiceData = append(a.ServiceData, sd)
			}
		}
		if d.peripheralDiscovered != nil {
			go d.peripheralDiscovered(&peripheral{id: xpc.UUID(u.b), d: d}, a, rssi)
		}

	case peripheralConnected:
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		p := &peripheral{
			id:    xpc.UUID(u.b),
			d:     d,
			reqc:  make(chan message),
			rspc:  make(chan message),
			quitc: make(chan struct{}),
			sub:   newSubscriber(),
		}
		d.plistmu.Lock()
		d.plist[u.String()] = p
		d.plistmu.Unlock()
		go p.loop()

		if d.peripheralConnected != nil {
			go d.peripheralConnected(p, nil)
		}

	case peripheralDisconnected:
		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		d.plistmu.Lock()
		p := d.plist[u.String()]
		delete(d.plist, u.String())
		d.plistmu.Unlock()
		if p != nil {
			if d.peripheralDisconnected != nil {
				d.peripheralDisconnected(p, nil) // TODO: Get Result as error?
			}
			close(p.quitc)
		}

	case // Peripheral events
		rssiRead,
		serviceDiscovered,
		includedServicesDiscovered,
		characteristicsDiscovered,
		characteristicRead,
		characteristicWritten,
		notificationValueSet,
		descriptorsDiscovered,
		descriptorRead,
		descriptorWritten:

		u := UUID{args.MustGetUUID("kCBMsgArgDeviceUUID")}
		d.plistmu.Lock()
		p := d.plist[u.String()]
		d.plistmu.Unlock()
		if p != nil {
			p.rspc <- message{id: id, args: args}
		}
	default:
		//log.Printf("Unhandled event: %#v", event)
	}
}

func (d *device) sendReq(id int, args xpc.Dict) xpc.Dict {
	m := message{id: id, args: args, rspc: make(chan xpc.Dict)}
	d.reqc <- m
	return <-m.rspc
}

func (d *device) sendCmd(id int, args xpc.Dict) {
	d.reqc <- message{id: id, args: args}
}

func (d *device) loop() {
	for req := range d.reqc {
		d.sendCBMsg(req.id, req.args)
		if req.rspc == nil {
			continue
		}
		m := <-d.rspc
		req.rspc <- m.args
	}
}

func (d *device) sendCBMsg(id int, args xpc.Dict) {
	// log.Printf("<< %d, %v", id, args)
	d.conn.Send(xpc.Dict{"kCBMsgId": id, "kCBMsgArgs": args}, false)
}
