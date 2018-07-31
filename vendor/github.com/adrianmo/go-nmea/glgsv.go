package nmea

const (
	// PrefixGLGSV prefix
	PrefixGLGSV = "GLGSV"
)

// GLGSV represents the GPS Satellites in view
// http://aprs.gids.nl/nmea/#glgsv
type GLGSV struct {
	BaseSentence
	TotalMessages   int64       // Total number of messages of this type in this cycle
	MessageNumber   int64       // Message number
	NumberSVsInView int64       // Total number of SVs in view
	Info            []GLGSVInfo // visible satellite info (0-4 of these)
}

// GLGSVInfo represents information about a visible satellite
type GLGSVInfo struct {
	SVPRNNumber int64 // SV PRN number, pseudo-random noise or gold code
	Elevation   int64 // Elevation in degrees, 90 maximum
	Azimuth     int64 // Azimuth, degrees from true north, 000 to 359
	SNR         int64 // SNR, 00-99 dB (null when not tracking)
}

// newGLGSV constructor
func newGLGSV(s BaseSentence) (GLGSV, error) {
	p := newParser(s, PrefixGLGSV)
	m := GLGSV{
		BaseSentence:    s,
		TotalMessages:   p.Int64(0, "total number of messages"),
		MessageNumber:   p.Int64(1, "message number"),
		NumberSVsInView: p.Int64(2, "number of SVs in view"),
	}
	for i := 0; i < 4; i++ {
		if 5*i+4 > len(m.Fields) {
			break
		}
		m.Info = append(m.Info, GLGSVInfo{
			SVPRNNumber: p.Int64(3+i*4, "SV prn number"),
			Elevation:   p.Int64(4+i*4, "elevation"),
			Azimuth:     p.Int64(5+i*4, "azimuth"),
			SNR:         p.Int64(6+i*4, "SNR"),
		})
	}
	return m, p.Err()
}
