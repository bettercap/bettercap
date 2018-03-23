package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPVTGGoodSentence(t *testing.T) {
	goodMsg := "$GPVTG,45.5,T,67.5,M,30.45,N,56.40,K*4B"
	s, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")
	assert.Equal(t, PrefixGPVTG, s.Prefix(), "Prefix does not match")

	sentence := s.(GPVTG)

	assert.Equal(t, 45.5, sentence.TrueTrack, "True track does not match")
	assert.Equal(t, 67.5, sentence.MagneticTrack, "Magnetic track does not match")
	assert.Equal(t, 30.45, sentence.GroundSpeedKnots, "Ground speed (knots) does not match")
	assert.Equal(t, 56.40, sentence.GroundSpeedKPH, "Ground speed (km/h) does not match")
}

func TestGPVTGBadSentence(t *testing.T) {
	badMsg := "$GPVTG,T,45.5,67.5,M,30.45,N,56.40,K*4B"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPVTG invalid true track: T", err.Error(), "Incorrect error message")
}

func TestGPVTGWrongSentence(t *testing.T) {
	wrongMsg := "$GPXTE,A,A,4.07,L,N*6D"
	_, err := Parse(wrongMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: sentence type 'GPXTE' not implemented", err.Error(), "Incorrect error message")
}
