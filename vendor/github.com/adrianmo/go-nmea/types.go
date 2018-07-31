package nmea

// Latitude / longitude representation.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
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

// ParseLatLong parses the supplied string into the LatLong.
//
// Supported formats are:
// - DMS (e.g. 33° 23' 22")
// - Decimal (e.g. 33.23454)
// - GPS (e.g 15113.4322S)
//
func ParseLatLong(s string) (float64, error) {
	var l float64
	if v, err := ParseDMS(s); err == nil {
		l = v
	} else if v, err := ParseGPS(s); err == nil {
		l = v
	} else if v, err := ParseDecimal(s); err == nil {
		l = v
	} else {
		return 0, fmt.Errorf("cannot parse [%s], unknown format", s)
	}
	if l < -180.0 || 180.0 < l {
		return 0, errors.New("coordinate is not in range -180, 180")
	}
	return l, nil
}

// ParseGPS parses a GPS/NMEA coordinate.
// e.g 15113.4322S
func ParseGPS(s string) (float64, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format: %s", s)
	}
	dir := parts[1]
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: %s", err.Error())
	}

	degrees := math.Floor(value / 100)
	minutes := value - (degrees * 100)
	value = degrees + minutes/60

	if dir == North || dir == East {
		return value, nil
	} else if dir == South || dir == West {
		return 0 - value, nil
	} else {
		return 0, fmt.Errorf("invalid direction [%s]", dir)
	}
}

// FormatGPS formats a GPS/NMEA coordinate
func FormatGPS(l float64) string {
	padding := ""
	degrees := math.Floor(math.Abs(l))
	fraction := (math.Abs(l) - degrees) * 60
	if fraction < 10 {
		padding = "0"
	}
	return fmt.Sprintf("%d%s%.4f", int(degrees), padding, fraction)
}

// ParseDecimal parses a decimal format coordinate.
// e.g: 151.196019
func ParseDecimal(s string) (float64, error) {
	// Make sure it parses as a float.
	l, err := strconv.ParseFloat(s, 64)
	if err != nil || s[0] != '-' && len(strings.Split(s, ".")[0]) > 3 {
		return 0.0, errors.New("parse error (not decimal coordinate)")
	}
	return l, nil
}

// ParseDMS parses a coordinate in degrees, minutes, seconds.
// - e.g. 33° 23' 22"
func ParseDMS(s string) (float64, error) {
	degrees := 0
	minutes := 0
	seconds := 0.0
	// Whether a number has finished parsing (i.e whitespace after it)
	endNumber := false
	// Temporary parse buffer.
	tmpBytes := []byte{}
	var err error

	for i, r := range s {
		switch {
		case unicode.IsNumber(r) || r == '.':
			if !endNumber {
				tmpBytes = append(tmpBytes, s[i])
			} else {
				return 0, errors.New("parse error (no delimiter)")
			}
		case unicode.IsSpace(r) && len(tmpBytes) > 0:
			endNumber = true
		case r == Degrees:
			if degrees, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (degrees)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		case s[i] == Minutes:
			if minutes, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (minutes)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		case s[i] == Seconds:
			if seconds, err = strconv.ParseFloat(string(tmpBytes), 64); err != nil {
				return 0, errors.New("parse error (seconds)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		case unicode.IsSpace(r) && len(tmpBytes) == 0:
			continue
		default:
			return 0, fmt.Errorf("parse error (unknown symbol [%d])", s[i])
		}
	}
	if len(tmpBytes) > 0 {
		return 0, fmt.Errorf("parse error (trailing data [%s])", string(tmpBytes))
	}
	val := float64(degrees) + (float64(minutes) / 60.0) + (float64(seconds) / 60.0 / 60.0)
	return val, nil
}

// FormatDMS returns the degrees, minutes, seconds format for the given LatLong.
func FormatDMS(l float64) string {
	val := math.Abs(l)
	degrees := int(math.Floor(val))
	minutes := int(math.Floor(60 * (val - float64(degrees))))
	seconds := 3600 * (val - float64(degrees) - (float64(minutes) / 60))
	return fmt.Sprintf("%d\u00B0 %d' %f\"", degrees, minutes, seconds)
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
