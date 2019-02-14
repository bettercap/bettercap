package service

import (
	"log"

	"github.com/bettercap/gatt"
)

var (
	attrGATTUUID           = gatt.UUID16(0x1801)
	attrServiceChangedUUID = gatt.UUID16(0x2A05)
)

// NOTE: OS X provides GAP and GATT services, and they can't be customized.
// For Linux/Embedded, however, this is something we want to fully control.
func NewGattService() *gatt.Service {
	s := gatt.NewService(attrGATTUUID)
	s.AddCharacteristic(attrServiceChangedUUID).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			go func() {
				log.Printf("TODO: indicate client when the services are changed")
			}()
		})
	return s
}
