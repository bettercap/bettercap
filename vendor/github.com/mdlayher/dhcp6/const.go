package dhcp6

// MessageType represents a DHCP message type, as defined in RFC 3315,
// Section 5.3.  Different DHCP message types are used to perform different
// actions between a client and server.
type MessageType uint8

// MessageType constants which indicate the message types described in
// RFCs 3315, 5007, 5460, 6977, and 7341.
//
// These message types are taken from IANA's DHCPv6 parameters registry:
// http://www.iana.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xhtml.
const (
	// RFC 3315
	MessageTypeSolicit            MessageType = 1
	MessageTypeAdvertise          MessageType = 2
	MessageTypeRequest            MessageType = 3
	MessageTypeConfirm            MessageType = 4
	MessageTypeRenew              MessageType = 5
	MessageTypeRebind             MessageType = 6
	MessageTypeReply              MessageType = 7
	MessageTypeRelease            MessageType = 8
	MessageTypeDecline            MessageType = 9
	MessageTypeReconfigure        MessageType = 10
	MessageTypeInformationRequest MessageType = 11
	MessageTypeRelayForw          MessageType = 12
	MessageTypeRelayRepl          MessageType = 13

	// RFC 5007
	MessageTypeLeasequery      MessageType = 14
	MessageTypeLeasequeryReply MessageType = 15

	// RFC 5460
	MessageTypeLeasequeryDone MessageType = 16
	MessageTypeLeasequeryData MessageType = 17

	// RFC 6977
	MessageTypeReconfigureRequest MessageType = 18
	MessageTypeReconfigureReply   MessageType = 19

	// RFC 7341
	MessageTypeDHCPv4Query    MessageType = 20
	MessageTypeDHCPv4Response MessageType = 21
)

// Status represesents a DHCP status code, as defined in RFC 3315,
// Section 5.4.  Status codes are used to communicate success or failure
// between client and server.
type Status uint16

// Status constants which indicate the status codes described in
// RFCs 3315, 3633, 5007, and 5460.
//
// These status codes are taken from IANA's DHCPv6 parameters registry:
// http://www.iana.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xhtml.
const (
	// RFC 3315
	StatusSuccess      Status = 0
	StatusUnspecFail   Status = 1
	StatusNoAddrsAvail Status = 2
	StatusNoBinding    Status = 3
	StatusNotOnLink    Status = 4
	StatusUseMulticast Status = 5

	// RFC 3633
	StatusNoPrefixAvail Status = 6

	// RFC 5007
	StatusUnknownQueryType Status = 7
	StatusMalformedQuery   Status = 8
	StatusNotConfigured    Status = 9
	StatusNotAllowed       Status = 10

	// RFC 5460
	StatusQueryTerminated Status = 11
)

// OptionCode represents a DHCP option, as defined in RFC 3315,
// Section 22.  Options are used to carry additional information and
// parameters in DHCP messages between client and server.
type OptionCode uint16

// OptionCode constants which indicate the option codes described in
// RFC 3315, RFC 3633, and RFC 5970.
//
// These option codes are taken from IANA's DHCPv6 parameters registry:
// http://www.iana.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xhtml.
const (
	// RFC 3315
	OptionClientID     OptionCode = 1
	OptionServerID     OptionCode = 2
	OptionIANA         OptionCode = 3
	OptionIATA         OptionCode = 4
	OptionIAAddr       OptionCode = 5
	OptionORO          OptionCode = 6
	OptionPreference   OptionCode = 7
	OptionElapsedTime  OptionCode = 8
	OptionRelayMsg     OptionCode = 9
	_                  OptionCode = 10
	OptionAuth         OptionCode = 11
	OptionUnicast      OptionCode = 12
	OptionStatusCode   OptionCode = 13
	OptionRapidCommit  OptionCode = 14
	OptionUserClass    OptionCode = 15
	OptionVendorClass  OptionCode = 16
	OptionVendorOpts   OptionCode = 17
	OptionInterfaceID  OptionCode = 18
	OptionReconfMsg    OptionCode = 19
	OptionReconfAccept OptionCode = 20

	// RFC 3646
	OptionDNSServers OptionCode = 23

	// RFC 3633
	OptionIAPD     OptionCode = 25
	OptionIAPrefix OptionCode = 26

	// RFC 4649
	OptionRemoteIdentifier OptionCode = 37

	// RFC 5970
	OptionBootFileURL    OptionCode = 59
	OptionBootFileParam  OptionCode = 60
	OptionClientArchType OptionCode = 61
	OptionNII            OptionCode = 62

	// BUG(mdlayher): add additional option code types defined by IANA
)
