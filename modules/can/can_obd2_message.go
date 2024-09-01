package can

import (
	"fmt"
)

// https://en.wikipedia.org/wiki/OBD-II_PIDs

// https://www.csselectronics.com/pages/obd2-explained-simple-intro
// https://www.csselectronics.com/pages/obd2-pid-table-on-board-diagnostics-j1979

// https://stackoverflow.com/questions/40826932/how-can-i-get-mode-pids-from-raw-obd2-identifier-11-or-29-bit

// https://github.com/ejvaughan/obdii/blob/master/src/OBDII.c

const OBD2BroadcastRequestID = 0x7DF
const OBD2ECUResponseMinID = 0x7E0
const OBD2ECUResponseMaxID = 0x7EF
const OBD2BroadcastRequestID29bit = 0x18DB33F1
const OBD2ECUResponseMinID29bit = 0x18DAF100
const OBD2ECUResponseMaxID29bit = 0x18DAF1FF

type OBD2Service uint8

func (s OBD2Service) String() string {
	switch s {
	case 0x01:
		return "Show current data"
	case 0x02:
		return "Show freeze frame data"
	case 0x03:
		return "Show stored Diagnostic Trouble Codes"
	case 0x04:
		return "Clear Diagnostic Trouble Codes and stored values"
	case 0x05:
		return "Test results, oxygen sensor monitoring (non CAN only)"
	case 0x06:
		return "Test results, other component/system monitoring (Test results, oxygen sensor monitoring for CAN only)"
	case 0x07:
		return "Show pending Diagnostic Trouble Codes (detected during current or last driving cycle)"
	case 0x08:
		return "Control operation of on-board component/system"
	case 0x09:
		return "Request vehicle information"
	case 0x0A:
		return "Permanent Diagnostic Trouble Codes (DTCs) (Cleared DTCs)"
	}

	return fmt.Sprintf("service 0x%x", uint8(s))
}

type OBD2MessageType uint8

const (
	OBD2MessageTypeRequest OBD2MessageType = iota
	OBD2MessageTypeResponse
)

func (t OBD2MessageType) String() string {
	if t == OBD2MessageTypeRequest {
		return "request"
	} else {
		return "response"
	}
}

type OBD2Message struct {
	Type    OBD2MessageType
	ECU     uint8
	Service OBD2Service
	PID     OBD2PID
	Size    uint8
	Data    []uint8
}
