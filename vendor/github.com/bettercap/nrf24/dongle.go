package nrf24

import (
	"fmt"
	"github.com/google/gousb"
)

type Dongle struct {
	ctx           *gousb.Context
	dev           *gousb.Device
	iface         *gousb.Interface
	ifaceDoneFunc func()
	writer        *gousb.OutEndpoint
	reader        *gousb.InEndpoint
}

func Open() (dongle *Dongle, err error) {
	dongle = &Dongle{ctx: gousb.NewContext()}

	if dongle.dev, err = dongle.ctx.OpenDeviceWithVIDPID(VendorID, ProductID); dongle.dev == nil {
		err = fmt.Errorf("usb device %s:%s not found", VendorID, ProductID)
	}
	if err != nil {
		dongle.Close()
		return nil, err
	}

	if dongle.iface, dongle.ifaceDoneFunc, err = dongle.dev.DefaultInterface(); err != nil {
		dongle.Close()
		return nil, err
	} else if dongle.writer, err = dongle.iface.OutEndpoint(1); err != nil {
		dongle.Close()
		return nil, err
	} else if dongle.reader, err = dongle.iface.InEndpoint(0x81); err != nil {
		dongle.Close()
		return nil, err
	}

	return dongle, nil
}

func (d *Dongle) String() string {
	return d.iface.String()
}

func (d *Dongle) Command(cmd Command, data []byte) (int, error) {
	return d.writer.Write(append([]byte{byte(cmd)}, data...))
}

func (d *Dongle) Read() (int, []byte, error) {
	buf := make([]byte, PacketSize)
	read, err := d.reader.Read(buf)
	return read, buf, err
}

func (d *Dongle) consumePacket() error {
	_, _, err := d.Read()
	return err
}

func (d *Dongle) EnterPromiscModeFor(prefix []byte) error {
	prData := []byte{0}
	if prefix != nil {
		prData = append([]byte{byte(len(prefix) & 0xff)}, prefix...)
	}

	if _, err := d.Command(CmdEnterPromiscMode, prData); err != nil {
		return err
	}

	return d.consumePacket()
}

func (d *Dongle) EnterPromiscMode() error {
	return d.EnterPromiscModeFor(nil)
}

func (d *Dongle) EnterPromiscModeGenericFor(prefix []byte, rate RfRate, payloadLength int) error {
	prData := []byte{0, byte(rate & 0xff), byte(payloadLength & 0xff)}
	if prefix != nil {
		prData[0] = byte(len(prefix) & 0xff)
		prData = append(prData, prefix...)
	}

	if _, err := d.Command(CmdEnterPromiscModeGeneric, prData); err != nil {
		return err
	}
	return d.consumePacket()
}

func (d *Dongle) EnterPromiscModeGenericDefaultFor(prefix []byte) error {
	return d.EnterPromiscModeGenericFor(prefix, RfRate2M, 32)
}

func (d *Dongle) EnterPromiscModeGeneric() error {
	return d.EnterPromiscModeGenericDefaultFor(nil)
}

func (d *Dongle) EnterSnifferModeFor(address []byte) error {
	adData := []byte{0}
	if address != nil {
		adData = append([]byte{byte(len(address) & 0xff)}, address...)
	}

	if _, err := d.Command(CmdEnterSnifferMode, adData); err != nil {
		return err
	}
	return d.consumePacket()
}

func (d *Dongle) EnterSnifferMode() error {
	return d.EnterSnifferModeFor(nil)
}

func (d *Dongle) EnterToneTestMode() error {
	if _, err := d.Command(CmdEnterToneTestMode, []byte{}); err != nil {
		return err
	}
	return d.consumePacket()
}

func (d *Dongle) ReceivePayload() ([]byte, error) {
	if _, err := d.Command(CmdReceivePayload, []byte{}); err != nil {
		return nil, err
	} else if read, buf, err := d.Read(); err != nil {
		return nil, err
	} else {
		return buf[:read], nil
	}
}

func (d *Dongle) checkResponseCode() error {
	if read, buf, err := d.Read(); err != nil {
		return err
	} else if read < 1 {
		return fmt.Errorf("invalid read value %d", read)
	} else if buf[0] == 0 {
		return fmt.Errorf("unexpected response %x", buf[:read])
	}
	return nil
}

func (d *Dongle) TransmitPayloadGeneric(payload []byte, address []byte) error {
	data := []byte{
		byte(len(payload) & 0xff),
		byte(len(address) & 0xff),
	}
	data = append(data, payload...)
	data = append(data, address...)

	if _, err := d.Command(CmdTransmitPayloadGeneric, data); err != nil {
		return err
	}
	return d.checkResponseCode()
}

func (d *Dongle) TransmitPayload(payload []byte, timeout int, retransmits int) error {
	data := []byte{
		byte(len(payload) & 0xff),
		byte(timeout & 0xff),
		byte(retransmits & 0xff),
	}
	data = append(data, payload...)

	if _, err := d.Command(CmdTransmitPayload, data); err != nil {
		return err
	}
	return d.checkResponseCode()
}

func (d *Dongle) TransmitAckPayload(payload []byte) error {
	data := append([]byte{byte(len(payload) & 0xff)}, payload...)
	if _, err := d.Command(CmdTransmitAckPayload, data); err != nil {
		return err
	}
	return d.checkResponseCode()
}

func (d *Dongle) SetChannel(ch int) error {
	if ch > MaxChannel {
		ch = MaxChannel
	}
	if _, err := d.Command(CmdSetChannel, []byte{byte(ch & 0xff)}); err != nil {
		return err
	}
	return d.consumePacket()
}

func (d *Dongle) GetChannel() (int, error) {
	if _, err := d.Command(CmdGetChannel, []byte{}); err != nil {
		return 0, err
	} else if read, buf, err := d.Read(); err != nil {
		return 0, err
	} else if read < 1 {
		return 0, fmt.Errorf("invalid read value %d", read)
	} else {
		return int(buf[0]), nil
	}
}

func (d *Dongle) EnableLNA() error {
	if _, err := d.Command(CmdEnableLNAPA, []byte{}); err != nil {
		return err
	}
	return d.consumePacket()
}

func (d *Dongle) Close() {
	if d.dev != nil {
		d.dev.Close()
		d.dev = nil
	}

	if d.iface != nil {
		d.ifaceDoneFunc()
		d.iface.Close()
		d.iface = nil
	}

	if d.ctx != nil {
		d.ctx.Close()
		d.ctx = nil
	}
}
