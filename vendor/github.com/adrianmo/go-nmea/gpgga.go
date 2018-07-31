package nmea

const (
	// PrefixGPGGA prefix
	PrefixGPGGA = "GPGGA"
	// Invalid fix quality.
	Invalid = "0"
	// GPS fix quality
	GPS = "1"
	// DGPS fix quality
	DGPS = "2"
	// PPS fix
	PPS = "3"
	// RTK real time kinematic fix
	RTK = "4"
	// FRTK float RTK fix
	FRTK = "5"
)

// GPGGA represents fix data.
// http://aprs.gids.nl/nmea/#gga
type GPGGA struct {
	BaseSentence
	Time          Time    // Time of fix.
	Latitude      float64 // Latitude.
	Longitude     float64 // Longitude.
	FixQuality    string  // Quality of fix.
	NumSatellites int64   // Number of satellites in use.
	HDOP          float64 // Horizontal dilution of precision.
	Altitude      float64 // Altitude.
	Separation    float64 // Geoidal separation
	DGPSAge       string  // Age of differential GPD data.
	DGPSId        string  // DGPS reference station ID.
}

// newGPGGA parses the GPGGA sentence into this struct.
// e.g: $GPGGA,034225.077,3356.4650,S,15124.5567,E,1,03,9.7,-25.0,M,21.0,M,,0000*58
func newGPGGA(s BaseSentence) (GPGGA, error) {
	p := newParser(s, PrefixGPGGA)
	return GPGGA{
		BaseSentence:  s,
		Time:          p.Time(0, "time"),
		Latitude:      p.LatLong(1, 2, "latitude"),
		Longitude:     p.LatLong(3, 4, "longitude"),
		FixQuality:    p.EnumString(5, "fix quality", Invalid, GPS, DGPS, PPS, RTK, FRTK),
		NumSatellites: p.Int64(6, "number of satellites"),
		HDOP:          p.Float64(7, "hdap"),
		Altitude:      p.Float64(8, "altitude"),
		Separation:    p.Float64(10, "separation"),
		DGPSAge:       p.String(12, "dgps age"),
		DGPSId:        p.String(13, "dgps id"),
	}, p.Err()
}
