package gatt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/bettercap/gatt/linux"
)

type peripheral struct {
	// NameChanged is called whenever the peripheral GAP device name has changed.
	NameChanged func(*peripheral)

	// ServicedModified is called when one or more service of a peripheral have changed.
	// A list of invalid service is provided in the parameter.
	ServicesModified func(*peripheral, []*Service)

	d    *device
	svcs []*Service

	sub *subscriber

	mtu uint16
	l2c io.ReadWriteCloser

	reqc  chan message
	quitc chan struct{}

	pd *linux.PlatData // platform specific data
}

func (p *peripheral) Device() Device       { return p.d }
func (p *peripheral) ID() string           { return strings.ToUpper(net.HardwareAddr(p.pd.Address[:]).String()) }
func (p *peripheral) Name() string         { return p.pd.Name }
func (p *peripheral) Services() []*Service { return p.svcs }

func finish(op byte, h uint16, b []byte) (bool, error) {
	done := b[0] == attOpError && b[1] == op && b[2] == byte(h) && b[3] == byte(h>>8)
	var err error
	if b[0] == attOpError {
		err = AttEcode(b[4])
		if err == AttEcodeAttrNotFound {
			// Expect attribute not found errors
			err = nil
		} else {
			// log.Printf("unexpected protocol error: %s", e)
			// FIXME: terminate the connection
		}
	}
	return done, err
}

func (p *peripheral) DiscoverServices(ds []UUID) ([]*Service, error) {
	// p.pd.Conn.Write([]byte{0x02, 0x87, 0x00}) // MTU
	done := false
	start := uint16(0x0001)
	var err error
	for !done {
		op := byte(attOpReadByGroupReq)
		b := make([]byte, 7)
		b[0] = op
		binary.LittleEndian.PutUint16(b[1:3], start)
		binary.LittleEndian.PutUint16(b[3:5], 0xFFFF)
		binary.LittleEndian.PutUint16(b[5:7], 0x2800)

		b = p.sendReq(op, b)
		done, err = finish(op, start, b)
		if done {
			break
		}
		b = b[1:]
		l, b := int(b[0]), b[1:]
		switch {
		case l == 6 && (len(b)%6 == 0):
		case l == 20 && (len(b)%20 == 0):
		default:
			return nil, ErrInvalidLength
		}

		for len(b) != 0 {
			endh := binary.LittleEndian.Uint16(b[2:4])
			u := UUID{b[4:l]}

			if UUIDContains(ds, u) {
				s := &Service{
					uuid: u,
					h:    binary.LittleEndian.Uint16(b[:2]),
					endh: endh,
				}
				p.svcs = append(p.svcs, s)
			}

			b = b[l:]
			done = endh == 0xFFFF
			start = endh + 1
		}
	}
	return p.svcs, err
}

func (p *peripheral) DiscoverIncludedServices(ss []UUID, s *Service) ([]*Service, error) {
	// TODO
	return nil, nil
}

func (p *peripheral) DiscoverCharacteristics(cs []UUID, s *Service) ([]*Characteristic, error) {
	done := false
	start := s.h
	var prev *Characteristic
	var err error
	for !done {
		op := byte(attOpReadByTypeReq)
		b := make([]byte, 7)
		b[0] = op
		binary.LittleEndian.PutUint16(b[1:3], start)
		binary.LittleEndian.PutUint16(b[3:5], s.endh)
		binary.LittleEndian.PutUint16(b[5:7], 0x2803)

		b = p.sendReq(op, b)
		if done = b[0] != byte(attOpReadByTypeRsp); done {
			break
		}

		b = b[1:]
		l, b := int(b[0]), b[1:]
		switch {
		case l == 7 && (len(b)%7 == 0):
		case l == 21 && (len(b)%21 == 0):
		default:
			return nil, ErrInvalidLength
		}

		for len(b) != 0 {
			h := binary.LittleEndian.Uint16(b[:2])
			props := Property(b[2])
			vh := binary.LittleEndian.Uint16(b[3:5])
			u := UUID{b[5:l]}
			s := searchService(p.svcs, h, vh)
			if s == nil {
				log.Printf("Can't find service range that contains 0x%04X - 0x%04X", h, vh)
				return nil, fmt.Errorf("Can't find service range that contains 0x%04X - 0x%04X", h, vh)
			}
			c := &Characteristic{
				uuid:  u,
				svc:   s,
				props: props,
				h:     h,
				vh:    vh,
			}
			if UUIDContains(cs, u) {
				s.chars = append(s.chars, c)
			}
			b = b[l:]
			done = vh == s.endh
			start = vh + 1
			if prev != nil {
				prev.endh = c.h - 1
			}
			prev = c
		}
	}
	if len(s.chars) > 1 {
		s.chars[len(s.chars)-1].endh = s.endh
	}
	return s.chars, err
}

func (p *peripheral) DiscoverDescriptors(ds []UUID, c *Characteristic) ([]*Descriptor, error) {
	done := false
	start := c.vh + 1
	var err error
	for !done {
		if c.endh == 0 {
			c.endh = c.svc.endh
		}
		op := byte(attOpFindInfoReq)
		b := make([]byte, 5)
		b[0] = op
		binary.LittleEndian.PutUint16(b[1:3], start)
		binary.LittleEndian.PutUint16(b[3:5], c.endh)

		b = p.sendReq(op, b)
		done, err = finish(op, start, b)
		if done {
			break
		}
		b = b[1:]

		var l int
		f, b := int(b[0]), b[1:]
		switch {
		case f == 1 && (len(b)%4 == 0):
			l = 4
		case f == 2 && (len(b)%18 == 0):
			l = 18
		default:
			return nil, ErrInvalidLength
		}

		for len(b) != 0 {
			h := binary.LittleEndian.Uint16(b[:2])
			u := UUID{b[2:l]}
			d := &Descriptor{uuid: u, h: h, char: c}
			if UUIDContains(ds, u) {
				c.descs = append(c.descs, d)
			}
			if u.Equal(attrClientCharacteristicConfigUUID) {
				c.cccd = d
			}
			b = b[l:]
			done = h == c.endh
			start = h + 1
		}
	}
	return c.descs, err
}

func (p *peripheral) ReadCharacteristic(c *Characteristic) ([]byte, error) {
	b := make([]byte, 3)
	op := byte(attOpReadReq)
	b[0] = op
	binary.LittleEndian.PutUint16(b[1:3], c.vh)

	b = p.sendReq(op, b)
	_, err := finish(op, c.vh, b)
	b = b[1:]
	return b, err
}

func (p *peripheral) ReadLongCharacteristic(c *Characteristic) ([]byte, error) {
	// The spec says that a read blob request should fail if the characteristic
	// is smaller than mtu - 1.  To simplify the API, the first read is done
	// with a regular read request.  If the buffer received is equal to mtu -1,
	// then we read the rest of the data using read blob.
	firstRead, err := p.ReadCharacteristic(c)
	if err != nil {
		return nil, err
	}
	if len(firstRead) < int(p.mtu)-1 {
		return firstRead, nil
	}

	var buf bytes.Buffer
	buf.Write(firstRead)
	off := uint16(len(firstRead))
	done := false
	err = AttEcodeSuccess
	for {
		b := make([]byte, 5)
		op := byte(attOpReadBlobReq)
		b[0] = op
		binary.LittleEndian.PutUint16(b[1:3], c.vh)
		binary.LittleEndian.PutUint16(b[3:5], off)

		b = p.sendReq(op, b)
		done, err = finish(op, c.vh, b)
		if done {
			break
		}
		b = b[1:]
		if len(b) == 0 {
			break
		}
		buf.Write(b)
		off += uint16(len(b))
		if len(b) < int(p.mtu)-1 {
			break
		}
	}
	return buf.Bytes(), err
}

func (p *peripheral) WriteCharacteristic(c *Characteristic, value []byte, noRsp bool) error {
	b := make([]byte, 3+len(value))
	op := byte(attOpWriteReq)
	b[0] = op
	if noRsp {
		b[0] = attOpWriteCmd
	}
	binary.LittleEndian.PutUint16(b[1:3], c.vh)
	copy(b[3:], value)

	if noRsp {
		p.sendCmd(op, b)
		return nil
	}
	b = p.sendReq(op, b)
	_, err := finish(op, c.vh, b)
	b = b[1:]
	return err
}

func (p *peripheral) ReadDescriptor(d *Descriptor) ([]byte, error) {
	b := make([]byte, 3)
	op := byte(attOpReadReq)
	b[0] = op
	binary.LittleEndian.PutUint16(b[1:3], d.h)

	b = p.sendReq(op, b)
	_, err := finish(op, d.h, b)
	b = b[1:]
	return b, err
}

func (p *peripheral) WriteDescriptor(d *Descriptor, value []byte) error {
	b := make([]byte, 3+len(value))
	op := byte(attOpWriteReq)
	b[0] = op
	binary.LittleEndian.PutUint16(b[1:3], d.h)
	copy(b[3:], value)

	b = p.sendReq(op, b)
	_, err := finish(op, d.h, b)
	b = b[1:]
	return err
}

func (p *peripheral) setNotifyValue(c *Characteristic, flag uint16,
	f func(*Characteristic, []byte, error)) error {
	if c.cccd == nil {
		return errors.New("no cccd") // FIXME
	}
	ccc := uint16(0)
	if f != nil {
		ccc = flag
		p.sub.subscribe(c.vh, func(b []byte, err error) { f(c, b, err) })
	}
	b := make([]byte, 5)
	op := byte(attOpWriteReq)
	b[0] = op
	binary.LittleEndian.PutUint16(b[1:3], c.cccd.h)
	binary.LittleEndian.PutUint16(b[3:5], ccc)

	b = p.sendReq(op, b)
	_, err := finish(op, c.cccd.h, b)
	b = b[1:]
	if f == nil {
		p.sub.unsubscribe(c.vh)
	}
	return err
}

func (p *peripheral) SetNotifyValue(c *Characteristic,
	f func(*Characteristic, []byte, error)) error {
	return p.setNotifyValue(c, gattCCCNotifyFlag, f)
}

func (p *peripheral) SetIndicateValue(c *Characteristic,
	f func(*Characteristic, []byte, error)) error {
	return p.setNotifyValue(c, gattCCCIndicateFlag, f)
}

func (p *peripheral) ReadRSSI() int {
	// TODO: implement
	return -1
}

func searchService(ss []*Service, start, end uint16) *Service {
	for _, s := range ss {
		if s.h < start && s.endh >= end {
			return s
		}
	}
	return nil
}

// TODO: unifiy the message with OS X pots and refactor
type message struct {
	op   byte
	b    []byte
	rspc chan []byte
}

func (p *peripheral) sendCmd(op byte, b []byte) {
	p.reqc <- message{op: op, b: b}
}

func (p *peripheral) sendReq(op byte, b []byte) []byte {
	m := message{op: op, b: b, rspc: make(chan []byte)}
	p.reqc <- m
	return <-m.rspc
}

func (p *peripheral) loop() {
	// Serialize the request.
	rspc := make(chan []byte)

	// Dequeue request loop
	go func() {
		for {
			select {
			case req := <-p.reqc:
				p.l2c.Write(req.b)
				if req.rspc == nil {
					break
				}

				for {
					r := <-rspc
					reqOp, rspOp := req.b[0], r[0]
					if rspOp == attRspFor[reqOp] || (rspOp == attOpError && r[1] == reqOp) {
						req.rspc <- r
						break
					}
					log.Printf("Request 0x%02x got a mismatched response: 0x%02x", reqOp, rspOp)
					p.l2c.Write(attErrorRsp(rspOp, 0x0000, AttEcodeReqNotSupp))
				}
			case <-p.quitc:
				return
			}
		}
	}()

	// L2CAP implementations shall support a minimum MTU size of 48 bytes.
	// The default value is 672 bytes
	buf := make([]byte, 672)

	// Handling response or notification/indication
	for {
		n, err := p.l2c.Read(buf)
		if n == 0 || err != nil {
			close(p.quitc)
			return
		}

		b := make([]byte, n)
		copy(b, buf)

		if (b[0] != attOpHandleNotify) && (b[0] != attOpHandleInd) {
			log.Printf("response 0x%x", b[0])
			rspc <- b
			continue
		}

		h := binary.LittleEndian.Uint16(b[1:3])
		f := p.sub.fn(h)
		if f == nil {
			log.Printf("notified by unsubscribed handle")
			// FIXME: terminate the connection?
		} else {
			go f(b[3:], nil)
		}

		if b[0] == attOpHandleInd {
			// write aknowledgement for indication
			p.l2c.Write([]byte{attOpHandleCnf})
		}

	}
}

func (p *peripheral) SetMTU(mtu uint16) error {
	b := make([]byte, 3)
	op := byte(attOpMtuReq)
	b[0] = op
	h := uint16(mtu)
	binary.LittleEndian.PutUint16(b[1:3], h)

	b = p.sendReq(op, b)
	done, err := finish(op, h, b)
	if !done {
		serverMTU := binary.LittleEndian.Uint16(b[1:3])
		if serverMTU < mtu {
			mtu = serverMTU
		}
		p.mtu = mtu
	}
	return err
}
