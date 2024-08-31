package can

import (
	"go.einride.tech/can"
)

func (msg *OBD2Message) ParseResponse(frame can.Frame) bool {
	msgSize := frame.Data[0]
	// validate data size
	if msgSize > 7 {
		// fmt.Printf("invalid response size %d\n", msgSize)
		return false
	}

	svcID := frame.Data[1] - 0x40

	msg.Type = OBD2MessageTypeResponse
	msg.ECU = uint8(uint16(frame.ID) - uint16(OBD2ECUResponseMinID))
	msg.Size = msgSize - 3
	msg.Service = OBD2Service(svcID)
	msg.PID = lookupPID(svcID, []uint8{frame.Data[2]})
	msg.Data = frame.Data[3 : 3+msg.Size]

	return true
}
