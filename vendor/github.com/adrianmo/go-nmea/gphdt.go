package nmea

const (
	// PrefixGPHDT prefix of GPHDT sentence type
	PrefixGPHDT = "GPHDT"
)

// GPHDT is the Actual vessel heading in degrees True.
// http://aprs.gids.nl/nmea/#hdt
type GPHDT struct {
	BaseSentence
	Heading float64 // Heading in degrees
	True    bool    // Heading is relative to true north
}

// newGPHDT constructor
func newGPHDT(s BaseSentence) (GPHDT, error) {
	p := newParser(s, PrefixGPHDT)
	m := GPHDT{
		BaseSentence: s,
		Heading:      p.Float64(0, "heading"),
		True:         p.EnumString(1, "true", "T") == "T",
	}
	return m, p.Err()
}
