package gatt

import "errors"

var notImplemented = errors.New("not implemented")

type State int

const (
	StateUnknown      State = 0
	StateResetting    State = 1
	StateUnsupported  State = 2
	StateUnauthorized State = 3
	StatePoweredOff   State = 4
	StatePoweredOn    State = 5
)

func (s State) String() string {
	str := []string{
		"Unknown",
		"Resetting",
		"Unsupported",
		"Unauthorized",
		"PoweredOff",
		"PoweredOn",
	}
	return str[int(s)]
}

// Device defines the interface for a BLE device.
// Since an interface can't define fields(properties). To implement the
// callback support for cerntain events, deviceHandler is defined and
// implementation of Device on different platforms should embed it in
// order to keep have keep compatible in API level.
// Package users can use the Handler to set these handlers.
type Device interface {
	Init(stateChanged func(Device, State)) error

	// Advertise advertise AdvPacket
	Advertise(a *AdvPacket) error

	// AdvertiseNameAndServices advertises device name, and specified service UUIDs.
	// It tres to fit the UUIDs in the advertising packet as much as possible.
	// If name doesn't fit in the advertising packet, it will be put in scan response.
	AdvertiseNameAndServices(name string, ss []UUID) error

	// AdvertiseIBeaconData advertise iBeacon with given manufacturer data.
	AdvertiseIBeaconData(b []byte) error

	// AdvertisingIbeacon advertises iBeacon with specified parameters.
	AdvertiseIBeacon(u UUID, major, minor uint16, pwr int8) error

	// StopAdvertising stops advertising.
	StopAdvertising() error

	// RemoveAllServices removes all services that are currently in the database.
	RemoveAllServices() error

	// Add Service add a service to database.
	AddService(s *Service) error

	// SetServices set the specified service to the database.
	// It removes all currently added services, if any.
	SetServices(ss []*Service) error

	// Scan discovers surounding remote peripherals that have the Service UUID specified in ss.
	// If ss is set to nil, all devices scanned are reported.
	// dup specifies weather duplicated advertisement should be reported or not.
	// When a remote peripheral is discovered, the PeripheralDiscovered Handler is called.
	Scan(ss []UUID, dup bool)

	// StopScanning stops scanning.
	StopScanning()

	// Stop calls OS specific close calls
	Stop() error

	// Connect connects to a remote peripheral.
	Connect(p Peripheral)

	// CancelConnection disconnects a remote peripheral.
	CancelConnection(p Peripheral)

	// Handle registers the specified handlers.
	Handle(h ...Handler)

	// Option sets the options specified.
	Option(o ...Option) error
}

// deviceHandler is the handlers(callbacks) of the Device.
type deviceHandler struct {
	// stateChanged is called when the device states changes.
	stateChanged func(d Device, s State)

	// connect is called when a remote central device connects to the device.
	centralConnected func(c Central)

	// disconnect is called when a remote central device disconnects to the device.
	centralDisconnected func(c Central)

	// peripheralDiscovered is called when a remote peripheral device is found during scan procedure.
	peripheralDiscovered func(p Peripheral, a *Advertisement, rssi int)

	// peripheralConnected is called when a remote peripheral is conneted.
	peripheralConnected func(p Peripheral, err error)

	// peripheralConnected is called when a remote peripheral is disconneted.
	peripheralDisconnected func(p Peripheral, err error)
}

func getDeviceHandler(d Device) *deviceHandler {
	switch t := d.(type) {
	case *device:
		return &t.deviceHandler
	case *simDevice:
		return &t.deviceHandler
	default:
		return nil
	}
}

// A Handler is a self-referential function, which registers the options specified.
// See http://commandcenter.blogspot.com.au/2014/01/self-referential-functions-and-design.html for more discussion.
type Handler func(Device)

// Handle registers the specified handlers.
func (d *device) Handle(hh ...Handler) {
	for _, h := range hh {
		h(d)
	}
}

// CentralConnected returns a Handler, which sets the specified function to be called when a device connects to the server.
func CentralConnected(f func(Central)) Handler {
	return func(d Device) { getDeviceHandler(d).centralConnected = f }
}

// CentralDisconnected returns a Handler, which sets the specified function to be called when a device disconnects from the server.
func CentralDisconnected(f func(Central)) Handler {
	return func(d Device) { getDeviceHandler(d).centralDisconnected = f }
}

// PeripheralDiscovered returns a Handler, which sets the specified function to be called when a remote peripheral device is found during scan procedure.
func PeripheralDiscovered(f func(Peripheral, *Advertisement, int)) Handler {
	return func(d Device) { getDeviceHandler(d).peripheralDiscovered = f }
}

// PeripheralConnected returns a Handler, which sets the specified function to be called when a remote peripheral device connects.
func PeripheralConnected(f func(Peripheral, error)) Handler {
	return func(d Device) { getDeviceHandler(d).peripheralConnected = f }
}

// PeripheralDisconnected returns a Handler, which sets the specified function to be called when a remote peripheral device disconnects.
func PeripheralDisconnected(f func(Peripheral, error)) Handler {
	return func(d Device) { getDeviceHandler(d).peripheralDisconnected = f }
}

// An Option is a self-referential function, which sets the option specified.
// Most Options are platform-specific, which gives more fine-grained control over the device at a cost of losing portibility.
// See http://commandcenter.blogspot.com.au/2014/01/self-referential-functions-and-design.html for more discussion.
type Option func(Device) error

// Option sets the options specified.
// Some options can only be set before the device is initialized; they are best used with NewDevice instead of Option.
func (d *device) Option(opts ...Option) error {
	var err error
	for _, opt := range opts {
		err = opt(d)
	}
	return err
}
