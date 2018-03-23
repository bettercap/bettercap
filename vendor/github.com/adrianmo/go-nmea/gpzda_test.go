package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPZDAGoodSentence(t *testing.T) {
	goodMsg := "$GPZDA,172809.456,12,07,1996,00,00*57"
	s, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")
	assert.Equal(t, PrefixGPZDA, s.Prefix(), "Prefix does not match")

	sentence := s.(GPZDA)

	assert.Equal(t, Time{true, 17, 28, 9, 456}, sentence.Time, "Time does not match")
	assert.Equal(t, int64(12), sentence.Day, "Day does not match")
	assert.Equal(t, int64(7), sentence.Month, "Month does not match")
	assert.Equal(t, int64(1996), sentence.Year, "Yeah does not match")
	assert.Equal(t, int64(0), sentence.OffsetHours, "Offset (hours) does not match")
	assert.Equal(t, int64(0), sentence.OffsetMinutes, "Offset (minutes) does not match")
}

func TestGPZDABadSentence(t *testing.T) {
	badMsg := "$GPZDA,220516,D,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*76"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPZDA invalid day: D", err.Error(), "Incorrect error message")
}

func TestGPZDAWrongSentence(t *testing.T) {
	wrongMsg := "$GPXTE,A,A,4.07,L,N*6D"
	_, err := Parse(wrongMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: sentence type 'GPXTE' not implemented", err.Error(), "Incorrect error message")
}
