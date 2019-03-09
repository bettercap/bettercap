package hid

import (
	"github.com/bettercap/bettercap/network"
)

type FrameBuilder interface {
	BuildFrames(*network.HIDDevice, []*Command) error
}

var FrameBuilders = map[network.HIDType]FrameBuilder{
	network.HIDTypeLogitech:  LogitechBuilder{},
	network.HIDTypeAmazon:    AmazonBuilder{},
	network.HIDTypeMicrosoft: MicrosoftBuilder{},
}

func availBuilders() []string {
	return []string{
		"logitech",
		"amazon",
		"microsoft",
	}
}

func builderFromName(name string) FrameBuilder {
	switch name {
	case "amazon":
		return AmazonBuilder{}
	case "microsoft":
		return MicrosoftBuilder{}
	default:
		return LogitechBuilder{}
	}
}
