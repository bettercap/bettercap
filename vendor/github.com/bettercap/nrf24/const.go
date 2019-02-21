package nrf24

import (
	"github.com/google/gousb"
)

const (
	VendorID  gousb.ID = 0x1915
	ProductID gousb.ID = 0x0102
)

const (
	PacketSize = 64
	MinChannel = 1
	TopChannel = 83
	MaxChannel = 125
)

type Command byte

// USB commands
const (
	CmdTransmitPayload         Command = 0x04
	CmdEnterSnifferMode        Command = 0x05
	CmdEnterPromiscMode        Command = 0x06
	CmdEnterToneTestMode       Command = 0x07
	CmdTransmitAckPayload      Command = 0x08
	CmdSetChannel              Command = 0x09
	CmdGetChannel              Command = 0x0A
	CmdEnableLNAPA             Command = 0x0B
	CmdTransmitPayloadGeneric  Command = 0x0C
	CmdEnterPromiscModeGeneric Command = 0x0D
	CmdReceivePayload          Command = 0x12
)

type RfRate byte

const (
	RfRate250K RfRate = 0
	RfRate1M   RfRate = 1
	RfRate2M   RfRate = 2
)
