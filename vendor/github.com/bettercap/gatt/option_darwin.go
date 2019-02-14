package gatt

const (
	CentralManager    = 0 // Client functions (default)
	PeripheralManager = 1 // Server functions
)

// MacDeviceRole specify the XPC connection type to connect blued.
// THis option can only be used with NewDevice on OS X implementation.
func MacDeviceRole(r int) Option {
	return func(d Device) error {
		d.(*device).role = r
		return nil
	}
}
