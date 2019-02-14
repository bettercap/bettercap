package gatt

import (
	"errors"
	"log"

	"github.com/bettercap/gatt/xpc"
)

type peripheral struct {
	// NameChanged is called whenever the peripheral GAP Device name has changed.
	NameChanged func(Peripheral)

	// ServicesModified is called when one or more service of a peripheral have changed.
	// A list of invalid service is provided in the parameter.
	ServicesModified func(Peripheral, []*Service)

	d    *device
	svcs []*Service

	sub *subscriber

	id   xpc.UUID
	name string

	reqc  chan message
	rspc  chan message
	quitc chan struct{}
}

func NewPeripheral(u UUID) Peripheral { return &peripheral{id: xpc.UUID(u.b)} }

func (p *peripheral) Device() Device       { return p.d }
func (p *peripheral) ID() string           { return p.id.String() }
func (p *peripheral) Name() string         { return p.name }
func (p *peripheral) Services() []*Service { return p.svcs }

func (p *peripheral) DiscoverServices(ss []UUID) ([]*Service, error) {
	rsp := p.sendReq(45, xpc.Dict{
		"kCBMsgArgDeviceUUID": p.id,
		"kCBMsgArgUUIDs":      uuidSlice(ss),
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return nil, AttEcode(res)
	}
	svcs := []*Service{}
	for _, xss := range rsp["kCBMsgArgServices"].(xpc.Array) {
		xs := xss.(xpc.Dict)
		u := MustParseUUID(xs.MustGetHexBytes("kCBMsgArgUUID"))
		h := uint16(xs.MustGetInt("kCBMsgArgServiceStartHandle"))
		endh := uint16(xs.MustGetInt("kCBMsgArgServiceEndHandle"))
		svcs = append(svcs, &Service{uuid: u, h: h, endh: endh})
	}
	p.svcs = svcs
	return svcs, nil
}

func (p *peripheral) DiscoverIncludedServices(ss []UUID, s *Service) ([]*Service, error) {
	rsp := p.sendReq(60, xpc.Dict{
		"kCBMsgArgDeviceUUID":         p.id,
		"kCBMsgArgServiceStartHandle": s.h,
		"kCBMsgArgServiceEndHandle":   s.endh,
		"kCBMsgArgUUIDs":              uuidSlice(ss),
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return nil, AttEcode(res)
	}
	// TODO
	return nil, notImplemented
}

func (p *peripheral) DiscoverCharacteristics(cs []UUID, s *Service) ([]*Characteristic, error) {
	rsp := p.sendReq(62, xpc.Dict{
		"kCBMsgArgDeviceUUID":         p.id,
		"kCBMsgArgServiceStartHandle": s.h,
		"kCBMsgArgServiceEndHandle":   s.endh,
		"kCBMsgArgUUIDs":              uuidSlice(cs),
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return nil, AttEcode(res)
	}
	for _, xcs := range rsp.MustGetArray("kCBMsgArgCharacteristics") {
		xc := xcs.(xpc.Dict)
		u := MustParseUUID(xc.MustGetHexBytes("kCBMsgArgUUID"))
		ch := uint16(xc.MustGetInt("kCBMsgArgCharacteristicHandle"))
		vh := uint16(xc.MustGetInt("kCBMsgArgCharacteristicValueHandle"))
		props := Property(xc.MustGetInt("kCBMsgArgCharacteristicProperties"))
		c := &Characteristic{uuid: u, svc: s, props: props, h: ch, vh: vh}
		s.chars = append(s.chars, c)
	}
	return s.chars, nil
}

func (p *peripheral) DiscoverDescriptors(ds []UUID, c *Characteristic) ([]*Descriptor, error) {
	rsp := p.sendReq(70, xpc.Dict{
		"kCBMsgArgDeviceUUID":                p.id,
		"kCBMsgArgCharacteristicHandle":      c.h,
		"kCBMsgArgCharacteristicValueHandle": c.vh,
		"kCBMsgArgUUIDs":                     uuidSlice(ds),
	})
	for _, xds := range rsp.MustGetArray("kCBMsgArgDescriptors") {
		xd := xds.(xpc.Dict)
		u := MustParseUUID(xd.MustGetHexBytes("kCBMsgArgUUID"))
		h := uint16(xd.MustGetInt("kCBMsgArgDescriptorHandle"))
		d := &Descriptor{uuid: u, char: c, h: h}
		c.descs = append(c.descs, d)
	}
	return c.descs, nil
}

func (p *peripheral) ReadCharacteristic(c *Characteristic) ([]byte, error) {
	rsp := p.sendReq(65, xpc.Dict{
		"kCBMsgArgDeviceUUID":                p.id,
		"kCBMsgArgCharacteristicHandle":      c.h,
		"kCBMsgArgCharacteristicValueHandle": c.vh,
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return nil, AttEcode(res)
	}
	b := rsp.MustGetBytes("kCBMsgArgData")
	return b, nil
}

func (p *peripheral) ReadLongCharacteristic(c *Characteristic) ([]byte, error) {
	return nil, errors.New("Not implemented")
}

func (p *peripheral) WriteCharacteristic(c *Characteristic, b []byte, noRsp bool) error {
	args := xpc.Dict{
		"kCBMsgArgDeviceUUID":                p.id,
		"kCBMsgArgCharacteristicHandle":      c.h,
		"kCBMsgArgCharacteristicValueHandle": c.vh,
		"kCBMsgArgData":                      b,
		"kCBMsgArgType":                      map[bool]int{false: 0, true: 1}[noRsp],
	}
	if noRsp {
		p.sendCmd(66, args)
		return nil
	}
	rsp := p.sendReq(66, args)
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return AttEcode(res)
	}
	return nil
}

func (p *peripheral) ReadDescriptor(d *Descriptor) ([]byte, error) {
	rsp := p.sendReq(77, xpc.Dict{
		"kCBMsgArgDeviceUUID":       p.id,
		"kCBMsgArgDescriptorHandle": d.h,
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return nil, AttEcode(res)
	}
	b := rsp.MustGetBytes("kCBMsgArgData")
	return b, nil
}

func (p *peripheral) WriteDescriptor(d *Descriptor, b []byte) error {
	rsp := p.sendReq(78, xpc.Dict{
		"kCBMsgArgDeviceUUID":       p.id,
		"kCBMsgArgDescriptorHandle": d.h,
		"kCBMsgArgData":             b,
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return AttEcode(res)
	}
	return nil
}

func (p *peripheral) SetNotifyValue(c *Characteristic, f func(*Characteristic, []byte, error)) error {
	set := 1
	if f == nil {
		set = 0
	}
	// To avoid race condition, registeration is handled before requesting the server.
	if f != nil {
		// Note: when notified, core bluetooth reports characteristic handle, not value's handle.
		p.sub.subscribe(c.h, func(b []byte, err error) { f(c, b, err) })
	}
	rsp := p.sendReq(68, xpc.Dict{
		"kCBMsgArgDeviceUUID":                p.id,
		"kCBMsgArgCharacteristicHandle":      c.h,
		"kCBMsgArgCharacteristicValueHandle": c.vh,
		"kCBMsgArgState":                     set,
	})
	if res := rsp.MustGetInt("kCBMsgArgResult"); res != 0 {
		return AttEcode(res)
	}
	// To avoid race condition, unregisteration is handled after server responses.
	if f == nil {
		p.sub.unsubscribe(c.h)
	}
	return nil
}

func (p *peripheral) SetIndicateValue(c *Characteristic,
	f func(*Characteristic, []byte, error)) error {
	// TODO: Implement set indications logic for darwin (https://github.com/paypal/gatt/issues/32)
	return nil
}

func (p *peripheral) ReadRSSI() int {
	rsp := p.sendReq(43, xpc.Dict{"kCBMsgArgDeviceUUID": p.id})
	return rsp.MustGetInt("kCBMsgArgData")
}

func (p *peripheral) SetMTU(mtu uint16) error {
	return errors.New("Not implemented")
}

func uuidSlice(uu []UUID) [][]byte {
	us := [][]byte{}
	for _, u := range uu {
		us = append(us, reverse(u.b))
	}
	return us
}

type message struct {
	id   int
	args xpc.Dict
	rspc chan xpc.Dict
}

func (p *peripheral) sendCmd(id int, args xpc.Dict) {
	p.reqc <- message{id: id, args: args}
}

func (p *peripheral) sendReq(id int, args xpc.Dict) xpc.Dict {
	m := message{id: id, args: args, rspc: make(chan xpc.Dict)}
	p.reqc <- m
	return <-m.rspc
}

func (p *peripheral) loop() {
	rspc := make(chan message)

	go func() {
		for {
			select {
			case req := <-p.reqc:
				p.d.sendCBMsg(req.id, req.args)
				if req.rspc == nil {
					break
				}
				m := <-rspc
				req.rspc <- m.args
			case <-p.quitc:
				return
			}
		}
	}()

	for {
		select {
		case rsp := <-p.rspc:
			// Notification
			if rsp.id == 71 && rsp.args.GetInt("kCBMsgArgIsNotification", 0) != 0 {
				// While we're notified with the value's handle, blued reports the characteristic handle.
				ch := uint16(rsp.args.MustGetInt("kCBMsgArgCharacteristicHandle"))
				b := rsp.args.MustGetBytes("kCBMsgArgData")
				f := p.sub.fn(ch)
				if f == nil {
					log.Printf("notified by unsubscribed handle")
					// FIXME: should terminate the connection?
				} else {
					go f(b, nil)
				}
				break
			}
			rspc <- rsp
		case <-p.quitc:
			return
		}
	}
}
