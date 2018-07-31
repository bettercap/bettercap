package nmea

const (
	// PrefixGPGLL prefix for GPGLL sentence type
	PrefixGPGLL = "GPGLL"
	// ValidGLL character
	ValidGLL = "A"
	// InvalidGLL character
	InvalidGLL = "V"
)

// GPGLL is Geographic Position, Latitude / Longitude and time.
// http://aprs.gids.nl/nmea/#gll
type GPGLL struct {
	BaseSentence
	Latitude  float64 // Latitude
	Longitude float64 // Longitude
	Time      Time    // Time Stamp
	Validity  string  // validity - A-valid
}

// newGPGLL constructor
func newGPGLL(s BaseSentence) (GPGLL, error) {
	p := newParser(s, PrefixGPGLL)
	return GPGLL{
		BaseSentence: s,
		Latitude:     p.LatLong(0, 1, "latitude"),
		Longitude:    p.LatLong(2, 3, "longitude"),
		Time:         p.Time(4, "time"),
		Validity:     p.EnumString(5, "validity", ValidGLL, InvalidGLL),
	}, p.Err()
}
