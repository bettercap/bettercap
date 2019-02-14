package gatt

// This file includes constants from the BLE spec.

var (
	attrGAPUUID  = UUID16(0x1800)
	attrGATTUUID = UUID16(0x1801)

	attrPrimaryServiceUUID   = UUID16(0x2800)
	attrSecondaryServiceUUID = UUID16(0x2801)
	attrIncludeUUID          = UUID16(0x2802)
	attrCharacteristicUUID   = UUID16(0x2803)

	attrClientCharacteristicConfigUUID = UUID16(0x2902)
	attrServerCharacteristicConfigUUID = UUID16(0x2903)

	attrDeviceNameUUID        = UUID16(0x2A00)
	attrAppearanceUUID        = UUID16(0x2A01)
	attrPeripheralPrivacyUUID = UUID16(0x2A02)
	attrReconnectionAddrUUID  = UUID16(0x2A03)
	attrPeferredParamsUUID    = UUID16(0x2A04)
	attrServiceChangedUUID    = UUID16(0x2A05)
)

const (
	gattCCCNotifyFlag   = 0x0001
	gattCCCIndicateFlag = 0x0002
)

const (
	attOpError              = 0x01
	attOpMtuReq             = 0x02
	attOpMtuRsp             = 0x03
	attOpFindInfoReq        = 0x04
	attOpFindInfoRsp        = 0x05
	attOpFindByTypeValueReq = 0x06
	attOpFindByTypeValueRsp = 0x07
	attOpReadByTypeReq      = 0x08
	attOpReadByTypeRsp      = 0x09
	attOpReadReq            = 0x0a
	attOpReadRsp            = 0x0b
	attOpReadBlobReq        = 0x0c
	attOpReadBlobRsp        = 0x0d
	attOpReadMultiReq       = 0x0e
	attOpReadMultiRsp       = 0x0f
	attOpReadByGroupReq     = 0x10
	attOpReadByGroupRsp     = 0x11
	attOpWriteReq           = 0x12
	attOpWriteRsp           = 0x13
	attOpWriteCmd           = 0x52
	attOpPrepWriteReq       = 0x16
	attOpPrepWriteRsp       = 0x17
	attOpExecWriteReq       = 0x18
	attOpExecWriteRsp       = 0x19
	attOpHandleNotify       = 0x1b
	attOpHandleInd          = 0x1d
	attOpHandleCnf          = 0x1e
	attOpSignedWriteCmd     = 0xd2
)

type AttEcode byte

const (
	AttEcodeSuccess           AttEcode = 0x00 // Success
	AttEcodeInvalidHandle     AttEcode = 0x01 // The attribute handle given was not valid on this server.
	AttEcodeReadNotPerm       AttEcode = 0x02 // The attribute cannot be read.
	AttEcodeWriteNotPerm      AttEcode = 0x03 // The attribute cannot be written.
	AttEcodeInvalidPDU        AttEcode = 0x04 // The attribute PDU was invalid.
	AttEcodeAuthentication    AttEcode = 0x05 // The attribute requires authentication before it can be read or written.
	AttEcodeReqNotSupp        AttEcode = 0x06 // Attribute server does not support the request received from the client.
	AttEcodeInvalidOffset     AttEcode = 0x07 // Offset specified was past the end of the attribute.
	AttEcodeAuthorization     AttEcode = 0x08 // The attribute requires authorization before it can be read or written.
	AttEcodePrepQueueFull     AttEcode = 0x09 // Too many prepare writes have been queued.
	AttEcodeAttrNotFound      AttEcode = 0x0a // No attribute found within the given attribute handle range.
	AttEcodeAttrNotLong       AttEcode = 0x0b // The attribute cannot be read or written using the Read Blob Request.
	AttEcodeInsuffEncrKeySize AttEcode = 0x0c // The Encryption Key Size used for encrypting this link is insufficient.
	AttEcodeInvalAttrValueLen AttEcode = 0x0d // The attribute value length is invalid for the operation.
	AttEcodeUnlikely          AttEcode = 0x0e // The attribute request that was requested has encountered an error that was unlikely, and therefore could not be completed as requested.
	AttEcodeInsuffEnc         AttEcode = 0x0f // The attribute requires encryption before it can be read or written.
	AttEcodeUnsuppGrpType     AttEcode = 0x10 // The attribute type is not a supported grouping attribute as defined by a higher layer specification.
	AttEcodeInsuffResources   AttEcode = 0x11 // Insufficient Resources to complete the request.
)

func (a AttEcode) Error() string {
	switch i := int(a); {
	case i < 0x11:
		return AttEcodeName[a]
	case i >= 0x12 && i <= 0x7F: // Reserved for future use
		return "reserved error code"
	case i >= 0x80 && i <= 0x9F: // Application Error, defined by higher level
		return "reserved error code"
	case i >= 0xA0 && i <= 0xDF: // Reserved for future use
		return "reserved error code"
	case i >= 0xE0 && i <= 0xFF: // Common profile and service error codes
		return "profile or service error"
	default: // can't happen, just make compiler happy
		return "unknown error"
	}
}

var AttEcodeName = map[AttEcode]string{
	AttEcodeSuccess:           "success",
	AttEcodeInvalidHandle:     "invalid handle",
	AttEcodeReadNotPerm:       "read not permitted",
	AttEcodeWriteNotPerm:      "write not permitted",
	AttEcodeInvalidPDU:        "invalid PDU",
	AttEcodeAuthentication:    "insufficient authentication",
	AttEcodeReqNotSupp:        "request not supported",
	AttEcodeInvalidOffset:     "invalid offset",
	AttEcodeAuthorization:     "insufficient authorization",
	AttEcodePrepQueueFull:     "prepare queue full",
	AttEcodeAttrNotFound:      "attribute not found",
	AttEcodeAttrNotLong:       "attribute not long",
	AttEcodeInsuffEncrKeySize: "insufficient encryption key size",
	AttEcodeInvalAttrValueLen: "invalid attribute value length",
	AttEcodeUnlikely:          "unlikely error",
	AttEcodeInsuffEnc:         "insufficient encryption",
	AttEcodeUnsuppGrpType:     "unsupported group type",
	AttEcodeInsuffResources:   "insufficient resources",
}

func attErrorRsp(op byte, h uint16, s AttEcode) []byte {
	return attErr{opcode: op, attr: h, status: s}.Marshal()
}

// attRspFor maps from att request
// codes to att response codes.
var attRspFor = map[byte]byte{
	attOpMtuReq:             attOpMtuRsp,
	attOpFindInfoReq:        attOpFindInfoRsp,
	attOpFindByTypeValueReq: attOpFindByTypeValueRsp,
	attOpReadByTypeReq:      attOpReadByTypeRsp,
	attOpReadReq:            attOpReadRsp,
	attOpReadBlobReq:        attOpReadBlobRsp,
	attOpReadMultiReq:       attOpReadMultiRsp,
	attOpReadByGroupReq:     attOpReadByGroupRsp,
	attOpWriteReq:           attOpWriteRsp,
	attOpPrepWriteReq:       attOpPrepWriteRsp,
	attOpExecWriteReq:       attOpExecWriteRsp,
}

type attErr struct {
	opcode uint8
	attr   uint16
	status AttEcode
}

// TODO: Reformulate in a way that lets the caller avoid allocs.
// Accept a []byte? Write directly to an io.Writer?
func (e attErr) Marshal() []byte {
	// little-endian encoding for attr
	return []byte{attOpError, e.opcode, byte(e.attr), byte(e.attr >> 8), byte(e.status)}
}
