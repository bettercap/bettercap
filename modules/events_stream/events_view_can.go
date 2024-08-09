package events_stream

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/bettercap/bettercap/v2/modules/can"
	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewCANEvent(output io.Writer, e session.Event) {
	if e.Tag == "can.device.new" {
		dev := e.Data.(*network.CANDevice)
		fmt.Fprintf(output, "[%s] [%s] new CAN device %s (%s) detected.\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Bold(dev.Name),
			tui.Dim(dev.Description))
	} else if e.Tag == "can.message" {
		msg := e.Data.(can.Message)

		// unparsed
		if msg.Name == "" {
			fmt.Fprintf(output, "[%s] [%s] <id %d> (%s): %s\n",
				e.Time.Format(mod.timeFormat),
				tui.Green(e.Tag),
				msg.Frame.ID,
				tui.Dim(humanize.Bytes(uint64(msg.Frame.Length))),
				hex.EncodeToString(msg.Frame.Data[:msg.Frame.Length]))
		} else {
			fmt.Fprintf(output, "[%s] [%s] <id %d> %s (%s) from %s:\n",
				e.Time.Format(mod.timeFormat),
				tui.Green(e.Tag),
				msg.Frame.ID,
				msg.Name,
				tui.Dim(humanize.Bytes(uint64(msg.Frame.Length))),
				tui.Bold(msg.Source.Name))

			for name, value := range msg.Signals {
				fmt.Fprintf(output, "  %s : %s\n", name, value)
			}
		}
	} else {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
