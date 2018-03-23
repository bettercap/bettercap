package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPGGAGoodSentence(t *testing.T) {
	goodMsg := "$GPGGA,034225.077,3356.4650,S,15124.5567,E,1,03,9.7,-25.0,M,21.0,M,,0000*51"
	sentence, err := Parse(goodMsg)

	assert.NoError(t, err, "Unexpected error parsing good sentence")

	lat, _ := ParseLatLong("3356.4650 S")
	lon, _ := ParseLatLong("15124.5567 E")
	// Attributes of the parsed sentence, and their expected values.
	expected := GPGGA{
		Sent: Sent{
			Type:     "GPGGA",
			Fields:   []string{"034225.077", "3356.4650", "S", "15124.5567", "E", "1", "03", "9.7", "-25.0", "M", "21.0", "M", "", "0000"},
			Checksum: "51",
			Raw:      "$GPGGA,034225.077,3356.4650,S,15124.5567,E,1,03,9.7,-25.0,M,21.0,M,,0000*51",
		},
		Time:          Time{true, 3, 42, 25, 77},
		Latitude:      lat,
		Longitude:     lon,
		FixQuality:    GPS,
		NumSatellites: 03,
		HDOP:          9.7,
		Altitude:      -25.0,
		Separation:    21.0,
		DGPSAge:       "",
		DGPSId:        "0000",
	}

	assert.EqualValues(t, expected, sentence, "Sentence values do not match")
}

func TestGPGGABadType(t *testing.T) {
	badType := "$GPRMC,220516,A,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*70"
	s, err := Parse(badType)

	assert.NoError(t, err, "Unexpected error parsing sentence")
	assert.NotEqual(t, "GPGGA", s.Prefix(), "Unexpected sentence type")
}

func TestGPGGABadLatitude(t *testing.T) {
	badLat := "$GPGGA,034225.077,A,S,15124.5567,E,1,03,9.7,-25.0,M,21.0,M,,0000*3A"
	_, err := Parse(badLat)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPGGA invalid latitude: cannot parse [A S], unknown format", err.Error(), "Error message does not match")
}

func TestGPGGABadLongitude(t *testing.T) {
	badLon := "$GPGGA,034225.077,3356.4650,S,A,E,1,03,9.7,-25.0,M,21.0,M,,0000*0C"
	_, err := Parse(badLon)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GPGGA invalid longitude: cannot parse [A E], unknown format", err.Error(), "Error message does not match")
}

func TestGPGGABadFixQuality(t *testing.T) {
	// Make sure bad fix mode is detected.
	badMode := "$GPGGA,034225.077,3356.4650,S,15124.5567,E,5,03,9.7,-25.0,M,21.0,M,,0000*55"
	_, err := Parse(badMode)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, err.Error(), "nmea: GPGGA invalid fix quality: 5", "Error message not as expected")
}
