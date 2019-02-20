package hid

import (
	"fmt"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/tui"

	"github.com/dustin/go-humanize"
)

func (mod *HIDRecon) isInjecting() bool {
	return mod.inInjectMode
}

func (mod *HIDRecon) setInjectionMode(address string) error {
	if err := mod.setSniffMode(address); err != nil {
		return err
	} else if address == "clear" {
		mod.inInjectMode = false
	} else {
		mod.inInjectMode = true
	}
	return nil
}

func errNoDevice(addr string) error {
	return fmt.Errorf("HID device %s not found, make sure that hid.recon is on and that device has been discovered", addr)
}

func errNoType(addr string) error {
	return fmt.Errorf("HID frame injection requires the device type to be detected, try to 'hid.sniff %s' for a few seconds.", addr)
}

func errNotSupported(dev *network.HIDDevice) error {
	return fmt.Errorf("HID frame injection is not supported for device type %s", dev.Type.String())
}

func errNoKeyMap(layout string) error {
	return fmt.Errorf("could not find keymap for '%s' layout, supported layouts are: %s", layout, SupportedLayouts())
}

func (mod *HIDRecon) prepInjection() (error, *network.HIDDevice, []*Command) {
	// we can only inject onto visible connections
	dev, found := mod.Session.HID.Get(mod.sniffAddr)
	if found == false {
		return errNoDevice(mod.sniffAddr), nil, nil
	}

	// get the device specific protocol handler
	builder, found := FrameBuilders[dev.Type]
	if found == false {
		if dev.Type == network.HIDTypeUnknown {
			return errNoType(mod.sniffAddr), nil, nil
		}
		return errNotSupported(dev), nil, nil
	}

	// get the keymap from the selected layout
	keyMap := KeyMapFor(mod.keyLayout)
	if keyMap == nil {
		return errNoKeyMap(mod.keyLayout), nil, nil
	}

	// parse the script into a list of Command objects
	cmds, err := mod.parser.Parse(keyMap, mod.scriptPath)
	if err != nil {
		return err, nil, nil
	}

	mod.Info("%s loaded ...", mod.scriptPath)

	// build the protocol specific frames to send
	if err := builder.BuildFrames(cmds); err != nil {
		return err, nil, nil
	}

	return nil, dev, cmds
}

func (mod *HIDRecon) doInjection() {
	err, dev, cmds := mod.prepInjection()
	if err != nil {
		mod.Error("%v", err)
		return
	}

	numFrames := 0
	szFrames := 0
	for _, cmd := range cmds {
		for _, frame := range cmd.Frames {
			numFrames++
			szFrames += len(frame.Data)
		}
	}

	mod.Info("sending %d (%s) HID frames to %s (type:%s layout:%s) ...",
		numFrames,
		humanize.Bytes(uint64(szFrames)),
		tui.Bold(mod.sniffAddr),
		tui.Yellow(dev.Type.String()),
		tui.Yellow(mod.keyLayout))

	for i, cmd := range cmds {
		for j, frame := range cmd.Frames {
			if err := mod.dongle.TransmitPayload(frame.Data, 500, 3); err != nil {
				mod.Warning("error sending frame #%d of HID command #%d: %v", j, i, err)
			}

			if frame.Delay > 0 {
				mod.Debug("sleeping %dms after frame #%d of command #%d ...", frame.Delay, j, i)
				time.Sleep(frame.Delay)
			}
		}
	}
}
