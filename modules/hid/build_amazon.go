package hid

import (
	"github.com/bettercap/bettercap/network"
)

const (
	amzFrameDelay = 5
)

type AmazonBuilder struct {
}

func (b AmazonBuilder) frameFor(cmd *Command) []byte {
	return []byte{0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
		0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
		0x0f, 0x0f, 0x0f, 0x0f, 0x0f, 0x0f,
		0x0f, 0, cmd.Mode, 0, cmd.HID, 0}
}

func (b AmazonBuilder) BuildFrames(dev *network.HIDDevice, commands []*Command) error {
	for i, cmd := range commands {
		if i == 0 {
			for j := 0; j < 5; j++ {
				cmd.AddFrame(b.frameFor(&Command{}), amzFrameDelay)
			}
		}

		if cmd.IsHID() {
			cmd.AddFrame(b.frameFor(cmd), amzFrameDelay)
			cmd.AddFrame(b.frameFor(&Command{}), amzFrameDelay)
		} else if cmd.IsSleep() {
			for i, num := 0, cmd.Sleep/10; i < num; i++ {
				cmd.AddFrame(b.frameFor(&Command{}), 10)
			}
		}
	}

	return nil
}
