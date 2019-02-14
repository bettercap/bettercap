package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/bettercap/gatt/linux/evt"
	"github.com/bettercap/gatt/linux/util"
)

type CmdParam interface {
	Marshal([]byte)
	Opcode() int
	Len() int
}

func NewCmd(d io.Writer) *Cmd {
	c := &Cmd{
		dev:     d,
		sent:    []*cmdPkt{},
		compc:   make(chan evt.CommandCompleteEP),
		statusc: make(chan evt.CommandStatusEP),
	}
	go c.processCmdEvents()
	return c
}

type cmdPkt struct {
	op   int
	cp   CmdParam
	done chan []byte
}

func (c cmdPkt) Marshal() []byte {
	b := make([]byte, 1+2+1+c.cp.Len())
	b[0] = byte(0x1) // typCommandPkt
	b[1] = byte(c.op)
	b[2] = byte(c.op >> 8)
	b[3] = byte(c.cp.Len())
	c.cp.Marshal(b[4:])
	return b
}

type Cmd struct {
	dev     io.Writer
	sent    []*cmdPkt
	compc   chan evt.CommandCompleteEP
	statusc chan evt.CommandStatusEP
}

func (c Cmd) trace(fmt string, v ...interface{}) {}

func (c *Cmd) HandleComplete(b []byte) error {
	var e evt.CommandCompleteEP
	if err := e.Unmarshal(b); err != nil {
		return err
	}
	c.compc <- e
	return nil
}

func (c *Cmd) HandleStatus(b []byte) error {
	var e evt.CommandStatusEP
	if err := e.Unmarshal(b); err != nil {
		return err
	}
	c.statusc <- e
	return nil
}

func (c *Cmd) Send(cp CmdParam) ([]byte, error) {
	op := cp.Opcode()
	p := &cmdPkt{op: op, cp: cp, done: make(chan []byte)}
	raw := p.Marshal()

	c.sent = append(c.sent, p)
	if n, err := c.dev.Write(raw); err != nil {
		return nil, err
	} else if n != len(raw) {
		return nil, errors.New("Failed to send whole Cmd pkt to HCI socket")
	}
	return <-p.done, nil
}

func (c *Cmd) SendAndCheckResp(cp CmdParam, exp []byte) error {
	rsp, err := c.Send(cp)
	if err != nil {
		return err
	}
	// Don't care about the response
	if len(exp) == 0 {
		return nil
	}
	// Check the if status is one of the expected value
	if !bytes.Contains(exp, rsp[0:1]) {
		return fmt.Errorf("HCI command: '0x%04x' return 0x%02X, expect: [%X] ", cp.Opcode(), rsp[0], exp)
	}
	return nil
}

func (c *Cmd) processCmdEvents() {
	for {
		select {
		case status := <-c.statusc:
			found := false
			for i, p := range c.sent {
				if uint16(p.op) == status.CommandOpcode {
					found = true
					c.sent = append(c.sent[:i], c.sent[i+1:]...)
					close(p.done)
					break
				}
			}
			if !found {
				log.Printf("Can't find the cmdPkt for this CommandStatusEP: %v", status)
			}
		case comp := <-c.compc:
			found := false
			for i, p := range c.sent {
				if uint16(p.op) == comp.CommandOPCode {
					found = true
					c.sent = append(c.sent[:i], c.sent[i+1:]...)
					p.done <- comp.ReturnParameters
					break
				}
			}
			if !found {
				log.Printf("Can't find the cmdPkt for this CommandCompleteEP: %v", comp)
			}
		}
	}
}

const (
	linkCtl     = 0x01
	linkPolicy  = 0x02
	hostCtl     = 0x03
	infoParam   = 0x04
	statusParam = 0x05
	testingCmd  = 0X3E
	leCtl       = 0x08
	vendorCmd   = 0X3F
)

const (
	opInquiry                = linkCtl<<10 | 0x0001 // Inquiry
	opInquiryCancel          = linkCtl<<10 | 0x0002 // Inquiry Cancel
	opPeriodicInquiry        = linkCtl<<10 | 0x0003 // Periodic Inquiry Mode
	opExitPeriodicInquiry    = linkCtl<<10 | 0x0004 // Exit Periodic Inquiry Mode
	opCreateConn             = linkCtl<<10 | 0x0005 // Create Connection
	opDisconnect             = linkCtl<<10 | 0x0006 // Disconnect
	opCreateConnCancel       = linkCtl<<10 | 0x0008 // Create Connection Cancel
	opAcceptConnReq          = linkCtl<<10 | 0x0009 // Accept Connection Request
	opRejectConnReq          = linkCtl<<10 | 0x000A // Reject Connection Request
	opLinkKeyReply           = linkCtl<<10 | 0x000B // Link Key Request Reply
	opLinkKeyNegReply        = linkCtl<<10 | 0x000C // Link Key Request Negative Reply
	opPinCodeReply           = linkCtl<<10 | 0x000D // PIN Code Request Reply
	opPinCodeNegReply        = linkCtl<<10 | 0x000E // PIN Code Request Negative Reply
	opSetConnPtype           = linkCtl<<10 | 0x000F // Change Connection Packet Type
	opAuthRequested          = linkCtl<<10 | 0x0011 // Authentication Request
	opSetConnEncrypt         = linkCtl<<10 | 0x0013 // Set Connection Encryption
	opChangeConnLinkKey      = linkCtl<<10 | 0x0015 // Change Connection Link Key
	opMasterLinkKey          = linkCtl<<10 | 0x0017 // Master Link Key
	opRemoteNameReq          = linkCtl<<10 | 0x0019 // Remote Name Request
	opRemoteNameReqCancel    = linkCtl<<10 | 0x001A // Remote Name Request Cancel
	opReadRemoteFeatures     = linkCtl<<10 | 0x001B // Read Remote Supported Features
	opReadRemoteExtFeatures  = linkCtl<<10 | 0x001C // Read Remote Extended Features
	opReadRemoteVersion      = linkCtl<<10 | 0x001D // Read Remote Version Information
	opReadClockOffset        = linkCtl<<10 | 0x001F // Read Clock Offset
	opReadLMPHandle          = linkCtl<<10 | 0x0020 // Read LMP Handle
	opSetupSyncConn          = linkCtl<<10 | 0x0028 // Setup Synchronous Connection
	opAcceptSyncConnReq      = linkCtl<<10 | 0x0029 // Aceept Synchronous Connection
	opRejectSyncConnReq      = linkCtl<<10 | 0x002A // Recject Synchronous Connection
	opIOCapabilityReply      = linkCtl<<10 | 0x002B // IO Capability Request Reply
	opUserConfirmReply       = linkCtl<<10 | 0x002C // User Confirmation Request Reply
	opUserConfirmNegReply    = linkCtl<<10 | 0x002D // User Confirmation Negative Reply
	opUserPasskeyReply       = linkCtl<<10 | 0x002E // User Passkey Request Reply
	opUserPasskeyNegReply    = linkCtl<<10 | 0x002F // User Passkey Request Negative Reply
	opRemoteOOBDataReply     = linkCtl<<10 | 0x0030 // Remote OOB Data Request Reply
	opRemoteOOBDataNegReply  = linkCtl<<10 | 0x0033 // Remote OOB Data Request Negative Reply
	opIOCapabilityNegReply   = linkCtl<<10 | 0x0034 // IO Capability Request Negative Reply
	opCreatePhysicalLink     = linkCtl<<10 | 0x0035 // Create Physical Link
	opAcceptPhysicalLink     = linkCtl<<10 | 0x0036 // Accept Physical Link
	opDisconnectPhysicalLink = linkCtl<<10 | 0x0037 // Disconnect Physical Link
	opCreateLogicalLink      = linkCtl<<10 | 0x0038 // Create Logical Link
	opAcceptLogicalLink      = linkCtl<<10 | 0x0039 // Accept Logical Link
	opDisconnectLogicalLink  = linkCtl<<10 | 0x003A // Disconnect Logical Link
	opLogicalLinkCancel      = linkCtl<<10 | 0x003B // Logical Link Cancel
	opFlowSpecModify         = linkCtl<<10 | 0x003C // Flow Spec Modify
)

const (
	opHoldMode               = linkPolicy<<10 | 0x0001 // Hold Mode
	opSniffMode              = linkPolicy<<10 | 0x0003 // Sniff Mode
	opExitSniffMode          = linkPolicy<<10 | 0x0004 // Exit Sniff Mode
	opParkMode               = linkPolicy<<10 | 0x0005 // Park State
	opExitParkMode           = linkPolicy<<10 | 0x0006 // Exit Park State
	opQoSSetup               = linkPolicy<<10 | 0x0007 // QoS Setup
	opRoleDiscovery          = linkPolicy<<10 | 0x0009 // Role Discovery
	opSwitchRole             = linkPolicy<<10 | 0x000B // Switch Role
	opReadLinkPolicy         = linkPolicy<<10 | 0x000C // Read Link Policy Settings
	opWriteLinkPolicy        = linkPolicy<<10 | 0x000D // Write Link Policy Settings
	opReadDefaultLinkPolicy  = linkPolicy<<10 | 0x000E // Read Default Link Policy Settings
	opWriteDefaultLinkPolicy = linkPolicy<<10 | 0x000F // Write Default Link Policy Settings
	opFlowSpecification      = linkPolicy<<10 | 0x0010 // Flow Specification
	opSniffSubrating         = linkPolicy<<10 | 0x0011 // Sniff Subrating
)

const (
	opSetEventMask                      = hostCtl<<10 | 0x0001 // Set Event Mask
	opReset                             = hostCtl<<10 | 0x0003 // Reset
	opSetEventFlt                       = hostCtl<<10 | 0x0005 // Set Event Filter
	opFlush                             = hostCtl<<10 | 0x0008 // Flush
	opReadPinType                       = hostCtl<<10 | 0x0009 // Read PIN Type
	opWritePinType                      = hostCtl<<10 | 0x000A // Write PIN Type
	opCreateNewUnitKey                  = hostCtl<<10 | 0x000B // Create New Unit Key
	opReadStoredLinkKey                 = hostCtl<<10 | 0x000D // Read Stored Link Key
	opWriteStoredLinkKey                = hostCtl<<10 | 0x0011 // Write Stored Link Key
	opDeleteStoredLinkKey               = hostCtl<<10 | 0x0012 // Delete Stored Link Key
	opWriteLocalName                    = hostCtl<<10 | 0x0013 // Write Local Name
	opReadLocalName                     = hostCtl<<10 | 0x0014 // Read Local Name
	opReadConnAcceptTimeout             = hostCtl<<10 | 0x0015 // Read Connection Accept Timeout
	opWriteConnAcceptTimeout            = hostCtl<<10 | 0x0016 // Write Connection Accept Timeout
	opReadPageTimeout                   = hostCtl<<10 | 0x0017 // Read Page Timeout
	opWritePageTimeout                  = hostCtl<<10 | 0x0018 // Write Page Timeout
	opReadScanEnable                    = hostCtl<<10 | 0x0019 // Read Scan Enable
	opWriteScanEnable                   = hostCtl<<10 | 0x001A // Write Scan Enable
	opReadPageActivity                  = hostCtl<<10 | 0x001B // Read Page Scan Activity
	opWritePageActivity                 = hostCtl<<10 | 0x001C // Write Page Scan Activity
	opReadInqActivity                   = hostCtl<<10 | 0x001D // Read Inquiry Scan Activity
	opWriteInqActivity                  = hostCtl<<10 | 0x001E // Write Inquiry Scan Activity
	opReadAuthEnable                    = hostCtl<<10 | 0x001F // Read Authentication Enable
	opWriteAuthEnable                   = hostCtl<<10 | 0x0020 // Write Authentication Enable
	opReadEncryptMode                   = hostCtl<<10 | 0x0021
	opWriteEncryptMode                  = hostCtl<<10 | 0x0022
	opReadClassOfDev                    = hostCtl<<10 | 0x0023 // Read Class of Device
	opWriteClassOfDevice                = hostCtl<<10 | 0x0024 // Write Class of Device
	opReadVoiceSetting                  = hostCtl<<10 | 0x0025 // Read Voice Setting
	opWriteVoiceSetting                 = hostCtl<<10 | 0x0026 // Write Voice Setting
	opReadAutomaticFlushTimeout         = hostCtl<<10 | 0x0027 // Read Automatic Flush Timeout
	opWriteAutomaticFlushTimeout        = hostCtl<<10 | 0x0028 // Write Automatic Flush Timeout
	opReadNumBroadcastRetrans           = hostCtl<<10 | 0x0029 // Read Num Broadcast Retransmissions
	opWriteNumBroadcastRetrans          = hostCtl<<10 | 0x002A // Write Num Broadcast Retransmissions
	opReadHoldModeActivity              = hostCtl<<10 | 0x002B // Read Hold Mode Activity
	opWriteHoldModeActivity             = hostCtl<<10 | 0x002C // Write Hold Mode Activity
	opReadTransmitPowerLevel            = hostCtl<<10 | 0x002D // Read Transmit Power Level
	opReadSyncFlowEnable                = hostCtl<<10 | 0x002E // Read Synchronous Flow Control
	opWriteSyncFlowEnable               = hostCtl<<10 | 0x002F // Write Synchronous Flow Control
	opSetControllerToHostFC             = hostCtl<<10 | 0x0031 // Set Controller To Host Flow Control
	opHostBufferSize                    = hostCtl<<10 | 0x0033 // Host Buffer Size
	opHostNumCompPkts                   = hostCtl<<10 | 0x0035 // Host Number Of Completed Packets
	opReadLinkSupervisionTimeout        = hostCtl<<10 | 0x0036 // Read Link Supervision Timeout
	opWriteLinkSupervisionTimeout       = hostCtl<<10 | 0x0037 // Write Link Supervision Timeout
	opReadNumSupportedIAC               = hostCtl<<10 | 0x0038 // Read Number Of Supported IAC
	opReadCurrentIACLAP                 = hostCtl<<10 | 0x0039 // Read Current IAC LAP
	opWriteCurrentIACLAP                = hostCtl<<10 | 0x003A // Write Current IAC LAP
	opReadPageScanPeriodMode            = hostCtl<<10 | 0x003B
	opWritePageScanPeriodMode           = hostCtl<<10 | 0x003C
	opReadPageScanMode                  = hostCtl<<10 | 0x003D
	opWritePageScanMode                 = hostCtl<<10 | 0x003E
	opSetAFHClassification              = hostCtl<<10 | 0x003F // Set AFH Host Channel Classification
	opReadInquiryScanType               = hostCtl<<10 | 0x0042 // Read Inquiry Scan Type
	opWriteInquiryScanType              = hostCtl<<10 | 0x0043 // Write Inquiry Scan Type
	opReadInquiryMode                   = hostCtl<<10 | 0x0044 // Read Inquiry Mode
	opWriteInquiryMode                  = hostCtl<<10 | 0x0045 // Write Inquiry Mode
	opReadPageScanType                  = hostCtl<<10 | 0x0046 // Read Page Scan Type
	opWritePageScanType                 = hostCtl<<10 | 0x0047 // Write Page Scan Type
	opReadAFHMode                       = hostCtl<<10 | 0x0048 // Read AFH Channel Assessment Mode
	opWriteAFHMode                      = hostCtl<<10 | 0x0049 // Write AFH Channel Assesment Mode
	opReadExtInquiryResponse            = hostCtl<<10 | 0x0051 // Read Extended Inquiry Response
	opWriteExtInquiryResponse           = hostCtl<<10 | 0x0052 // Write Extended Inquiry Response
	opRefreshEncryptionKey              = hostCtl<<10 | 0x0053 // Refresh Encryption Key
	opReadSimplePairingMode             = hostCtl<<10 | 0x0055 // Read Simple Pairing Mode
	opWriteSimplePairingMode            = hostCtl<<10 | 0x0056 // Write Simple Pairing Mode
	opReadLocalOobData                  = hostCtl<<10 | 0x0057 // Read Local OOB Data
	opReadInqResponseTransmitPowerLevel = hostCtl<<10 | 0x0058 // Read Inquiry Response Transmit Power Level
	opWriteInquiryTransmitPowerLevel    = hostCtl<<10 | 0x0059 // Write Inquiry Response Transmit Power Level
	opReadDefaultErrorDataReporting     = hostCtl<<10 | 0x005A // Read Default Erroneous Data Reporting
	opWriteDefaultErrorDataReporting    = hostCtl<<10 | 0x005B // Write Default Erroneous Data Reporting
	opEnhancedFlush                     = hostCtl<<10 | 0x005F // Enhanced Flush
	opSendKeypressNotify                = hostCtl<<10 | 0x0060 // send Keypress Notification
	opReadLogicalLinkAcceptTimeout      = hostCtl<<10 | 0x0061 // Read Logical Link Accept Timeout
	opWriteLogicalLinkAcceptTimeout     = hostCtl<<10 | 0x0062 // Write Logical Link Accept Timeout
	opSetEventMaskPage2                 = hostCtl<<10 | 0x0063 // Set Event Mask Page 2
	opReadLocationData                  = hostCtl<<10 | 0x0064 // Read Location Data
	opWriteLocationData                 = hostCtl<<10 | 0x0065 // Write Location Data
	opReadFlowControlMode               = hostCtl<<10 | 0x0066 // Read Flow Control Mode
	opWriteFlowControlMode              = hostCtl<<10 | 0x0067 // Write Flow Control Mode
	opReadEnhancedTransmitpowerLevel    = hostCtl<<10 | 0x0068 // Read Enhanced Transmit Power Level
	opReadBestEffortFlushTimeout        = hostCtl<<10 | 0x0069 // Read Best Effort Flush Timeout
	opWriteBestEffortFlushTimeout       = hostCtl<<10 | 0x006A // Write Best Effort Flush Timeout
	opReadLEHostSupported               = hostCtl<<10 | 0x006C // Read LE Host Supported
	opWriteLEHostSupported              = hostCtl<<10 | 0x006D // Write LE Host Supported
)
const (
	opReadLocalVersionInformation = infoParam<<10 | 0x0001 // Read Local Version Information
	opReadLocalSupportedCommands  = infoParam<<10 | 0x0002 // Read Local Supported Commands
	opReadLocalSupportedFeatures  = infoParam<<10 | 0x0003 // Read Local Supported Features
	opReadLocalExtendedFeatures   = infoParam<<10 | 0x0004 // Read Local Extended Features
	opReadBufferSize              = infoParam<<10 | 0x0005 // Read Buffer Size
	opReadBDADDR                  = infoParam<<10 | 0x0009 // Read BD_ADDR
	opReadDataBlockSize           = infoParam<<10 | 0x000A // Read Data Block Size
	opReadLocalSupportedCodecs    = infoParam<<10 | 0x000B // Read Local Supported Codecs
)
const (
	opLESetEventMask                      = leCtl<<10 | 0x0001 // LE Set Event Mask
	opLEReadBufferSize                    = leCtl<<10 | 0x0002 // LE Read Buffer Size
	opLEReadLocalSupportedFeatures        = leCtl<<10 | 0x0003 // LE Read Local Supported Features
	opLESetRandomAddress                  = leCtl<<10 | 0x0005 // LE Set Random Address
	opLESetAdvertisingParameters          = leCtl<<10 | 0x0006 // LE Set Advertising Parameters
	opLEReadAdvertisingChannelTxPower     = leCtl<<10 | 0x0007 // LE Read Advertising Channel Tx Power
	opLESetAdvertisingData                = leCtl<<10 | 0x0008 // LE Set Advertising Data
	opLESetScanResponseData               = leCtl<<10 | 0x0009 // LE Set Scan Response Data
	opLESetAdvertiseEnable                = leCtl<<10 | 0x000a // LE Set Advertising Enable
	opLESetScanParameters                 = leCtl<<10 | 0x000b // LE Set Scan Parameters
	opLESetScanEnable                     = leCtl<<10 | 0x000c // LE Set Scan Enable
	opLECreateConn                        = leCtl<<10 | 0x000d // LE Create Connection
	opLECreateConnCancel                  = leCtl<<10 | 0x000e // LE Create Connection Cancel
	opLEReadWhiteListSize                 = leCtl<<10 | 0x000f // LE Read White List Size
	opLEClearWhiteList                    = leCtl<<10 | 0x0010 // LE Clear White List
	opLEAddDeviceToWhiteList              = leCtl<<10 | 0x0011 // LE Add Device To White List
	opLERemoveDeviceFromWhiteList         = leCtl<<10 | 0x0012 // LE Remove Device From White List
	opLEConnUpdate                        = leCtl<<10 | 0x0013 // LE Connection Update
	opLESetHostChannelClassification      = leCtl<<10 | 0x0014 // LE Set Host Channel Classification
	opLEReadChannelMap                    = leCtl<<10 | 0x0015 // LE Read Channel Map
	opLEReadRemoteUsedFeatures            = leCtl<<10 | 0x0016 // LE Read Remote Used Features
	opLEEncrypt                           = leCtl<<10 | 0x0017 // LE Encrypt
	opLERand                              = leCtl<<10 | 0x0018 // LE Rand
	opLEStartEncryption                   = leCtl<<10 | 0x0019 // LE Star Encryption
	opLELTKReply                          = leCtl<<10 | 0x001a // LE Long Term Key Request Reply
	opLELTKNegReply                       = leCtl<<10 | 0x001b // LE Long Term Key Request Negative Reply
	opLEReadSupportedStates               = leCtl<<10 | 0x001c // LE Read Supported States
	opLEReceiverTest                      = leCtl<<10 | 0x001d // LE Reciever Test
	opLETransmitterTest                   = leCtl<<10 | 0x001e // LE Transmitter Test
	opLETestEnd                           = leCtl<<10 | 0x001f // LE Test End
	opLERemoteConnectionParameterReply    = leCtl<<10 | 0x0020 // LE Remote Connection Parameter Request Reply
	opLERemoteConnectionParameterNegReply = leCtl<<10 | 0x0021 // LE Remote Connection Parameter Request Negative Reply
)

var o = util.Order

// Link Control Commands

// Disconnect (0x0006)
type Disconnect struct {
	ConnectionHandle uint16
	Reason           uint8
}

func (c Disconnect) Opcode() int { return opDisconnect }
func (c Disconnect) Len() int    { return 3 }
func (c Disconnect) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	b[2] = c.Reason
}

// No Return Parameters, Check for Disconnection Complete Event
type DisconnectRP struct{}

// Link Policy Commands

// Write Default Link Policy
type WriteDefaultLinkPolicy struct{ DefaultLinkPolicySettings uint16 }

func (c WriteDefaultLinkPolicy) Opcode() int      { return opWriteDefaultLinkPolicy }
func (c WriteDefaultLinkPolicy) Len() int         { return 2 }
func (c WriteDefaultLinkPolicy) Marshal(b []byte) { o.PutUint16(b, c.DefaultLinkPolicySettings) }

type WriteDefaultLinkPolicyRP struct{ Status uint8 }

// Host Control Commands

// Set Event Mask (0x0001)
type SetEventMask struct{ EventMask uint64 }

func (c SetEventMask) Opcode() int      { return opSetEventMask }
func (c SetEventMask) Len() int         { return 8 }
func (c SetEventMask) Marshal(b []byte) { o.PutUint64(b, c.EventMask) }

type SetEventMaskRP struct{ Status uint8 }

// Reset (0x0002)
type Reset struct{}

func (c Reset) Opcode() int      { return opReset }
func (c Reset) Len() int         { return 0 }
func (c Reset) Marshal(b []byte) {}

type ResetRP struct{ Status uint8 }

// Set Event Filter (0x0003)
// FIXME: This structures are overloading.
// Both Marshal() and Len() are just placeholder.
// Need more effort for decoding.
// type SetEventFlt struct {
// 	FilterType          uint8
// 	FilterConditionType uint8
// 	Condition           uint8
// }

// func (c SetEventFlt) Opcode() int   { return opSetEventFlt }
// func (c SetEventFlt) Len() int         { return 0 }
// func (c SetEventFlt) Marshal(b []byte) {}

type SetEventFltRP struct{ Status uint8 }

// Flush (0x0008)
type Flush struct{ ConnectionHandle uint16 }

func (c Flush) Opcode() int      { return opFlush }
func (c Flush) Len() int         { return 2 }
func (c Flush) Marshal(b []byte) { o.PutUint16(b, c.ConnectionHandle) }

type flushRP struct{ status uint8 }

// Write Page Timeout (0x0018)
type WritePageTimeout struct{ PageTimeout uint16 }

func (c WritePageTimeout) Opcode() int      { return opWritePageTimeout }
func (c WritePageTimeout) Len() int         { return 2 }
func (c WritePageTimeout) Marshal(b []byte) { o.PutUint16(b, c.PageTimeout) }

type WritePageTimeoutRP struct{}

// Write Class of Device (0x0024)
type WriteClassOfDevice struct{ ClassOfDevice [3]byte }

func (c WriteClassOfDevice) Opcode() int      { return opWriteClassOfDevice }
func (c WriteClassOfDevice) Len() int         { return 3 }
func (c WriteClassOfDevice) Marshal(b []byte) { copy(b, c.ClassOfDevice[:]) }

type WriteClassOfDevRP struct{ status uint8 }

// Write Host Buffer Size (0x0033)
type HostBufferSize struct {
	HostACLDataPacketLength            uint16
	HostSynchronousDataPacketLength    uint8
	HostTotalNumACLDataPackets         uint16
	HostTotalNumSynchronousDataPackets uint16
}

func (c HostBufferSize) Opcode() int { return opHostBufferSize }
func (c HostBufferSize) Len() int    { return 7 }
func (c HostBufferSize) Marshal(b []byte) {
	o.PutUint16(b[0:], c.HostACLDataPacketLength)
	o.PutUint8(b[2:], c.HostSynchronousDataPacketLength)
	o.PutUint16(b[3:], c.HostTotalNumACLDataPackets)
	o.PutUint16(b[5:], c.HostTotalNumSynchronousDataPackets)
}

type HostBufferSizeRP struct{ Status uint8 }

// Write Inquiry Scan Type (0x0043)
type WriteInquiryScanType struct{ ScanType uint8 }

func (c WriteInquiryScanType) Opcode() int      { return opWriteInquiryScanType }
func (c WriteInquiryScanType) Len() int         { return 1 }
func (c WriteInquiryScanType) Marshal(b []byte) { b[0] = c.ScanType }

type WriteInquiryScanTypeRP struct{ Status uint8 }

// Write Inquiry Mode (0x0045)
type WriteInquiryMode struct {
	InquiryMode uint8
}

func (c WriteInquiryMode) Opcode() int      { return opWriteInquiryMode }
func (c WriteInquiryMode) Len() int         { return 1 }
func (c WriteInquiryMode) Marshal(b []byte) { b[0] = c.InquiryMode }

type WriteInquiryModeRP struct{ Status uint8 }

// Write Page Scan Type (0x0046)
type WritePageScanType struct{ PageScanType uint8 }

func (c WritePageScanType) Opcode() int      { return opWritePageScanType }
func (c WritePageScanType) Len() int         { return 1 }
func (c WritePageScanType) Marshal(b []byte) { b[0] = c.PageScanType }

type WritePageScanTypeRP struct{ Status uint8 }

// Write Simple Pairing Mode (0x0056)
type WriteSimplePairingMode struct{ SimplePairingMode uint8 }

func (c WriteSimplePairingMode) Opcode() int      { return opWriteSimplePairingMode }
func (c WriteSimplePairingMode) Len() int         { return 1 }
func (c WriteSimplePairingMode) Marshal(b []byte) { b[0] = c.SimplePairingMode }

type WriteSimplePairingModeRP struct{}

// Set Event Mask Page 2 (0x0063)
type SetEventMaskPage2 struct{ EventMaskPage2 uint64 }

func (c SetEventMaskPage2) Opcode() int      { return opSetEventMaskPage2 }
func (c SetEventMaskPage2) Len() int         { return 8 }
func (c SetEventMaskPage2) Marshal(b []byte) { o.PutUint64(b, c.EventMaskPage2) }

type SetEventMaskPage2RP struct{ Status uint8 }

// Write LE Host Supported (0x006D)
type WriteLEHostSupported struct {
	LESupportedHost    uint8
	SimultaneousLEHost uint8
}

func (c WriteLEHostSupported) Opcode() int      { return opWriteLEHostSupported }
func (c WriteLEHostSupported) Len() int         { return 2 }
func (c WriteLEHostSupported) Marshal(b []byte) { b[0], b[1] = c.LESupportedHost, c.SimultaneousLEHost }

type WriteLeHostSupportedRP struct{ Status uint8 }

// LE Controller Commands

// LE Set Event Mask (0x0001)
type LESetEventMask struct{ LEEventMask uint64 }

func (c LESetEventMask) Opcode() int      { return opLESetEventMask }
func (c LESetEventMask) Len() int         { return 8 }
func (c LESetEventMask) Marshal(b []byte) { o.PutUint64(b, c.LEEventMask) }

type LESetEventMaskRP struct{ Status uint8 }

// LE Read Buffer Size (0x0002)
type LEReadBufferSize struct{}

func (c LEReadBufferSize) Opcode() int      { return opLEReadBufferSize }
func (c LEReadBufferSize) Len() int         { return 1 }
func (c LEReadBufferSize) Marshal(b []byte) {}

type LEReadBufferSizeRP struct {
	Status                     uint8
	HCLEACLDataPacketLength    uint16
	HCTotalNumLEACLDataPackets uint8
}

// LE Read Local Supported Features (0x0003)
type LEReadLocalSupportedFeatures struct{}

func (c LEReadLocalSupportedFeatures) Opcode() int      { return opLEReadLocalSupportedFeatures }
func (c LEReadLocalSupportedFeatures) Len() int         { return 0 }
func (c LEReadLocalSupportedFeatures) Marshal(b []byte) {}

type LEReadLocalSupportedFeaturesRP struct {
	Status     uint8
	LEFeatures uint64
}

// LE Set Random Address (0x0005)
type LESetRandomAddress struct{ RandomAddress [6]byte }

func (c LESetRandomAddress) Opcode() int      { return opLESetRandomAddress }
func (c LESetRandomAddress) Len() int         { return 6 }
func (c LESetRandomAddress) Marshal(b []byte) { o.PutMAC(b, c.RandomAddress) }

type LESetRandomAddressRP struct{ Status uint8 }

// LE Set Advertising Parameters (0x0006)
type LESetAdvertisingParameters struct {
	AdvertisingIntervalMin  uint16
	AdvertisingIntervalMax  uint16
	AdvertisingType         uint8
	OwnAddressType          uint8
	DirectAddressType       uint8
	DirectAddress           [6]byte
	AdvertisingChannelMap   uint8
	AdvertisingFilterPolicy uint8
}

func (c LESetAdvertisingParameters) Opcode() int { return opLESetAdvertisingParameters }
func (c LESetAdvertisingParameters) Len() int    { return 15 }
func (c LESetAdvertisingParameters) Marshal(b []byte) {
	o.PutUint16(b[0:], c.AdvertisingIntervalMin)
	o.PutUint16(b[2:], c.AdvertisingIntervalMax)
	o.PutUint8(b[4:], c.AdvertisingType)
	o.PutUint8(b[5:], c.OwnAddressType)
	o.PutUint8(b[6:], c.DirectAddressType)
	o.PutMAC(b[7:], c.DirectAddress)
	o.PutUint8(b[13:], c.AdvertisingChannelMap)
	o.PutUint8(b[14:], c.AdvertisingFilterPolicy)
}

type LESetAdvertisingParametersRP struct{ Status uint8 }

// LE Read Advertising Channel Tx Power (0x0007)
type LEReadAdvertisingChannelTxPower struct{}

func (c LEReadAdvertisingChannelTxPower) Opcode() int      { return opLEReadAdvertisingChannelTxPower }
func (c LEReadAdvertisingChannelTxPower) Len() int         { return 0 }
func (c LEReadAdvertisingChannelTxPower) Marshal(b []byte) {}

type LEReadAdvertisingChannelTxPowerRP struct {
	Status             uint8
	TransmitPowerLevel uint8
}

// LE Set Advertising Data (0x0008)
type LESetAdvertisingData struct {
	AdvertisingDataLength uint8
	AdvertisingData       [31]byte
}

func (c LESetAdvertisingData) Opcode() int { return opLESetAdvertisingData }
func (c LESetAdvertisingData) Len() int    { return 32 }
func (c LESetAdvertisingData) Marshal(b []byte) {
	b[0] = c.AdvertisingDataLength
	copy(b[1:], c.AdvertisingData[:c.AdvertisingDataLength])
}

type LESetAdvertisingDataRP struct{ Status uint8 }

// LE Set Scan Response Data (0x0009)
type LESetScanResponseData struct {
	ScanResponseDataLength uint8
	ScanResponseData       [31]byte
}

func (c LESetScanResponseData) Opcode() int { return opLESetScanResponseData }
func (c LESetScanResponseData) Len() int    { return 32 }
func (c LESetScanResponseData) Marshal(b []byte) {
	b[0] = c.ScanResponseDataLength
	copy(b[1:], c.ScanResponseData[:c.ScanResponseDataLength])
}

type LESetScanResponseDataRP struct{ Status uint8 }

// LE Set Advertising Enable (0x000A)
type LESetAdvertiseEnable struct{ AdvertisingEnable uint8 }

func (c LESetAdvertiseEnable) Opcode() int      { return opLESetAdvertiseEnable }
func (c LESetAdvertiseEnable) Len() int         { return 1 }
func (c LESetAdvertiseEnable) Marshal(b []byte) { b[0] = c.AdvertisingEnable }

type LESetAdvertiseEnableRP struct{ Status uint8 }

// LE Set Scan Parameters (0x000B)
type LESetScanParameters struct {
	LEScanType           uint8
	LEScanInterval       uint16
	LEScanWindow         uint16
	OwnAddressType       uint8
	ScanningFilterPolicy uint8
}

func (c LESetScanParameters) Opcode() int { return opLESetScanParameters }
func (c LESetScanParameters) Len() int    { return 7 }
func (c LESetScanParameters) Marshal(b []byte) {
	o.PutUint8(b[0:], c.LEScanType)
	o.PutUint16(b[1:], c.LEScanInterval)
	o.PutUint16(b[3:], c.LEScanWindow)
	o.PutUint8(b[5:], c.OwnAddressType)
	o.PutUint8(b[6:], c.ScanningFilterPolicy)
}

type LESetScanParametersRP struct{ Status uint8 }

// LE Set Scan Enable (0x000C)
type LESetScanEnable struct {
	LEScanEnable     uint8
	FilterDuplicates uint8
}

func (c LESetScanEnable) Opcode() int      { return opLESetScanEnable }
func (c LESetScanEnable) Len() int         { return 2 }
func (c LESetScanEnable) Marshal(b []byte) { b[0], b[1] = c.LEScanEnable, c.FilterDuplicates }

type LESetScanEnableRP struct{ Status uint8 }

// LE Create Connection (0x000D)
type LECreateConn struct {
	LEScanInterval        uint16
	LEScanWindow          uint16
	InitiatorFilterPolicy uint8
	PeerAddressType       uint8
	PeerAddress           [6]byte
	OwnAddressType        uint8
	ConnIntervalMin       uint16
	ConnIntervalMax       uint16
	ConnLatency           uint16
	SupervisionTimeout    uint16
	MinimumCELength       uint16
	MaximumCELength       uint16
}

func (c LECreateConn) Opcode() int { return opLECreateConn }
func (c LECreateConn) Len() int    { return 25 }
func (c LECreateConn) Marshal(b []byte) {
	o.PutUint16(b[0:], c.LEScanInterval)
	o.PutUint16(b[2:], c.LEScanWindow)
	o.PutUint8(b[4:], c.InitiatorFilterPolicy)
	o.PutUint8(b[5:], c.PeerAddressType)
	o.PutMAC(b[6:], c.PeerAddress)
	o.PutUint8(b[12:], c.OwnAddressType)
	o.PutUint16(b[13:], c.ConnIntervalMin)
	o.PutUint16(b[15:], c.ConnIntervalMax)
	o.PutUint16(b[17:], c.ConnLatency)
	o.PutUint16(b[19:], c.SupervisionTimeout)
	o.PutUint16(b[21:], c.MinimumCELength)
	o.PutUint16(b[23:], c.MaximumCELength)
}

type LECreateConnRP struct{}

// LE Create Connection Cancel (0x000E)
type LECreateConnCancel struct{}

func (c LECreateConnCancel) Opcode() int      { return opLECreateConnCancel }
func (c LECreateConnCancel) Len() int         { return 0 }
func (c LECreateConnCancel) Marshal(b []byte) {}

type LECreateConnCancelRP struct{ Status uint8 }

// LE Read White List Size (0x000F)
type LEReadWhiteListSize struct{}

func (c LEReadWhiteListSize) Opcode() int      { return opLEReadWhiteListSize }
func (c LEReadWhiteListSize) Len() int         { return 0 }
func (c LEReadWhiteListSize) Marshal(b []byte) {}

type LEReadWhiteListSizeRP struct {
	Status        uint8
	WhiteListSize uint8
}

// LE Clear White List (0x0010)
type LEClearWhiteList struct{}

func (c LEClearWhiteList) Opcode() int      { return opLEClearWhiteList }
func (c LEClearWhiteList) Len() int         { return 0 }
func (c LEClearWhiteList) Marshal(b []byte) {}

type LEClearWhiteListRP struct{ Status uint8 }

// LE Add Device To White List (0x0011)
type LEAddDeviceToWhiteList struct {
	AddressType uint8
	Address     [6]byte
}

func (c LEAddDeviceToWhiteList) Opcode() int { return opLEAddDeviceToWhiteList }
func (c LEAddDeviceToWhiteList) Len() int    { return 7 }
func (c LEAddDeviceToWhiteList) Marshal(b []byte) {
	b[0] = c.AddressType
	o.PutMAC(b[1:], c.Address)
}

type LEAddDeviceToWhiteListRP struct{ Status uint8 }

// LE Remove Device From White List (0x0012)
type LERemoveDeviceFromWhiteList struct {
	AddressType uint8
	Address     [6]byte
}

func (c LERemoveDeviceFromWhiteList) Opcode() int { return opLERemoveDeviceFromWhiteList }
func (c LERemoveDeviceFromWhiteList) Len() int    { return 7 }
func (c LERemoveDeviceFromWhiteList) Marshal(b []byte) {
	b[0] = c.AddressType
	o.PutMAC(b[1:], c.Address)
}

type LERemoveDeviceFromWhiteListRP struct{ Status uint8 }

// LE Connection Update (0x0013)
type LEConnUpdate struct {
	ConnectionHandle   uint16
	ConnIntervalMin    uint16
	ConnIntervalMax    uint16
	ConnLatency        uint16
	SupervisionTimeout uint16
	MinimumCELength    uint16
	MaximumCELength    uint16
}

func (c LEConnUpdate) Opcode() int { return opLEConnUpdate }
func (c LEConnUpdate) Len() int    { return 14 }
func (c LEConnUpdate) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	o.PutUint16(b[2:], c.ConnIntervalMin)
	o.PutUint16(b[4:], c.ConnIntervalMax)
	o.PutUint16(b[6:], c.ConnLatency)
	o.PutUint16(b[8:], c.SupervisionTimeout)
	o.PutUint16(b[10:], c.MinimumCELength)
	o.PutUint16(b[12:], c.MaximumCELength)
}

type LEConnUpdateRP struct{}

// LE Set Host Channel Classification (0x0014)
type LESetHostChannelClassification struct{ ChannelMap [5]byte }

func (c LESetHostChannelClassification) Opcode() int      { return opLESetHostChannelClassification }
func (c LESetHostChannelClassification) Len() int         { return 5 }
func (c LESetHostChannelClassification) Marshal(b []byte) { copy(b, c.ChannelMap[:]) }

type LESetHostChannelClassificationRP struct{ Status uint8 }

// LE Read Channel Map (0x0015)
type LEReadChannelMap struct{ ConnectionHandle uint16 }

func (c LEReadChannelMap) Opcode() int      { return opLEReadChannelMap }
func (c LEReadChannelMap) Len() int         { return 2 }
func (c LEReadChannelMap) Marshal(b []byte) { o.PutUint16(b, c.ConnectionHandle) }

type LEReadChannelMapRP struct {
	Status           uint8
	ConnectionHandle uint16
	ChannelMap       [5]byte
}

// LE Read Remote Used Features (0x0016)
type LEReadRemoteUsedFeatures struct{ ConnectionHandle uint16 }

func (c LEReadRemoteUsedFeatures) Opcode() int      { return opLEReadRemoteUsedFeatures }
func (c LEReadRemoteUsedFeatures) Len() int         { return 8 }
func (c LEReadRemoteUsedFeatures) Marshal(b []byte) { o.PutUint16(b, c.ConnectionHandle) }

type LEReadRemoteUsedFeaturesRP struct{}

// LE Encrypt (0x0017)
type LEEncrypt struct {
	Key           [16]byte
	PlaintextData [16]byte
}

func (c LEEncrypt) Opcode() int { return opLEEncrypt }
func (c LEEncrypt) Len() int    { return 32 }
func (c LEEncrypt) Marshal(b []byte) {
	copy(b[0:], c.Key[:])
	copy(b[16:], c.PlaintextData[:])
}

type LEEncryptRP struct {
	Stauts        uint8
	EncryptedData [16]byte
}

// LE Rand (0x0018)
type LERand struct{}

func (c LERand) Opcode() int      { return opLERand }
func (c LERand) Len() int         { return 0 }
func (c LERand) Marshal(b []byte) {}

type LERandRP struct {
	Status       uint8
	RandomNumber uint64
}

// LE Start Encryption (0x0019)
type LEStartEncryption struct {
	ConnectionHandle     uint16
	RandomNumber         uint64
	EncryptedDiversifier uint16
	LongTermKey          [16]byte
}

func (c LEStartEncryption) Opcode() int { return opLEStartEncryption }
func (c LEStartEncryption) Len() int    { return 28 }
func (c LEStartEncryption) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	o.PutUint64(b[2:], c.RandomNumber)
	o.PutUint16(b[10:], c.EncryptedDiversifier)
	copy(b[12:], c.LongTermKey[:])
}

type LEStartEncryptionRP struct{}

// LE Long Term Key Reply (0x001A)
type LELTKReply struct {
	ConnectionHandle uint16
	LongTermKey      [16]byte
}

func (c LELTKReply) Opcode() int { return opLELTKReply }
func (c LELTKReply) Len() int    { return 18 }
func (c LELTKReply) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	copy(b[2:], c.LongTermKey[:])
}

type LELTKReplyRP struct {
	Status           uint8
	ConnectionHandle uint16
}

// LE Long Term Key  Negative Reply (0x001B)
type LELTKNegReply struct{ ConnectionHandle uint16 }

func (c LELTKNegReply) Opcode() int      { return opLELTKNegReply }
func (c LELTKNegReply) Len() int         { return 2 }
func (c LELTKNegReply) Marshal(b []byte) { o.PutUint16(b, c.ConnectionHandle) }

type LELTKNegReplyRP struct {
	Status           uint8
	ConnectionHandle uint16
}

// LE Read Supported States (0x001C)
type LEReadSupportedStates struct{}

func (c LEReadSupportedStates) Opcode() int      { return opLEReadSupportedStates }
func (c LEReadSupportedStates) Len() int         { return 0 }
func (c LEReadSupportedStates) Marshal(b []byte) {}

type LEReadSupportedStatesRP struct {
	Status   uint8
	LEStates [8]byte
}

// LE Reciever Test (0x001D)
type LEReceiverTest struct{ RxChannel uint8 }

func (c LEReceiverTest) Opcode() int      { return opLEReceiverTest }
func (c LEReceiverTest) Len() int         { return 1 }
func (c LEReceiverTest) Marshal(b []byte) { b[0] = c.RxChannel }

type LEReceiverTestRP struct{ Status uint8 }

// LE Transmitter Test (0x001E)
type LETransmitterTest struct {
	TxChannel        uint8
	LengthOfTestData uint8
	PacketPayload    uint8
}

func (c LETransmitterTest) Opcode() int { return opLETransmitterTest }
func (c LETransmitterTest) Len() int    { return 3 }
func (c LETransmitterTest) Marshal(b []byte) {
	b[0], b[1], b[2] = c.TxChannel, c.LengthOfTestData, c.PacketPayload
}

type LETransmitterTestRP struct{ Status uint8 }

// LE Test End (0x001F)
type LETestEnd struct{}

func (c LETestEnd) Opcode() int      { return opLETestEnd }
func (c LETestEnd) Len() int         { return 0 }
func (c LETestEnd) Marshal(b []byte) {}

type LETestEndRP struct {
	Status          uint8
	NumberOfPackets uint16
}

// LE Remote Connection Parameters Reply (0x0020)
type LERemoteConnectionParameterReply struct {
	ConnectionHandle uint16
	IntervalMin      uint16
	IntervalMax      uint16
	Latency          uint16
	Timeout          uint16
	MinimumCELength  uint16
	MaximumCELength  uint16
}

func (c LERemoteConnectionParameterReply) Opcode() int { return opLERemoteConnectionParameterReply }
func (c LERemoteConnectionParameterReply) Len() int    { return 14 }
func (c LERemoteConnectionParameterReply) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	o.PutUint16(b[2:], c.IntervalMin)
	o.PutUint16(b[4:], c.IntervalMax)
	o.PutUint16(b[6:], c.Latency)
	o.PutUint16(b[8:], c.Timeout)
	o.PutUint16(b[10:], c.MinimumCELength)
	o.PutUint16(b[12:], c.MaximumCELength)
}

type LERemoteConnectionParameterReplyRP struct {
	Status           uint8
	ConnectionHandle uint16
}

// LE Remote Connection Parameters Negative Reply (0x0021)
type LERemoteConnectionParameterNegReply struct {
	ConnectionHandle uint16
	Reason           uint8
}

func (c LERemoteConnectionParameterNegReply) Opcode() int {
	return opLERemoteConnectionParameterNegReply
}
func (c LERemoteConnectionParameterNegReply) Len() int { return 3 }
func (c LERemoteConnectionParameterNegReply) Marshal(b []byte) {
	o.PutUint16(b[0:], c.ConnectionHandle)
	b[2] = c.Reason
}

type LERemoteConnectionParameterNegReplyRP struct {
	Status           uint8
	ConnectionHandle uint16
}
