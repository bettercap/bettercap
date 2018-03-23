package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPGRMEGoodSentence(t *testing.T) {
	goodMsg := "$PGRME,3.3,M,4.9,M,6.0,M*25"
	s, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")
	assert.Equal(t, PrefixPGRME, s.Prefix(), "Prefix does not match")

	sentence := s.(PGRME)

	assert.Equal(t, 3.3, sentence.Horizontal, "Horizontal error does not match")
	assert.Equal(t, 4.9, sentence.Vertical, "Vertical error does not match")
	assert.Equal(t, 6.0, sentence.Spherical, "Spherical error does not match")

}

func TestPGRMEInvalidHorizontalError(t *testing.T) {
	badMsg := "$PGRME,A,M,4.9,M,6.0,M*4A"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid horizontal error: A", err.Error(), "Incorrect error message")
}

func TestPGRMEInvalidHorizontalErrorUnit(t *testing.T) {
	badMsg := "$PGRME,3.3,A,4.9,M,6.0,M*29"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid horizontal error unit: A", err.Error(), "Incorrect error message")
}

func TestPGRMEInvalidVerticalError(t *testing.T) {
	badMsg := "$PGRME,3.3,M,A,M,6.0,M*47"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid vertical error: A", err.Error(), "Incorrect error message")
}

func TestPGRMEInvalidVerticalErrorUnit(t *testing.T) {
	badMsg := "$PGRME,3.3,M,4.9,A,6.0,M*29"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid vertical error unit: A", err.Error(), "Incorrect error message")
}

func TestPGRMEInvalidSphericalError(t *testing.T) {
	badMsg := "$PGRME,3.3,M,4.9,M,A,M*4C"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid spherical error: A", err.Error(), "Incorrect error message")
}

func TestPGRMEInvalidSphericalErrorUnit(t *testing.T) {
	badMsg := "$PGRME,3.3,M,4.9,M,6.0,A*29"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: PGRME invalid spherical error unit: A", err.Error(), "Incorrect error message")
}
