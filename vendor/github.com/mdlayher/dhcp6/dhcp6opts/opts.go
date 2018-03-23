package dhcp6opts

import (
	"errors"
)

//go:generate stringer -output=string.go -type=ArchType,DUIDType

var (
	// ErrHardwareTypeNotImplemented is returned when HardwareType is not
	// implemented on the current platform.
	ErrHardwareTypeNotImplemented = errors.New("hardware type detection not implemented on this platform")

	// ErrInvalidDUIDLLTTime is returned when a time before midnight (UTC),
	// January 1, 2000 is used in NewDUIDLLT.
	ErrInvalidDUIDLLTTime = errors.New("DUID-LLT time must be after midnight (UTC), January 1, 2000")

	// ErrInvalidIP is returned when an input net.IP value is not recognized as a
	// valid IPv6 address.
	ErrInvalidIP = errors.New("IP must be an IPv6 address")

	// ErrInvalidLifetimes is returned when an input preferred lifetime is shorter
	// than a valid lifetime parameter.
	ErrInvalidLifetimes = errors.New("preferred lifetime must be less than valid lifetime")

	// ErrParseHardwareType is returned when a valid hardware type could
	// not be found for a given interface.
	ErrParseHardwareType = errors.New("could not parse hardware type for interface")
)
