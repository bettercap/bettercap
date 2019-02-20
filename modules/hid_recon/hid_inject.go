package hid_recon

import (
	"fmt"
	"time"

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

func (mod *HIDRecon) doInjection() {
	dev, found := mod.Session.HID.Get(mod.sniffAddr)
	if found == false {
		mod.Warning("could not find HID device %s", mod.sniffAddr)
		return
	}

	builder, found := FrameBuilders[dev.Type]
	if found == false {
		mod.Warning("HID frame injection is not supported for device type %s", dev.Type.String())
		return
	}

	keyLayout := KeyMapFor(mod.keyLayout)
	if keyLayout == nil {
		mod.Warning("could not find keymap for '%s' layout", mod.keyLayout)
		return
	}

	str := "hello world from bettercap ^_^"
	cmds := make([]*Command, 0)
	for _, c := range str {
		ch := fmt.Sprintf("%c", c)
		if m, found := keyLayout[ch]; found {
			cmds = append(cmds, &Command{
				Char: ch,
				HID:  m.HID,
				Mode: m.Mode,
			})
		} else {
			mod.Warning("could not find HID command for '%c'", ch)
			return
		}
	}

	builder.BuildFrames(cmds)
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
