package nmea

// Latitude / longitude representation.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	//  "unicode/utf8"
)

const (
	// Degrees value
	Degrees = '\u00B0'
	// Minutes value
	Minutes = '\''
	// Seconds value
	Seconds = '"'
	// Point value
	Point = '.'
	// North value
	North = "N"
	// South value
	South = "S"
	// East value
	East = "E"
	// West value
	West = "W"
)

// LatLong type
type LatLong float64

// PrintGPS returns the GPS format for the given LatLong.
func (l LatLong) PrintGPS() string {
	padding := ""
	value := float64(l)
	degrees := math.Floor(math.Abs(value))
	fraction := (math.Abs(value) - degrees) * 60
	if fraction < 10 {
		padding = "0"
	}
	return fmt.Sprintf("%d%s%.4f", int(degrees), padding, fraction)
}

// PrintDMS returns the degrees, minutes, seconds format for the given LatLong.
func (l LatLong) PrintDMS() string {
	val := math.Abs(float64(l))
	degrees := int(math.Floor(val))
	minutes := int(math.Floor(60 * (val - float64(degrees))))
	seconds := 3600 * (val - float64(degrees) - (float64(minutes) / 60))

	return fmt.Sprintf("%d\u00B0 %d' %f\"", degrees, minutes, seconds)
}

//ValidRange validates if the range is between -180 and +180.
func (l LatLong) ValidRange() bool {
	return -180.0 <= l && l <= 180.0
}

// IsNear returns whether the coordinate is near the other coordinate,
// by no further than the given distance away.
func (l LatLong) IsNear(o LatLong, max float64) bool {
	return math.Abs(float64(l-o)) <= max
}

// ParseLatLong parses the supplied string into the LatLong.
//
// Supported formats are:
// - DMS (e.g. 33° 23' 22")
// - Decimal (e.g. 33.23454)
// - GPS (e.g 15113.4322S)
//
func ParseLatLong(s string) (LatLong, error) {
	var l LatLong
	var err error
	invalid := LatLong(0.0) // The invalid value to return.
	if l, err = ParseDMS(s); err == nil {
		return l, nil
	} else if l, err = ParseGPS(s); err == nil {
		return l, nil
	} else if l, err = ParseDecimal(s); err == nil {
		return l, nil
	}
	if !l.ValidRange() {
		return invalid, errors.New("coordinate is not in range -180, 180")
	}
	return invalid, fmt.Errorf("cannot parse [%s], unknown format", s)
}

// ParseGPS parses a GPS/NMEA coordinate.
// e.g 15113.4322S
func ParseGPS(s string) (LatLong, error) {
	parts := strings.Split(s, " ")
	dir := parts[1]
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: %s", err.Error())
	}

	degrees := math.Floor(value / 100)
	minutes := value - (degrees * 100)
	value = degrees + minutes/60

	if dir == North || dir == East {
		return LatLong(value), nil
	} else if dir == South || dir == West {
		return LatLong(0 - value), nil
	} else {
		return 0, fmt.Errorf("invalid direction [%s]", dir)
	}
}

// ParseDecimal parses a decimal format coordinate.
// e.g: 151.196019
func ParseDecimal(s string) (LatLong, error) {
	// Make sure it parses as a float.
	l, err := strconv.ParseFloat(s, 64)
	if err != nil || s[0] != '-' && len(strings.Split(s, ".")[0]) > 3 {
		return LatLong(0.0), errors.New("parse error (not decimal coordinate)")
	}
	return LatLong(l), nil
}

// ParseDMS parses a coordinate in degrees, minutes, seconds.
// - e.g. 33° 23' 22"
func ParseDMS(s string) (LatLong, error) {
	degrees := 0
	minutes := 0
	seconds := 0.0
	// Whether a number has finished parsing (i.e whitespace after it)
	endNumber := false
	// Temporary parse buffer.
	tmpBytes := []byte{}
	var err error

	for i, r := range s {
		if unicode.IsNumber(r) || r == '.' {
			if !endNumber {
				tmpBytes = append(tmpBytes, s[i])
			} else {
				return 0, errors.New("parse error (no delimiter)")
			}
		} else if unicode.IsSpace(r) && len(tmpBytes) > 0 {
			endNumber = true
		} else if r == Degrees {
			if degrees, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (degrees)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if s[i] == Minutes {
			if minutes, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (minutes)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if s[i] == Seconds {
			if seconds, err = strconv.ParseFloat(string(tmpBytes), 64); err != nil {
				return 0, errors.New("parse error (seconds)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if unicode.IsSpace(r) && len(tmpBytes) == 0 {
			continue
		} else {
			return 0, fmt.Errorf("parse error (unknown symbol [%d])", s[i])
		}
	}
	val := LatLong(float64(degrees) + (float64(minutes) / 60.0) + (float64(seconds) / 60.0 / 60.0))
	return val, nil
}

// Time type
type Time struct {
	Valid       bool
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

// String representation of Time
func (t Time) String() string {
	return fmt.Sprintf("%02d:%02d:%02d.%04d", t.Hour, t.Minute, t.Second, t.Millisecond)
}

// ParseTime parses wall clock time.
// e.g. hhmmss.ssss
// An empty time string will result in an invalid time.
func ParseTime(s string) (Time, error) {
	if s == "" {
		return Time{}, nil
	}
	ms := "0000"
	hhmmss := s
	if parts := strings.SplitN(s, ".", 2); len(parts) > 1 {
		hhmmss, ms = parts[0], parts[1]
	}
	if len(hhmmss) != 6 {
		return Time{}, fmt.Errorf("parse time: exptected hhmmss.ss format, got '%s'", s)
	}
	hour, err := strconv.Atoi(hhmmss[0:2])
	if err != nil {
		return Time{}, errors.New(hhmmss)
	}
	minute, err := strconv.Atoi(hhmmss[2:4])
	if err != nil {
		return Time{}, errors.New(hhmmss)
	}
	second, err := strconv.Atoi(hhmmss[4:6])
	if err != nil {
		return Time{}, errors.New(hhmmss)
	}
	millisecond, err := strconv.Atoi(ms)
	if err != nil {
		return Time{}, errors.New(hhmmss)
	}
	return Time{true, hour, minute, second, millisecond}, nil
}

// Date type
type Date struct {
	Valid bool
	DD    int
	MM    int
	YY    int
}

// String representation of date
func (d Date) String() string {
	return fmt.Sprintf("%02d/%02d/%02d", d.DD, d.MM, d.YY)
}

// ParseDate field ddmmyy format
func ParseDate(ddmmyy string) (Date, error) {
	if ddmmyy == "" {
		return Date{}, nil
	}
	if len(ddmmyy) != 6 {
		return Date{}, fmt.Errorf("parse date: exptected ddmmyy format, got '%s'", ddmmyy)
	}
	dd, err := strconv.Atoi(ddmmyy[0:2])
	if err != nil {
		return Date{}, errors.New(ddmmyy)
	}
	mm, err := strconv.Atoi(ddmmyy[2:4])
	if err != nil {
		return Date{}, errors.New(ddmmyy)
	}
	yy, err := strconv.Atoi(ddmmyy[4:6])
	if err != nil {
		return Date{}, errors.New(ddmmyy)
	}
	return Date{true, dd, mm, yy}, nil
}
