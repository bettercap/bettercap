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

func (mod *EventsStream) viewCANDeviceNew(output io.Writer, e session.Event) {
	dev := e.Data.(*network.CANDevice)
	fmt.Fprintf(output, "[%s] [%s] new CAN device %s (%s) detected.\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		tui.Bold(dev.Name),
		tui.Dim(dev.Description))
}

func (mod *EventsStream) viewCANRawMessage(output io.Writer, e session.Event) {
	msg := e.Data.(can.Message)

	fmt.Fprintf(output, "[%s] [%s] %s <0x%x> (%s): %s\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		tui.Dim("raw"),
		msg.Frame.ID,
		tui.Dim(humanize.Bytes(uint64(msg.Frame.Length))),
		hex.EncodeToString(msg.Frame.Data[:msg.Frame.Length]))
}

func (mod *EventsStream) viewCANDBCMessage(output io.Writer, e session.Event) {
	msg := e.Data.(can.Message)
	src := ""
	if msg.Source != nil && msg.Source.Name != "" {
		src = fmt.Sprintf(" from %s", msg.Source.Name)
	}

	fmt.Fprintf(output, "[%s] [%s] (dbc) <0x%x> %s (%s)%s:\n",
		e.Time.Format(mod.timeFormat),
		tui.Green(e.Tag),
		msg.Frame.ID,
		msg.Name,
		tui.Dim(humanize.Bytes(uint64(msg.Frame.Length))),
		tui.Bold(src))

	for name, value := range msg.Signals {
		fmt.Fprintf(output, "  %s : %s\n", name, value)
	}
}

func (mod *EventsStream) viewCANOBDMessage(output io.Writer, e session.Event) {
	msg := e.Data.(can.Message)
	obd2 := msg.OBD2

	if obd2.Type == can.OBD2MessageTypeRequest {
		fmt.Fprintf(output, "[%s] [%s] %s : %s > %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Yellow("obd2.request"),
			obd2.Service, obd2.PID)
	} else {

		fmt.Fprintf(output, "[%s] [%s] %s : %s > %s > %s : 0x%x\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Yellow("obd2.response"),
			tui.Bold(msg.Source.Name),
			obd2.Service, obd2.PID,
			obd2.Data)
	}

}

func (mod *EventsStream) viewCANEvent(output io.Writer, e session.Event) {
	if e.Tag == "can.device.new" {
		mod.viewCANDeviceNew(output, e)
	} else if e.Tag == "can.message" {
		msg := e.Data.(can.Message)
		if msg.OBD2 != nil {
			// OBD-2 PID
			mod.viewCANOBDMessage(output, e)
		} else if msg.Name != "" {
			// parsed from DBC
			mod.viewCANDBCMessage(output, e)
		} else {
			// raw unparsed frame
			mod.viewCANRawMessage(output, e)
		}
	} else {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
