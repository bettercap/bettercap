package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPGSAGoodSentence(t *testing.T) {
	goodMsg := "$GPGSA,A,3,22,19,18,27,14,03,,,,,,,3.1,2.0,2.4*36"
	sentence, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")

	// Attributes of the parsed sentence, and their expected values.
	expected := GPGSA{
		Sent: Sent{
			Type:     "GPGSA",
			Fields:   []string{"A", "3", "22", "19", "18", "27", "14", "03", "", "", "", "", "", "", "3.1", "2.0", "2.4"},
			Checksum: "36",
			Raw:      "$GPGSA,A,3,22,19,18,27,14,03,,,,,,,3.1,2.0,2.4*36",
		},
		Mode:    Auto,
		FixType: Fix3D,
		PDOP:    3.1,
		HDOP:    2.0,
		VDOP:    2.4,
		SV:      []string{"22", "19", "18", "27", "14", "03"},
	}

	assert.EqualValues(t, expected, sentence, "Sentence values do not match")
}

func TestGPGSABadMode(t *testing.T) {
	// Make sure bad fix mode is detected.
	badMode := "$GPGSA,F,3,22,19,18,27,14,03,,,,,,,3.1,2.0,2.4*31"
	_, err := Parse(badMode)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPGSA invalid selection mode: F", err.Error(), "Error message does not match")
}

func TestGPGSABadFix(t *testing.T) {
	// Make sure bad fix type is detected.
	badFixType := "$GPGSA,A,6,22,19,18,27,14,03,,,,,,,3.1,2.0,2.4*33"
	_, err := Parse(badFixType)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPGSA invalid fix type: 6", err.Error(), "Error message does not match")
}
