package linux

import (
	"fmt"
	"io"
	"log"

	"github.com/bettercap/gatt/linux/cmd"
)

type aclData struct {
	attr  uint16
	flags uint8
	dlen  uint16
	b     []byte
}

func (a *aclData) unmarshal(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("malformed acl packet")
	}
	attr := uint16(b[0]) | (uint16(b[1]&0x0f) << 8)
	flags := b[1] >> 4
	dlen := uint16(b[2]) | (uint16(b[3]) << 8)
	if len(b) != 4+int(dlen) {
		return fmt.Errorf("malformed acl packet")
	}

	*a = aclData{attr: attr, flags: flags, dlen: dlen, b: b[4:]}
	return nil
}

type conn struct {
	hci   *HCI
	attr  uint16
	aclc  chan *aclData
	datac chan []byte
}

func newConn(hci *HCI, hh uint16) *conn {
	c := &conn{
		hci:   hci,
		attr:  hh,
		aclc:  make(chan *aclData),
		datac: make(chan []byte, 32),
	}
	go c.loop()
	return c
}

func (c *conn) loop() {
	defer close(c.datac)
	for a := range c.aclc {
		if len(a.b) < 4 {
			log.Printf("l2conn: short/corrupt packet, %v [% X]", a, a.b)
			return
		}
		cid := uint16(a.b[2]) | (uint16(a.b[3]) << 8)
		if cid == 5 {
			c.handleSignal(a)
			continue
		}
		b := make([]byte, 512)
		tlen := int(uint16(a.b[0]) | uint16(a.b[1])<<8)
		d := a.b[4:] // skip l2cap header
		copy(b, d)
		n := len(d)

		// Keep receiving and reassemble continued l2cap segments
		for n != tlen {
			a, ok := <-c.aclc
			if !ok || (a.flags&0x1) == 0 {
				return
			}
			n += copy(b[n:], a.b)
		}
		c.datac <- b[:n]
	}
}

func (c *conn) updateConnection() (int, error) {
	b := []byte{
		0x12,       // Code (Connection Param Update)
		0x02,       // ID
		0x08, 0x00, // DataLength
		0x08, 0x00, // IntervalMin
		0x18, 0x00, // IntervalMax
		0x00, 0x00, // SlaveLatency
		0xC8, 0x00} // TimeoutMultiplier
	return c.write(0x05, b)
}

// write writes the l2cap payload to the controller.
// It first prepend the l2cap header (4-bytes), and diassemble the payload
// if it is larger than the HCI LE buffer size that the conntroller can support.
func (c *conn) write(cid int, b []byte) (int, error) {
	flag := uint8(0) // ACL data continuation flag
	tlen := len(b)   // Total length of the l2cap payload

	logger.Info("l2cap", "W", fmt.Sprintf("[% X]", b))
	w := append(
		[]byte{
			0,    // packet type
			0, 0, // attr
			0, 0, // dlen
			uint8(tlen), uint8(tlen >> 8), // l2cap header
			uint8(cid), uint8(cid >> 8), // l2cap header
		}, b...)

	n := 4 + tlen // l2cap header + l2cap payload
	for n > 0 {
		dlen := n
		if dlen > c.hci.bufSize {
			dlen = c.hci.bufSize
		}
		w[0] = 0x02 // packetTypeACL
		w[1] = uint8(c.attr)
		w[2] = uint8(c.attr>>8) | flag
		w[3] = uint8(dlen)
		w[4] = uint8(dlen >> 8)

		// make sure we don't send more buffers than the controller can handdle
		c.hci.bufCnt <- struct{}{}

		c.hci.d.Write(w[:5+dlen])
		w = w[dlen:] // advance the pointer to the next segment, if any.
		flag = 0x10  // the rest of iterations attr continued segments, if any.
		n -= dlen
	}

	return len(b), nil
}

func (c *conn) Read(b []byte) (int, error) {
	d, ok := <-c.datac
	if !ok {
		return 0, io.EOF
	}
	if len(d) > len(b) {
		return copy(b, d), io.ErrShortBuffer
	}
	logger.Info("l2cap", "R", fmt.Sprintf("[% X]", d))
	n := copy(b, d)
	return n, nil
}

func (c *conn) Write(b []byte) (int, error) {
	return c.write(0x04, b)
}

// Close disconnects the connection by sending HCI disconnect command to the device.
func (c *conn) Close() error {
	h := c.hci
	hh := c.attr
	h.connsmu.Lock()
	defer h.connsmu.Unlock()
	_, found := h.conns[hh]
	if !found {
		log.Printf("l2conn: 0x%04x already disconnected", hh)
		return nil
	}
	if err, _ := h.c.Send(cmd.Disconnect{ConnectionHandle: hh, Reason: 0x13}); err != nil {
		return fmt.Errorf("l2conn: failed to disconnect, %s", err)
	}
	return nil
}

// Signal Packets
// 0x00 Reserved								Any
// 0x01 Command reject							0x0001 and 0x0005
// 0x02 Connection request						0x0001
// 0x03 Connection response 					0x0001
// 0x04 Configure request						0x0001
// 0x05 Configure response						0x0001
// 0x06 Disconnection request					0x0001 and 0x0005
// 0x07 Disconnection response					0x0001 and 0x0005
// 0x08 Echo request							0x0001
// 0x09 Echo response							0x0001
// 0x0A Information request						0x0001
// 0x0B Information response					0x0001
// 0x0C Create Channel request					0x0001
// 0x0D Create Channel response					0x0001
// 0x0E Move Channel request					0x0001
// 0x0F Move Channel response					0x0001
// 0x10 Move Channel Confirmation				0x0001
// 0x11 Move Channel Confirmation response		0x0001
// 0x12 Connection Parameter Update request		0x0005
// 0x13 Connection Parameter Update response	0x0005
// 0x14 LE Credit Based Connection request		0x0005
// 0x15 LE Credit Based Connection response		0x0005
// 0x16 LE Flow Control Credit					0x0005
func (c *conn) handleSignal(a *aclData) error {
	log.Printf("ignore l2cap signal:[ % X ]", a.b)
	// FIXME: handle LE signaling channel (CID: 5)
	return nil
}
