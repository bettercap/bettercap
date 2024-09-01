package can

import (
	"fmt"
	"sync"
)

type OBD2 struct {
	sync.RWMutex

	enabled bool
}

func (obd *OBD2) Enabled() bool {
	obd.RLock()
	defer obd.RUnlock()
	return obd.enabled
}

func (obd *OBD2) Enable(enable bool) {
	obd.RLock()
	defer obd.RUnlock()
	obd.enabled = enable
}

func (obd *OBD2) Parse(mod *CANModule, msg *Message) bool {
	obd.RLock()
	defer obd.RUnlock()

	// did we load any DBC database?
	if !obd.enabled {
		return false
	}

	odbMessage := &OBD2Message{}

	if msg.Frame.ID == OBD2BroadcastRequestID || msg.Frame.ID == OBD2BroadcastRequestID29bit {
		// parse as request
		if odbMessage.ParseRequest(msg.Frame) {
			msg.OBD2 = odbMessage
			return true
		}
	} else if (msg.Frame.ID >= OBD2ECUResponseMinID && msg.Frame.ID <= OBD2ECUResponseMaxID) ||
		(msg.Frame.ID >= OBD2ECUResponseMinID29bit && msg.Frame.ID <= OBD2ECUResponseMaxID29bit) {
		// parse as response
		if odbMessage.ParseResponse(msg.Frame) {
			msg.OBD2 = odbMessage
			// add CAN source if new
			_, msg.Source = mod.Session.CAN.AddIfNew(fmt.Sprintf("ECU_%d", odbMessage.ECU), "", msg.Frame.Data[:])
			return true
		}
	}

	return false
}
