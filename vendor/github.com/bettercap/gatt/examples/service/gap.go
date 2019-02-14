package service

import "github.com/bettercap/gatt"

var (
	attrGAPUUID = gatt.UUID16(0x1800)

	attrDeviceNameUUID        = gatt.UUID16(0x2A00)
	attrAppearanceUUID        = gatt.UUID16(0x2A01)
	attrPeripheralPrivacyUUID = gatt.UUID16(0x2A02)
	attrReconnectionAddrUUID  = gatt.UUID16(0x2A03)
	attrPeferredParamsUUID    = gatt.UUID16(0x2A04)
)

// https://developer.bluetooth.org/gatt/characteristics/Pages/CharacteristicViewer.aspx?u=org.bluetooth.characteristic.gap.appearance.xml
var gapCharAppearanceGenericComputer = []byte{0x00, 0x80}

// NOTE: OS X provides GAP and GATT services, and they can't be customized.
// For Linux/Embedded, however, this is something we want to fully control.
func NewGapService(name string) *gatt.Service {
	s := gatt.NewService(attrGAPUUID)
	s.AddCharacteristic(attrDeviceNameUUID).SetValue([]byte(name))
	s.AddCharacteristic(attrAppearanceUUID).SetValue(gapCharAppearanceGenericComputer)
	s.AddCharacteristic(attrPeripheralPrivacyUUID).SetValue([]byte{0x00})
	s.AddCharacteristic(attrReconnectionAddrUUID).SetValue([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	s.AddCharacteristic(attrPeferredParamsUUID).SetValue([]byte{0x06, 0x00, 0x06, 0x00, 0x00, 0x00, 0xd0, 0x07})
	return s
}
