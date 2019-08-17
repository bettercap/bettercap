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
	if err := mod.setSniffMode(address, true); err != nil {
		return err
	} else if address == "clear" {
		mod.inInjectMode = false
	} else {
		mod.inInjectMode = true
	}
	return nil
}

func errNoDevice(addr string) error {
	return fmt.Errorf("HID device %s not found, make sure that hid.recon is on and that this device has been discovered", addr)
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
	var err error

	if err, mod.sniffType = mod.StringParam("hid.force.type"); err != nil {
		return err, nil, nil
	}

	dev, found := mod.Session.HID.Get(mod.sniffAddr)
	if found == false {
		mod.Warning("device %s is not visible, will use HID type %s", mod.sniffAddr, tui.Yellow(mod.sniffType))
	} else if dev.Type == network.HIDTypeUnknown {
		mod.Warning("device %s type has not been detected yet, falling back to '%s'", mod.sniffAddr, tui.Yellow(mod.sniffType))
	}

	var builder FrameBuilder
	if found && dev.Type != network.HIDTypeUnknown {
		// get the device specific protocol handler
		builder, found = FrameBuilders[dev.Type]
		if found == false {
			return errNotSupported(dev), nil, nil
		}
	} else {
		// get the device protocol handler from the hid.force.type parameter
		builder = builderFromName(mod.sniffType)
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
	if err := builder.BuildFrames(dev, cmds); err != nil {
		return err, nil, nil
	}

	return nil, dev, cmds
}

func (mod *HIDRecon) doInjection() {
	mod.writeLock.Lock()
	defer mod.writeLock.Unlock()

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

	devType := mod.sniffType
	if dev != nil {
		devType = dev.Type.String()
	}

	mod.Info("sending %d (%s) HID frames to %s (type:%s layout:%s) ...",
		numFrames,
		humanize.Bytes(uint64(szFrames)),
		tui.Bold(mod.sniffAddr),
		tui.Yellow(devType),
		tui.Yellow(mod.keyLayout))

	for i, cmd := range cmds {
		for j, frame := range cmd.Frames {
			for attempt := 0; attempt < 3; attempt++ {
				if err := mod.dongle.TransmitPayload(frame.Data, 500, 5); err != nil {
					if attempt < 2 {
						mod.Debug("error sending frame #%d of HID command #%d: %v, retrying ...", j, i, err)
					} else {
						mod.Error("error sending frame #%d of HID command #%d: %v", j, i, err)
					}
				} else {
					break
				}
			}

			if frame.Delay > 0 {
				mod.Debug("sleeping %dms after frame #%d of command #%d ...", frame.Delay, j, i)
				time.Sleep(frame.Delay)
			}
		}
		if cmd.Sleep > 0 {
			mod.Debug("sleeping %dms after command #%d ...", cmd.Sleep, i)
			time.Sleep(time.Duration(cmd.Sleep) * time.Millisecond)
		}
	}
}
