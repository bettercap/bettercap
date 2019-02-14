package gatt

// Supported statuses for GATT characteristic read/write operations.
// These correspond to att constants in the BLE spec
const (
	StatusSuccess         = 0
	StatusInvalidOffset   = 1
	StatusUnexpectedError = 2
)

// A Request is the context for a request from a connected central device.
// TODO: Replace this with more general context, such as:
// http://godoc.org/golang.org/x/net/context
type Request struct {
	Central Central
}

// A ReadRequest is a characteristic read request from a connected device.
type ReadRequest struct {
	Request
	Cap    int // maximum allowed reply length
	Offset int // request value offset
}

type Property int

// Characteristic property flags (spec 3.3.3.1)
const (
	CharBroadcast   Property = 0x01 // may be brocasted
	CharRead        Property = 0x02 // may be read
	CharWriteNR     Property = 0x04 // may be written to, with no reply
	CharWrite       Property = 0x08 // may be written to, with a reply
	CharNotify      Property = 0x10 // supports notifications
	CharIndicate    Property = 0x20 // supports Indications
	CharSignedWrite Property = 0x40 // supports signed write
	CharExtended    Property = 0x80 // supports extended properties
)

func (p Property) String() (result string) {
	if (p & CharBroadcast) != 0 {
		result += "broadcast "
	}
	if (p & CharRead) != 0 {
		result += "read "
	}
	if (p & CharWriteNR) != 0 {
		result += "writeWithoutResponse "
	}
	if (p & CharWrite) != 0 {
		result += "write "
	}
	if (p & CharNotify) != 0 {
		result += "notify "
	}
	if (p & CharIndicate) != 0 {
		result += "indicate "
	}
	if (p & CharSignedWrite) != 0 {
		result += "authenticateSignedWrites "
	}
	if (p & CharExtended) != 0 {
		result += "extendedProperties "
	}
	return
}

// A Service is a BLE service.
type Service struct {
	uuid  UUID
	chars []*Characteristic

	h    uint16
	endh uint16
}

// NewService creates and initialize a new Service using u as it's UUID.
func NewService(u UUID) *Service {
	return &Service{uuid: u}
}

// AddCharacteristic adds a characteristic to a service.
// AddCharacteristic panics if the service already contains another
// characteristic with the same UUID.
func (s *Service) AddCharacteristic(u UUID) *Characteristic {
	for _, c := range s.chars {
		if c.uuid.Equal(u) {
			panic("service already contains a characteristic with uuid " + u.String())
		}
	}
	c := &Characteristic{uuid: u, svc: s}
	s.chars = append(s.chars, c)
	return c
}

// UUID returns the UUID of the service.
func (s *Service) UUID() UUID { return s.uuid }

// Name returns the specificatin name of the service according to its UUID.
// If the UUID is not assigne, Name returns an empty string.
func (s *Service) Name() string {
	return knownServices[s.uuid.String()].Name
}

// Handle returns the Handle of the service.
func (s *Service) Handle() uint16 { return s.h }

// EndHandle returns the End Handle of the service.
func (s *Service) EndHandle() uint16 { return s.endh }

// SetHandle sets the Handle of the service.
func (s *Service) SetHandle(h uint16) { s.h = h }

// SetEndHandle sets the End Handle of the service.
func (s *Service) SetEndHandle(endh uint16) { s.endh = endh }

// SetCharacteristics sets the Characteristics of the service.
func (s *Service) SetCharacteristics(chars []*Characteristic) { s.chars = chars }

// Characteristic returns the contained characteristic of this service.
func (s *Service) Characteristics() []*Characteristic { return s.chars }

// A Characteristic is a BLE characteristic.
type Characteristic struct {
	uuid   UUID
	props  Property // enabled properties
	secure Property // security enabled properties
	svc    *Service
	cccd   *Descriptor
	descs  []*Descriptor

	value []byte

	// All the following fields are only used in peripheral/server implementation.
	rhandler ReadHandler
	whandler WriteHandler
	nhandler NotifyHandler

	h    uint16
	vh   uint16
	endh uint16
}

// NewCharacteristic creates and returns a Characteristic.
func NewCharacteristic(u UUID, s *Service, props Property, h uint16, vh uint16) *Characteristic {
	c := &Characteristic{
		uuid:  u,
		svc:   s,
		props: props,
		h:     h,
		vh:    vh,
	}

	return c
}

// Handle returns the Handle of the characteristic.
func (c *Characteristic) Handle() uint16 { return c.h }

// VHandle returns the Value Handle of the characteristic.
func (c *Characteristic) VHandle() uint16 { return c.vh }

// EndHandle returns the End Handle of the characteristic.
func (c *Characteristic) EndHandle() uint16 { return c.endh }

// Descriptor returns the Descriptor of the characteristic.
func (c *Characteristic) Descriptor() *Descriptor { return c.cccd }

// SetHandle sets the Handle of the characteristic.
func (c *Characteristic) SetHandle(h uint16) { c.h = h }

// SetVHandle sets the Value Handle of the characteristic.
func (c *Characteristic) SetVHandle(vh uint16) { c.vh = vh }

// SetEndHandle sets the End Handle of the characteristic.
func (c *Characteristic) SetEndHandle(endh uint16) { c.endh = endh }

// SetDescriptor sets the Descriptor of the characteristic.
func (c *Characteristic) SetDescriptor(cccd *Descriptor) { c.cccd = cccd }

// SetDescriptors sets the list of Descriptor of the characteristic.
func (c *Characteristic) SetDescriptors(descs []*Descriptor) { c.descs = descs }

// UUID returns the UUID of the characteristic.
func (c *Characteristic) UUID() UUID {
	return c.uuid
}

// Name returns the specificatin name of the characteristic.
// If the UUID is not assigned, Name returns empty string.
func (c *Characteristic) Name() string {
	return knownCharacteristics[c.uuid.String()].Name
}

// Service returns the containing service of this characteristic.
func (c *Characteristic) Service() *Service {
	return c.svc
}

// Properties returns the properties of this characteristic.
func (c *Characteristic) Properties() Property {
	return c.props
}

// Descriptors returns the contained descriptors of this characteristic.
func (c *Characteristic) Descriptors() []*Descriptor {
	return c.descs
}

// AddDescriptor adds a descriptor to a characteristic.
// AddDescriptor panics if the characteristic already contains another
// descriptor with the same UUID.
func (c *Characteristic) AddDescriptor(u UUID) *Descriptor {
	for _, d := range c.descs {
		if d.uuid.Equal(u) {
			panic("service already contains a characteristic with uuid " + u.String())
		}
	}
	d := &Descriptor{uuid: u, char: c}
	c.descs = append(c.descs, d)
	return d
}

// SetValue makes the characteristic support read requests, and returns a
// static value. SetValue must be called before the containing service is
// added to a server.
// SetValue panics if the characteristic has been configured with a ReadHandler.
func (c *Characteristic) SetValue(b []byte) {
	if c.rhandler != nil {
		panic("charactristic has been configured with a read handler")
	}
	c.props |= CharRead
	// c.secure |= CharRead
	c.value = make([]byte, len(b))
	copy(c.value, b)
}

// HandleRead makes the characteristic support read requests, and routes read
// requests to h. HandleRead must be called before the containing service is
// added to a server.
// HandleRead panics if the characteristic has been configured with a static value.
func (c *Characteristic) HandleRead(h ReadHandler) {
	if c.value != nil {
		panic("charactristic has been configured with a static value")
	}
	c.props |= CharRead
	// c.secure |= CharRead
	c.rhandler = h
}

// HandleReadFunc calls HandleRead(ReadHandlerFunc(f)).
func (c *Characteristic) HandleReadFunc(f func(rsp ResponseWriter, req *ReadRequest)) {
	c.HandleRead(ReadHandlerFunc(f))
}

func (c *Characteristic) GetReadHandler() ReadHandler {
	return c.rhandler
}

// HandleWrite makes the characteristic support write and write-no-response
// requests, and routes write requests to h.
// The WriteHandler does not differentiate between write and write-no-response
// requests; it is handled automatically.
// HandleWrite must be called before the containing service is added to a server.
func (c *Characteristic) HandleWrite(h WriteHandler) {
	c.props |= CharWrite | CharWriteNR
	// c.secure |= CharWrite | CharWriteNR
	c.whandler = h
}

// HandleWriteFunc calls HandleWrite(WriteHandlerFunc(f)).
func (c *Characteristic) HandleWriteFunc(f func(r Request, data []byte) (status byte)) {
	c.HandleWrite(WriteHandlerFunc(f))
}

func (c *Characteristic) GetWriteHandler() WriteHandler {
	return c.whandler
}

// HandleNotify makes the characteristic support notify requests, and routes
// notification requests to h. HandleNotify must be called before the
// containing service is added to a server.
func (c *Characteristic) HandleNotify(h NotifyHandler) {
	if c.cccd != nil {
		return
	}
	p := CharNotify | CharIndicate
	c.props |= p
	c.nhandler = h

	// add ccc (client characteristic configuration) descriptor
	secure := Property(0)
	// If the characteristic requested secure notifications,
	// then set ccc security to r/w.
	if c.secure&p != 0 {
		secure = CharRead | CharWrite
	}
	cd := &Descriptor{
		uuid:   attrClientCharacteristicConfigUUID,
		props:  CharRead | CharWrite | CharWriteNR,
		secure: secure,
		// FIXME: currently, we always return 0, which is inaccurate.
		// Each connection should have it's own copy of this value.
		value: []byte{0x00, 0x00},
		char:  c,
	}
	c.cccd = cd
	c.descs = append(c.descs, cd)
}

// HandleNotifyFunc calls HandleNotify(NotifyHandlerFunc(f)).
func (c *Characteristic) HandleNotifyFunc(f func(r Request, n Notifier)) {
	c.HandleNotify(NotifyHandlerFunc(f))
}

func (c *Characteristic) GetNotifyHandler() NotifyHandler {
	return c.nhandler
}

// TODO
// func (c *Characteristic) SubscribedCentrals() []Central{
// }

// Descriptor is a BLE descriptor
type Descriptor struct {
	uuid   UUID
	char   *Characteristic
	props  Property // enabled properties
	secure Property // security enabled properties

	h        uint16
	value    []byte
	valuestr string

	rhandler ReadHandler
	whandler WriteHandler
}

// Handle returns the Handle of the descriptor.
func (d *Descriptor) Handle() uint16 { return d.h }

// SetHandle sets the Handle of the descriptor.
func (d *Descriptor) SetHandle(h uint16) { d.h = h }

// NewDescriptor creates and returns a Descriptor.
func NewDescriptor(u UUID, h uint16, char *Characteristic) *Descriptor {
	cd := &Descriptor{
		uuid: u,
		h:    h,
		char: char,
	}
	return cd
}

// UUID returns the UUID of the descriptor.
func (d *Descriptor) UUID() UUID {
	return d.uuid
}

// Name returns the specificatin name of the descriptor.
// If the UUID is not assigned, returns an empty string.
func (d *Descriptor) Name() string {
	return knownDescriptors[d.uuid.String()].Name
}

// Characteristic returns the containing characteristic of the descriptor.
func (d *Descriptor) Characteristic() *Characteristic {
	return d.char
}

// SetValue makes the descriptor support read requests, and returns a static value.
// SetValue must be called before the containing service is added to a server.
// SetValue panics if the descriptor has already configured with a ReadHandler.
func (d *Descriptor) SetValue(b []byte) {
	if d.rhandler != nil {
		panic("descriptor has been configured with a read handler")
	}
	d.props |= CharRead
	// d.secure |= CharRead
	d.value = make([]byte, len(b))
	copy(d.value, b)
}

// SetStringValue makes the descriptor support read requests, and returns a static value.
// SetStringValue must be called before the containing service is added to a server.
// SetStringValue panics if the descriptor has already configured with a ReadHandler.
func (d *Descriptor) SetStringValue(s string) {
	if d.rhandler != nil {
		panic("descriptor has been configured with a read handler")
	}
	d.props |= CharRead
	// d.secure |= CharRead
	d.valuestr = s
}

// HandleRead makes the descriptor support read requests, and routes read requests to h.
// HandleRead must be called before the containing service is added to a server.
// HandleRead panics if the descriptor has been configured with a static value.
func (d *Descriptor) HandleRead(h ReadHandler) {
	if d.value != nil {
		panic("descriptor has been configured with a static value")
	}
	d.props |= CharRead
	// d.secure |= CharRead
	d.rhandler = h
}

// HandleReadFunc calls HandleRead(ReadHandlerFunc(f)).
func (d *Descriptor) HandleReadFunc(f func(rsp ResponseWriter, req *ReadRequest)) {
	d.HandleRead(ReadHandlerFunc(f))
}

// HandleWrite makes the descriptor support write and write-no-response requests, and routes write requests to h.
// The WriteHandler does not differentiate between write and write-no-response requests; it is handled automatically.
// HandleWrite must be called before the containing service is added to a server.
func (d *Descriptor) HandleWrite(h WriteHandler) {
	d.props |= CharWrite | CharWriteNR
	// d.secure |= CharWrite | CharWriteNR
	d.whandler = h
}

// HandleWriteFunc calls HandleWrite(WriteHandlerFunc(f)).
func (d *Descriptor) HandleWriteFunc(f func(r Request, data []byte) (status byte)) {
	d.HandleWrite(WriteHandlerFunc(f))
}
