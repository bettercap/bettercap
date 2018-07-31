package nmea

const (
	// PrefixGPRMC prefix of GPRMC sentence type
	PrefixGPRMC = "GPRMC"
	// ValidRMC character
	ValidRMC = "A"
	// InvalidRMC character
	InvalidRMC = "V"
)

// GPRMC is the Recommended Minimum Specific GNSS data.
// http://aprs.gids.nl/nmea/#rmc
type GPRMC struct {
	BaseSentence
	Time      Time    // Time Stamp
	Validity  string  // validity - A-ok, V-invalid
	Latitude  float64 // Latitude
	Longitude float64 // Longitude
	Speed     float64 // Speed in knots
	Course    float64 // True course
	Date      Date    // Date
	Variation float64 // Magnetic variation
}

// newGPRMC constructor
func newGPRMC(s BaseSentence) (GPRMC, error) {
	p := newParser(s, PrefixGPRMC)
	m := GPRMC{
		BaseSentence: s,
		Time:         p.Time(0, "time"),
		Validity:     p.EnumString(1, "validity", ValidRMC, InvalidRMC),
		Latitude:     p.LatLong(2, 3, "latitude"),
		Longitude:    p.LatLong(4, 5, "longitude"),
		Speed:        p.Float64(6, "speed"),
		Course:       p.Float64(7, "course"),
		Date:         p.Date(8, "date"),
		Variation:    p.Float64(9, "variation"),
	}
	if p.EnumString(10, "variation", West, East) == West {
		m.Variation = 0 - m.Variation
	}
	return m, p.Err()
}
