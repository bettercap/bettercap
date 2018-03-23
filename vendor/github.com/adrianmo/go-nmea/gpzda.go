package nmea

const (
	// PrefixGPZDA prefix
	PrefixGPZDA = "GPZDA"
)

// GPZDA represents date & time data.
// http://aprs.gids.nl/nmea/#zda
type GPZDA struct {
	Sent
	Time          Time
	Day           int64
	Month         int64
	Year          int64
	OffsetHours   int64 // Local time zone offset from GMT, hours
	OffsetMinutes int64 // Local time zone offset from GMT, minutes
}

// NewGPZDA constructor
func NewGPZDA(s Sent) (GPZDA, error) {
	p := newParser(s, PrefixGPZDA)
	return GPZDA{
		Sent:          s,
		Time:          p.Time(0, "time"),
		Day:           p.Int64(1, "day"),
		Month:         p.Int64(2, "month"),
		Year:          p.Int64(3, "year"),
		OffsetHours:   p.Int64(4, "offset (hours)"),
		OffsetMinutes: p.Int64(5, "offset (minutes)"),
	}, p.Err()
}
