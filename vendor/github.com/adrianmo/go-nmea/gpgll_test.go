package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPGLLGoodSentence(t *testing.T) {
	goodMsg := "$GPGLL,3926.7952,N,12000.5947,W,022732,A,A*58"
	s, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")
	assert.Equal(t, PrefixGPGLL, s.Prefix(), "Prefix does not match")

	sentence := s.(GPGLL)

	assert.Equal(t, "3926.7952", sentence.Latitude.PrintGPS(), "Latitude does not match")
	assert.Equal(t, "12000.5947", sentence.Longitude.PrintGPS(), "Longitude does not match")
	assert.Equal(t, Time{true, 2, 27, 32, 0}, sentence.Time, "Time does not match")
	assert.Equal(t, "A", sentence.Validity, "Status does not match")
}

func TestGPGLLBadSentence(t *testing.T) {
	badMsg := "$GPGLL,3926.7952,N,12000.5947,W,022732,D,A*5D"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPGLL invalid validity: D", err.Error(), "Incorrect error message")
}

func TestGPGLLWrongSentence(t *testing.T) {
	wrongMsg := "$GPXTE,A,A,4.07,L,N*6D"
	_, err := Parse(wrongMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: sentence type 'GPXTE' not implemented", err.Error(), "Incorrect error message")
}
