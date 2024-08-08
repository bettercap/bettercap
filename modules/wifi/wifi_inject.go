//go:build !darwin

package wifi

import (
	"time"
)

func (mod *WiFiModule) injectPacket(data []byte) {
	if err := mod.handle.WritePacketData(data); err != nil {
		mod.Error("could not inject WiFi packet: %s", err)
		mod.Session.Queue.TrackError()
	} else {
		mod.Session.Queue.TrackSent(uint64(len(data)))
	}
	// let the network card breath a little
	time.Sleep(10 * time.Millisecond)
}
