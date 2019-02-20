package hid_recon

import (
	"github.com/bettercap/bettercap/network"
)

type FrameBuilder interface {
	BuildFrames(commands []Command)
}

var FrameBuilders = map[network.HIDType]FrameBuilder{
	network.HIDTypeLogitech: LogitechBuilder{},
}
