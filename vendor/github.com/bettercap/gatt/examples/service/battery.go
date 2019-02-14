package service

import "github.com/bettercap/gatt"

func NewBatteryService() *gatt.Service {
	lv := byte(100)
	s := gatt.NewService(gatt.UUID16(0x180F))
	c := s.AddCharacteristic(gatt.UUID16(0x2A19))
	c.HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			rsp.Write([]byte{lv})
			lv--
		})

	// Characteristic User Description
	c.AddDescriptor(gatt.UUID16(0x2901)).SetStringValue("Battery level between 0 and 100 percent")

	// Characteristic Presentation Format
	c.AddDescriptor(gatt.UUID16(0x2904)).SetValue([]byte{4, 1, 39, 173, 1, 0, 0})

	return s
}
