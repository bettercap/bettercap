package nmea

const (
	// PrefixPGRME prefix for PGRME sentence type
	PrefixPGRME = "PGRME"
	// ErrorUnit must be meters (M)
	ErrorUnit = "M"
)

// PGRME is Estimated Position Error (Garmin proprietary sentence)
// http://aprs.gids.nl/nmea/#rme
type PGRME struct {
	BaseSentence
	Horizontal float64 // Estimated horizontal position error (HPE) in metres
	Vertical   float64 // Estimated vertical position error (VPE) in metres
	Spherical  float64 // Overall spherical equivalent position error in meters
}

// newPGRME constructor
func newPGRME(s BaseSentence) (PGRME, error) {
	p := newParser(s, PrefixPGRME)

	horizontal := p.Float64(0, "horizontal error")
	_ = p.EnumString(1, "horizontal error unit", ErrorUnit)

	vertial := p.Float64(2, "vertical error")
	_ = p.EnumString(3, "vertical error unit", ErrorUnit)

	spherical := p.Float64(4, "spherical error")
	_ = p.EnumString(5, "spherical error unit", ErrorUnit)

	return PGRME{
		BaseSentence: s,
		Horizontal:   horizontal,
		Vertical:     vertial,
		Spherical:    spherical,
	}, p.Err()
}
