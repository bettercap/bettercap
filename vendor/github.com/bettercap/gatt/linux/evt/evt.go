package evt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/bettercap/gatt/linux/util"
)

type EventHandler interface {
	HandleEvent([]byte) error
}

type HandlerFunc func(b []byte) error

func (f HandlerFunc) HandleEvent(b []byte) error {
	return f(b)
}

type Evt struct {
	evtHandlers map[int]EventHandler
}

func NewEvt() *Evt {
	return &Evt{
		evtHandlers: map[int]EventHandler{},
	}
}

func (e *Evt) HandleEvent(c int, h EventHandler) {
	e.evtHandlers[c] = h
}

func (e *Evt) Dispatch(b []byte) error {
	h := &EventHeader{}
	if err := h.Unmarshal(b); err != nil {
		return err
	}
	b = b[2:] // Skip Event Header (uint8 + uint8)
	if f, found := e.evtHandlers[h.code]; found {
		e.trace("> HCI Event: %s (0x%02X) plen %d: [ % X ])\n", h.code, uint8(h.code), h.plen, b)
		return f.HandleEvent(b)
	}
	e.trace("> HCI Event: no handler for %s (0x%02X)\n", h.code, uint8(h.code))
	return nil
}

func (e *Evt) trace(fmt string, v ...interface{}) {}

const (
	InquiryComplete                              = 0x01 // Inquiry Complete
	InquiryResult                                = 0x02 // Inquiry Result
	ConnectionComplete                           = 0x03 // Connection Complete
	ConnectionRequest                            = 0x04 // Connection Request
	DisconnectionComplete                        = 0x05 // Disconnection Complete
	AuthenticationComplete                       = 0x06 // Authentication
	RemoteNameReqComplete                        = 0x07 // Remote Name Request Complete
	EncryptionChange                             = 0x08 // Encryption Change
	ChangeConnectionLinkKeyComplete              = 0x09 // Change Conection Link Key Complete
	MasterLinkKeyComplete                        = 0x0A // Master Link Keye Complete
	ReadRemoteSupportedFeaturesComplete          = 0x0B // Read Remote Supported Features Complete
	ReadRemoteVersionInformationComplete         = 0x0C // Read Remote Version Information Complete
	QoSSetupComplete                             = 0x0D // QoSSetupComplete
	CommandComplete                              = 0x0E // Command Complete
	CommandStatus                                = 0x0F // Command status
	HardwareError                                = 0x10 // Hardware Error
	FlushOccurred                                = 0x11 // Flush Occured
	RoleChange                                   = 0x12 // Role Change
	NumberOfCompletedPkts                        = 0x13 // Number Of Completed Packets
	ModeChange                                   = 0x14 // Mode Change
	ReturnLinkKeys                               = 0x15 // Return Link Keys
	PinCodeRequest                               = 0x16 // PIN Code Request
	LinkKeyRequest                               = 0x17 // Link Key Request
	LinkKeyNotification                          = 0x18 // Link Key Notification
	LoopbackCommand                              = 0x19 // Loopback Command
	DataBufferOverflow                           = 0x1A // Data Buffer Overflow
	MaxSlotsChange                               = 0x1B // Max Slots Change
	ReadClockOffsetComplete                      = 0x1C // Read Clock Offset Complete
	ConnectionPtypeChanged                       = 0x1D // Connection Packet Type Changed
	QoSViolation                                 = 0x1E // QoS Violation
	PageScanRepetitionModeChange                 = 0x20 // Page Scan Repetition Mode Change
	FlowSpecificationComplete                    = 0x21 // Flow Specification
	InquiryResultWithRssi                        = 0x22 // Inquery Result with RSSI
	ReadRemoteExtendedFeaturesComplete           = 0x23 // Read Remote Extended Features Complete
	SyncConnectionComplete                       = 0x2C // Synchronous Connection Complete
	SyncConnectionChanged                        = 0x2D // Synchronous Connection Changed
	SniffSubrating                               = 0x2E // Sniff Subrating
	ExtendedInquiryResult                        = 0x2F // Extended Inquiry Result
	EncryptionKeyRefreshComplete                 = 0x30 // Encryption Key Refresh Complete
	IOCapabilityRequest                          = 0x31 // IO Capability Request
	IOCapabilityResponse                         = 0x32 // IO Capability Changed
	UserConfirmationRequest                      = 0x33 // User Confirmation Request
	UserPasskeyRequest                           = 0x34 // User Passkey Request
	RemoteOOBDataRequest                         = 0x35 // Remote OOB Data
	SimplePairingComplete                        = 0x36 // Simple Pairing Complete
	LinkSupervisionTimeoutChanged                = 0x38 // Link Supervision Timeout Changed
	EnhancedFlushComplete                        = 0x39 // Enhanced Flush Complete
	UserPasskeyNotify                            = 0x3B // User Passkey Notification
	KeypressNotify                               = 0x3C // Keypass Notification
	RemoteHostFeaturesNotify                     = 0x3D // Remote Host Supported Features Notification
	LEMeta                                       = 0x3E // LE Meta
	PhysicalLinkComplete                         = 0x40 // Physical Link Complete
	ChannelSelected                              = 0x41 // Channel Selected
	DisconnectionPhysicalLinkComplete            = 0x42 // Disconnection Physical Link Complete
	PhysicalLinkLossEarlyWarning                 = 0x43 // Physical Link Loss Early Warning
	PhysicalLinkRecovery                         = 0x44 // Physical Link Recovery
	LogicalLinkComplete                          = 0x45 // Logical Link Complete
	DisconnectionLogicalLinkComplete             = 0x46 // Disconnection Logical Link Complete
	FlowSpecModifyComplete                       = 0x47 // Flow Spec Modify Complete
	NumberOfCompletedBlocks                      = 0x48 // Number Of Completed Data Blocks
	AMPStartTest                                 = 0x49 // AMP Start Test
	AMPTestEnd                                   = 0x4A // AMP Test End
	AMPReceiverReport                            = 0x4b // AMP Receiver Report
	AMPStatusChange                              = 0x4D // AMP status Change
	TriggeredClockCapture                        = 0x4e // Triggered Clock Capture
	SynchronizationTrainComplete                 = 0x4F // Synchronization Train Complete
	SynchronizationTrainReceived                 = 0x50 // Synchronization Train Received
	ConnectionlessSlaveBroadcastReceive          = 0x51 // Connectionless Slave Broadcast Receive
	ConnectionlessSlaveBroadcastTimeout          = 0x52 // Connectionless Slave Broadcast Timeout
	TruncatedPageComplete                        = 0x53 // Truncated Page Complete
	SlavePageResponseTimeout                     = 0x54 // Slave Page Response Timeout
	ConnectionlessSlaveBroadcastChannelMapChange = 0x55 // Connectionless Slave Broadcast Channel Map Change
	InquiryResponseNotification                  = 0x56 // Inquiry Response Notification
	AuthenticatedPayloadTimeoutExpired           = 0x57 // Authenticated Payload Timeout Expired
)

type LEEventCode int

const (
	LEConnectionComplete               LEEventCode = 0x01 // LE Connection Complete
	LEAdvertisingReport                            = 0x02 // LE Advertising Report
	LEConnectionUpdateComplete                     = 0x03 // LE Connection Update Complete
	LEReadRemoteUsedFeaturesComplete               = 0x04 // LE Read Remote Used Features Complete
	LELTKRequest                                   = 0x05 // LE LTK Request
	LERemoteConnectionParameterRequest             = 0x06 // LE Remote Connection Parameter Request
)

type EventHeader struct {
	code int
	plen uint8
}

func (h *EventHeader) Unmarshal(b []byte) error {
	if len(b) < 2 {
		return errors.New("malformed header")
	}
	h.code = int(b[0])
	h.plen = b[1]
	if uint8(len(b)) != 2+h.plen {
		return errors.New("wrong length")
	}
	return nil
}

var o = util.Order

// Event Parameters

type InquiryCompleteEP struct {
	Status uint8
}

type InquiryResultEP struct {
	NumResponses           uint8
	BDAddr                 [][6]byte
	PageScanRepetitionMode []uint8
	Reserved1              []byte
	Reserved2              []byte
	ClassOfDevice          [][3]byte
	ClockOffset            []uint16
}

type ConnectionCompleteEP struct {
	Status            uint8
	ConnectionHandle  uint16
	BDAddr            [6]byte
	LinkType          uint8
	EncryptionEnabled uint8
}

type ConnectionRequestEP struct {
	BDAddr        [6]byte
	ClassofDevice [3]byte
	LinkType      uint8
}

type DisconnectionCompleteEP struct {
	Status           uint8
	ConnectionHandle uint16
	Reason           uint8
}

func (e *DisconnectionCompleteEP) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.LittleEndian, &e.Status)
	binary.Read(buf, binary.LittleEndian, &e.ConnectionHandle)
	return binary.Read(buf, binary.LittleEndian, &e.Reason)
}

type CommandCompleteEP struct {
	NumHCICommandPackets uint8
	CommandOPCode        uint16
	ReturnParameters     []byte
}

func (e *CommandCompleteEP) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.LittleEndian, &e.NumHCICommandPackets); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &e.CommandOPCode); err != nil {
		return err
	}
	e.ReturnParameters = buf.Bytes()
	return nil
}

type CommandStatusEP struct {
	Status               uint8
	NumHCICommandPackets uint8
	CommandOpcode        uint16
}

func (e *CommandStatusEP) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.LittleEndian, &e.Status)
	binary.Read(buf, binary.LittleEndian, &e.NumHCICommandPackets)
	return binary.Read(buf, binary.LittleEndian, &e.CommandOpcode)
}

type NumOfCompletedPkt struct {
	ConnectionHandle   uint16
	NumOfCompletedPkts uint16
}

type NumberOfCompletedPktsEP struct {
	NumberOfHandles uint8
	Packets         []NumOfCompletedPkt
}

func (e *NumberOfCompletedPktsEP) Unmarshal(b []byte) error {
	e.NumberOfHandles = b[0]
	n := int(e.NumberOfHandles)
	buf := bytes.NewBuffer(b[1:])
	e.Packets = make([]NumOfCompletedPkt, n)
	for i := 0; i < n; i++ {
		binary.Read(buf, binary.LittleEndian, &e.Packets[i].ConnectionHandle)
		binary.Read(buf, binary.LittleEndian, &e.Packets[i].NumOfCompletedPkts)

		e.Packets[i].ConnectionHandle &= 0xfff
	}
	return nil
}

// LE Meta Subevents
type LEConnectionCompleteEP struct {
	SubeventCode        uint8
	Status              uint8
	ConnectionHandle    uint16
	Role                uint8
	PeerAddressType     uint8
	PeerAddress         [6]byte
	ConnInterval        uint16
	ConnLatency         uint16
	SupervisionTimeout  uint16
	MasterClockAccuracy uint8
}

func (e *LEConnectionCompleteEP) Unmarshal(b []byte) error {
	if len(b) < 18 {
		return fmt.Errorf("expected at least 18 bytes, got %d", len(b))
	}
	e.SubeventCode = o.Uint8(b[0:])
	e.Status = o.Uint8(b[1:])
	e.ConnectionHandle = o.Uint16(b[2:])
	e.Role = o.Uint8(b[4:])
	e.PeerAddressType = o.Uint8(b[5:])
	e.PeerAddress = o.MAC(b[6:])
	e.ConnInterval = o.Uint16(b[12:])
	e.ConnLatency = o.Uint16(b[14:])
	e.SupervisionTimeout = o.Uint16(b[16:])
	e.MasterClockAccuracy = o.Uint8(b[17:])
	return nil
}

type LEAdvertisingReportEP struct {
	SubeventCode uint8
	NumReports   uint8
	EventType    []uint8
	AddressType  []uint8
	Address      [][6]byte
	Length       []uint8
	Data         [][]byte
	RSSI         []int8
}

func (e *LEAdvertisingReportEP) Unmarshal(b []byte) error {
	if len(b) < 2 {
		return errors.New("expected at least 2 bytes")
	}
	e.SubeventCode = o.Uint8(b)
	b = b[1:]
	e.NumReports = o.Uint8(b)
	b = b[1:]
	n := int(e.NumReports)

	e.EventType = make([]uint8, n)
	e.AddressType = make([]uint8, n)
	e.Address = make([][6]byte, n)
	e.Length = make([]uint8, n)
	e.Data = make([][]byte, n)
	e.RSSI = make([]int8, n)

	if len(b) < (1+1+6+1)*n {
		return fmt.Errorf("expected %d more bytes, got %d", (1+1+6+1)*n, len(b))
	}

	for i := 0; i < n; i++ {
		e.EventType[i] = o.Uint8(b)
		b = b[1:]
	}
	for i := 0; i < n; i++ {
		e.AddressType[i] = o.Uint8(b)
		b = b[1:]
	}
	for i := 0; i < n; i++ {
		e.Address[i] = o.MAC(b)
		b = b[6:]
	}
	var sumLength int
	for i := 0; i < n; i++ {
		e.Length[i] = o.Uint8(b)
		sumLength += int(e.Length[i])
		b = b[1:]
	}

	if len(b) < sumLength+(1)*n {
		return fmt.Errorf("expected %d more bytes, got %d", sumLength+(1)*n, len(b))
	}

	for i := 0; i < n; i++ {
		e.Data[i] = make([]byte, e.Length[i])
		copy(e.Data[i], b)
		b = b[e.Length[i]:]
	}
	for i := 0; i < n; i++ {
		e.RSSI[i] = o.Int8(b)
		b = b[1:]
	}
	return nil
}

type LEConnectionUpdateCompleteEP struct {
	SubeventCode       uint8
	Status             uint8
	ConnectionHandle   uint16
	ConnInterval       uint16
	ConnLatency        uint16
	SupervisionTimeout uint16
}

func (e *LEConnectionUpdateCompleteEP) Unmarshal(b []byte) error {
	return binary.Read(bytes.NewBuffer(b), binary.LittleEndian, e)
}

type LEReadRemoteUsedFeaturesCompleteEP struct {
	SubeventCode     uint8
	Status           uint8
	ConnectionHandle uint16
	LEFeatures       uint64
}

func (e *LEReadRemoteUsedFeaturesCompleteEP) Unmarshal(b []byte) error {
	return binary.Read(bytes.NewBuffer(b), binary.LittleEndian, e)
}

type LELTKRequestEP struct {
	SubeventCode          uint8
	ConnectionHandle      uint16
	RandomNumber          uint64
	EncryptionDiversifier uint16
}

func (e *LELTKRequestEP) Unmarshal(b []byte) error {
	return binary.Read(bytes.NewBuffer(b), binary.LittleEndian, e)
}

type LERemoteConnectionParameterRequestEP struct {
	SubeventCode     uint8
	ConnectionHandle uint16
	IntervalMin      uint16
	IntervalMax      uint16
	Latency          uint16
	Timeout          uint16
}

func (e *LERemoteConnectionParameterRequestEP) Unmarshal(b []byte) error {
	return binary.Read(bytes.NewBuffer(b), binary.LittleEndian, e)
}
