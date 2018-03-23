package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var gnrmctests = []struct {
	Input  string
	Output GNRMC
}{
	{
		"$GNRMC,220516,A,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*6E",
		GNRMC{
			Time:      Time{true, 22, 05, 16, 0},
			Validity:  "A",
			Speed:     173.8,
			Course:    231.8,
			Date:      Date{true, 13, 06, 94},
			Variation: -4.2,
			Latitude:  MustParseGPS("5133.82 N"),
			Longitude: MustParseGPS("00042.24 W"),
		},
	},
	{
		"$GNRMC,142754.0,A,4302.539570,N,07920.379823,W,0.0,,070617,0.0,E,A*21",
		GNRMC{
			Time:      Time{true, 14, 27, 54, 0},
			Validity:  "A",
			Speed:     0,
			Course:    0,
			Date:      Date{true, 7, 6, 17},
			Variation: 0,
			Latitude:  MustParseGPS("4302.539570 N"),
			Longitude: MustParseGPS("07920.379823 W"),
		},
	},
}

func TestGNRMCGoodSentence(t *testing.T) {

	for _, tt := range gnrmctests {

		s, err := Parse(tt.Input)

		assert.NoError(t, err, "Unexpected error parsing good sentence")
		assert.Equal(t, PrefixGNRMC, s.Prefix(), "Prefix does not match")

		sentence := s.(GNRMC)

		assert.Equal(t, tt.Output.Time, sentence.Time, "Time does not match")
		assert.Equal(t, tt.Output.Validity, sentence.Validity, "Status does not match")
		assert.Equal(t, tt.Output.Speed, sentence.Speed, "Speed does not match")
		assert.Equal(t, tt.Output.Course, sentence.Course, "Course does not match")
		assert.Equal(t, tt.Output.Date, sentence.Date, "Date does not match")
		assert.Equal(t, tt.Output.Variation, sentence.Variation, "Variation does not match")
		assert.Equal(t, tt.Output.Latitude, sentence.Latitude, "Latitude does not match")
		assert.Equal(t, tt.Output.Longitude, sentence.Longitude, "Longitude does not match")
	}

}

func TestGNRMCBadSentence(t *testing.T) {
	badMsg := "$GNRMC,220516,D,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*6B"
	_, err := Parse(badMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: GNRMC invalid validity: D", err.Error(), "Incorrect error message")
}

func TestGNRMCWrongSentence(t *testing.T) {
	wrongMsg := "$GPXTE,A,A,4.07,L,N*6D"
	_, err := Parse(wrongMsg)

	assert.Error(t, err, "Parse error not returned")
	assert.Equal(t, "nmea: sentence type 'GPXTE' not implemented", err.Error(), "Incorrect error message")
}
