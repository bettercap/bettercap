package nmea

import (
	"fmt"
	"strings"
)

const (
	// SentenceStart is the token to indicate the start of a sentence.
	SentenceStart = "$"

	// FieldSep is the token to delimit fields of a sentence.
	FieldSep = ","

	// ChecksumSep is the token to delimit the checksum of a sentence.
	ChecksumSep = "*"
)

// Sentence interface for all NMEA sentence
type Sentence interface {
	fmt.Stringer
	Prefix() string
}

// BaseSentence contains the information about the NMEA sentence
type BaseSentence struct {
	Type     string   // The sentence type (e.g $GPGSA)
	Fields   []string // Array of fields
	Checksum string   // The Checksum
	Raw      string   // The raw NMEA sentence received
}

// Prefix returns the type of the message
func (s BaseSentence) Prefix() string { return s.Type }

// String formats the sentence into a string
func (s BaseSentence) String() string { return s.Raw }

// parseSentence parses a raw message into it's fields
func parseSentence(raw string) (BaseSentence, error) {
	startIndex := strings.Index(raw, SentenceStart)
	if startIndex != 0 {
		return BaseSentence{}, fmt.Errorf("nmea: sentence does not start with a '$'")
	}
	sumSepIndex := strings.Index(raw, ChecksumSep)
	if sumSepIndex == -1 {
		return BaseSentence{}, fmt.Errorf("nmea: sentence does not contain checksum separator")
	}
	var (
		fieldsRaw   = raw[startIndex+1 : sumSepIndex]
		fields      = strings.Split(fieldsRaw, FieldSep)
		checksumRaw = strings.ToUpper(raw[sumSepIndex+1:])
		checksum    = xorChecksum(fieldsRaw)
	)
	// Validate the checksum
	if checksum != checksumRaw {
		return BaseSentence{}, fmt.Errorf(
			"nmea: sentence checksum mismatch [%s != %s]", checksum, checksumRaw)
	}
	return BaseSentence{
		Type:     fields[0],
		Fields:   fields[1:],
		Checksum: checksumRaw,
		Raw:      raw,
	}, nil
}

// xor all the bytes in a string an return it
// as an uppercase hex string
func xorChecksum(s string) string {
	var checksum uint8
	for i := 0; i < len(s); i++ {
		checksum ^= s[i]
	}
	return fmt.Sprintf("%02X", checksum)
}

// Parse parses the given string into the correct sentence type.
func Parse(raw string) (Sentence, error) {
	s, err := parseSentence(raw)
	if err != nil {
		return nil, err
	}
	switch s.Type {
	case PrefixGPRMC:
		return newGPRMC(s)
	case PrefixGNRMC:
		return newGNRMC(s)
	case PrefixGPGGA:
		return newGPGGA(s)
	case PrefixGNGGA:
		return newGNGGA(s)
	case PrefixGPGSA:
		return newGPGSA(s)
	case PrefixGPGLL:
		return newGPGLL(s)
	case PrefixGPVTG:
		return newGPVTG(s)
	case PrefixGPZDA:
		return newGPZDA(s)
	case PrefixPGRME:
		return newPGRME(s)
	case PrefixGPGSV:
		return newGPGSV(s)
	case PrefixGLGSV:
		return newGLGSV(s)
	case PrefixGPHDT:
		return newGPHDT(s)
	default:
		return nil, fmt.Errorf("nmea: sentence type '%s' not implemented", s.Type)
	}
}
