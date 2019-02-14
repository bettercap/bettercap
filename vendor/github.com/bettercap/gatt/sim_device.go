package gatt

import (
	"errors"
)

const (
	DefaultMTU = 1024
)

type simDevice struct {
	deviceHandler

	s              *Service
	advertisedName string
}

func NewSimDeviceClient(service *Service, advertisedName string) *simDevice {
	return &simDevice{
		s:              service,
		advertisedName: advertisedName,
	}
}

func (d *simDevice) Init(stateChanged func(Device, State)) error {
	d.stateChanged = stateChanged
	go stateChanged(d, StatePoweredOn)
	return nil
}

func (d *simDevice) Advertise(a *AdvPacket) error {
	return errors.New("Method not supported")
}

func (d *simDevice) AdvertiseNameAndServices(name string, ss []UUID) error {
	return errors.New("Method not supported")
}

func (d *simDevice) AdvertiseIBeaconData(b []byte) error {
	return errors.New("Method not supported")
}

func (d *simDevice) AdvertiseIBeacon(u UUID, major, minor uint16, pwr int8) error {
	return errors.New("Method not supported")
}

func (d *simDevice) StopAdvertising() error {
	return errors.New("Method not supported")
}

func (d *simDevice) RemoveAllServices() error {
	return errors.New("Method not supported")
}

func (d *simDevice) AddService(s *Service) error {
	return errors.New("Method not supported")
}

func (d *simDevice) SetServices(ss []*Service) error {
	return errors.New("Method not supported")
}

func (d *simDevice) Scan(ss []UUID, dup bool) {
	for _, s := range ss {
		if s.Equal(d.s.UUID()) {
			go d.peripheralDiscovered(
				&simPeripheral{d},
				&Advertisement{LocalName: d.advertisedName},
				0,
			)
		}
	}
}

func (d *simDevice) StopScanning() {
}

func (d *simDevice) Stop() error {
	go d.stateChanged(d, StatePoweredOff)
	return nil
}

func (d *simDevice) Connect(p Peripheral) {
	go d.peripheralConnected(p, nil)
}

func (d *simDevice) CancelConnection(p Peripheral) {
	go d.peripheralDisconnected(p, nil)
}

func (d *simDevice) Handle(hh ...Handler) {
	for _, h := range hh {
		h(d)
	}
}

func (d *simDevice) Option(o ...Option) error {
	return errors.New("Method not supported")
}

type simPeripheral struct {
	d *simDevice
}

func (p *simPeripheral) Device() Device {
	return p.d
}

func (p *simPeripheral) ID() string {
	return "Sim ID"
}

func (p *simPeripheral) Name() string {
	return "Sim"
}

func (p *simPeripheral) Services() []*Service {
	return []*Service{p.d.s}
}

func (p *simPeripheral) DiscoverServices(ss []UUID) ([]*Service, error) {
	for _, s := range ss {
		if s.Equal(p.d.s.UUID()) {
			return []*Service{p.d.s}, nil
		}
	}
	return []*Service{}, nil
}

func (p *simPeripheral) DiscoverIncludedServices(ss []UUID, s *Service) ([]*Service, error) {
	return nil, errors.New("Method not supported")
}

func (p *simPeripheral) DiscoverCharacteristics(cc []UUID, s *Service) ([]*Characteristic, error) {
	requestedUUIDs := make(map[string]bool)
	for _, c := range cc {
		requestedUUIDs[c.String()] = true
	}
	foundChars := make([]*Characteristic, 0)
	for _, c := range p.d.s.Characteristics() {
		if _, present := requestedUUIDs[c.UUID().String()]; present {
			foundChars = append(foundChars, c)
		}
	}
	return foundChars, nil
}

func (p *simPeripheral) DiscoverDescriptors(d []UUID, c *Characteristic) ([]*Descriptor, error) {
	return nil, errors.New("Method not supported")
}

func (p *simPeripheral) ReadCharacteristic(c *Characteristic) ([]byte, error) {
	rhandler := c.GetReadHandler()
	if rhandler != nil {
		rsp := newResponseWriter(DefaultMTU)
		req := &ReadRequest{}
		rhandler.ServeRead(rsp, req)
		return rsp.buf.Bytes(), nil
	} else {
		return nil, AttEcodeReadNotPerm
	}
}

func (p *simPeripheral) ReadLongCharacteristic(c *Characteristic) ([]byte, error) {
	return p.ReadCharacteristic(c)
}

func (p *simPeripheral) ReadDescriptor(d *Descriptor) ([]byte, error) {
	return nil, errors.New("Method not supported")
}

func (p *simPeripheral) WriteCharacteristic(c *Characteristic, b []byte, noRsp bool) error {
	whandler := c.GetWriteHandler()
	if whandler != nil {
		r := Request{}
		if res := whandler.ServeWrite(r, b); res != 0 {
			return AttEcode(res)
		} else {
			return nil
		}
	} else {
		return AttEcodeWriteNotPerm
	}
}

func (p *simPeripheral) WriteDescriptor(d *Descriptor, b []byte) error {
	return errors.New("Method not supported")
}

func (p *simPeripheral) SetNotifyValue(c *Characteristic, f func(*Characteristic, []byte, error)) error {
	return errors.New("Method not supported")
}

func (p *simPeripheral) SetIndicateValue(c *Characteristic, f func(*Characteristic, []byte, error)) error {
	return errors.New("Method not supported")
}

func (p *simPeripheral) ReadRSSI() int {
	return 0
}

func (p *simPeripheral) SetMTU(mtu uint16) error {
	return errors.New("Method not supported")
}
