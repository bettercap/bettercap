package nmea

const (
	// PrefixGPVTG prefix
	PrefixGPVTG = "GPVTG"
)

// GPVTG represents track & speed data.
// http://aprs.gids.nl/nmea/#vtg
type GPVTG struct {
	BaseSentence
	TrueTrack        float64
	MagneticTrack    float64
	GroundSpeedKnots float64
	GroundSpeedKPH   float64
}

// newGPVTG parses the GPVTG sentence into this struct.
// e.g: $GPVTG,360.0,T,348.7,M,000.0,N,000.0,K*43
func newGPVTG(s BaseSentence) (GPVTG, error) {
	p := newParser(s, PrefixGPVTG)
	return GPVTG{
		BaseSentence:     s,
		TrueTrack:        p.Float64(0, "true track"),
		MagneticTrack:    p.Float64(2, "magnetic track"),
		GroundSpeedKnots: p.Float64(4, "ground speed (knots)"),
		GroundSpeedKPH:   p.Float64(6, "ground speed (km/h)"),
	}, p.Err()
}
