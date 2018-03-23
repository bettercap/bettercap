package nmea

const (
	// PrefixGNRMC prefix of GNRMC sentence type
	PrefixGNRMC = "GNRMC"
)

// GNRMC is the Recommended Minimum Specific GNSS data.
// http://aprs.gids.nl/nmea/#rmc
type GNRMC struct {
	Sent
	Time      Time    // Time Stamp
	Validity  string  // validity - A-ok, V-invalid
	Latitude  LatLong // Latitude
	Longitude LatLong // Longitude
	Speed     float64 // Speed in knots
	Course    float64 // True course
	Date      Date    // Date
	Variation float64 // Magnetic variation
}

// NewGNRMC constructor
func NewGNRMC(s Sent) (GNRMC, error) {
	p := newParser(s, PrefixGNRMC)
	m := GNRMC{
		Sent:      s,
		Time:      p.Time(0, "time"),
		Validity:  p.EnumString(1, "validity", ValidRMC, InvalidRMC),
		Latitude:  p.LatLong(2, 3, "latitude"),
		Longitude: p.LatLong(4, 5, "longitude"),
		Speed:     p.Float64(6, "speed"),
		Course:    p.Float64(7, "course"),
		Date:      p.Date(8, "date"),
		Variation: p.Float64(9, "variation"),
	}
	if p.EnumString(10, "direction", West, East) == West {
		m.Variation = 0 - m.Variation
	}
	return m, p.Err()
}
