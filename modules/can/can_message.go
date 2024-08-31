package can

import (
	"github.com/bettercap/bettercap/v2/network"
	"go.einride.tech/can"
)

type Message struct {
	// the raw frame
	Frame can.Frame
	// parsed as OBD2
	OBD2 *OBD2Message
	// parsed from DBC
	Name    string
	Source  *network.CANDevice
	Signals map[string]string
}

func NewCanMessage(frame can.Frame) Message {
	return Message{
		Frame:   frame,
		Signals: make(map[string]string),
	}
}
