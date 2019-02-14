package gatt

import (
	"errors"
	"sync"
)

// Peripheral is the interface that represent a remote peripheral device.
type Peripheral interface {
	// Device returns the underlying device.
	Device() Device

	// ID is the platform specific unique ID of the remote peripheral, e.g. MAC for Linux, Peripheral UUID for MacOS.
	ID() string

	// Name returns the name of the remote peripheral.
	// This can be the advertised name, if exists, or the GAP device name, which takes priority
	Name() string

	// Services returnns the services of the remote peripheral which has been discovered.
	Services() []*Service

	// DiscoverServices discover the specified services of the remote peripheral.
	// If the specified services is set to nil, all the available services of the remote peripheral are returned.
	DiscoverServices(s []UUID) ([]*Service, error)

	// DiscoverIncludedServices discovers the specified included services of a service.
	// If the specified services is set to nil, all the included services of the service are returned.
	DiscoverIncludedServices(ss []UUID, s *Service) ([]*Service, error)

	// DiscoverCharacteristics discovers the specified characteristics of a service.
	// If the specified characterstics is set to nil, all the characteristic of the service are returned.
	DiscoverCharacteristics(c []UUID, s *Service) ([]*Characteristic, error)

	// DiscoverDescriptors discovers the descriptors of a characteristic.
	// If the specified descriptors is set to nil, all the descriptors of the characteristic are returned.
	DiscoverDescriptors(d []UUID, c *Characteristic) ([]*Descriptor, error)

	// ReadCharacteristic retrieves the value of a specified characteristic.
	ReadCharacteristic(c *Characteristic) ([]byte, error)

	// ReadLongCharacteristic retrieves the value of a specified characteristic that is longer than the
	// MTU.
	ReadLongCharacteristic(c *Characteristic) ([]byte, error)

	// ReadDescriptor retrieves the value of a specified characteristic descriptor.
	ReadDescriptor(d *Descriptor) ([]byte, error)

	// WriteCharacteristic writes the value of a characteristic.
	WriteCharacteristic(c *Characteristic, b []byte, noRsp bool) error

	// WriteDescriptor writes the value of a characteristic descriptor.
	WriteDescriptor(d *Descriptor, b []byte) error

	// SetNotifyValue sets notifications for the value of a specified characteristic.
	SetNotifyValue(c *Characteristic, f func(*Characteristic, []byte, error)) error

	// SetIndicateValue sets indications for the value of a specified characteristic.
	SetIndicateValue(c *Characteristic, f func(*Characteristic, []byte, error)) error

	// ReadRSSI retrieves the current RSSI value for the remote peripheral.
	ReadRSSI() int

	// SetMTU sets the mtu for the remote peripheral.
	SetMTU(mtu uint16) error
}

type subscriber struct {
	sub map[uint16]subscribefn
	mu  *sync.Mutex
}

type subscribefn func([]byte, error)

func newSubscriber() *subscriber {
	return &subscriber{
		sub: make(map[uint16]subscribefn),
		mu:  &sync.Mutex{},
	}
}

func (s *subscriber) subscribe(h uint16, f subscribefn) {
	s.mu.Lock()
	s.sub[h] = f
	s.mu.Unlock()
}

func (s *subscriber) unsubscribe(h uint16) {
	s.mu.Lock()
	delete(s.sub, h)
	s.mu.Unlock()
}

func (s *subscriber) fn(h uint16) subscribefn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sub[h]
}

var (
	ErrInvalidLength = errors.New("invalid length")
)
