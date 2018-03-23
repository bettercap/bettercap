package dhcp6opts

// An ArchType is a client system architecture type, as defined in RFC 4578,
// Section 2.1.  Though this RFC indicates these constants are for DHCPv4,
// they are carried over for use in DHCPv6 in RFC 5970, Section 3.3.
type ArchType uint16

// ArchType constants which indicate the client system architecture types
// described in RFC 4578, Section 2.1.
const (
	// RFC 4578
	ArchTypeIntelx86PC      ArchType = 0
	ArchTypeNECPC98         ArchType = 1
	ArchTypeEFIItanium      ArchType = 2
	ArchTypeDECAlpha        ArchType = 3
	ArchtypeArcx86          ArchType = 4
	ArchTypeIntelLeanClient ArchType = 5
	ArchTypeEFIIA32         ArchType = 6
	ArchTypeEFIBC           ArchType = 7
	ArchTypeEFIXscale       ArchType = 8
	ArchTypeEFIx8664        ArchType = 9
)
