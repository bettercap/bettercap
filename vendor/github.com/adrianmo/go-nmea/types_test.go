package nmea

import "testing"

var nearDistance = 0.001

func TestLatLongParse(t *testing.T) {
	var l LatLong
	var err error
	value, expected := "3345.1232 N", LatLong(33.752054)
	if l, err = ParseGPS(value); err != nil {
		t.Errorf("ParseGPS error: %s", err)
	} else if !l.IsNear(expected, nearDistance) {
		t.Errorf("ParseGPS got %f, expected %f", l, expected)
	}

	value, expected = "15145.9877 S", LatLong(-151.76646)
	if l, err = ParseGPS(value); err != nil {
		t.Errorf("ParseGPS error: %s", err)
	} else if !l.IsNear(expected, nearDistance) {
		t.Errorf("ParseGPS got %f, expected %f", l, expected)
	}

	value, expected = "33\u00B0 12' 34.3423\"", LatLong(33.209540)
	if l, err = ParseDMS(value); err != nil {
		t.Errorf("ParseDMS error: %s", err)
	} else if !l.IsNear(expected, nearDistance) {
		t.Errorf("ParseDMS got %f, expected %f", l, expected)
	}

	value, expected = "151.234532", LatLong(151.234532)
	if l, err = ParseDecimal(value); err != nil {
		t.Errorf("ParseDecimal error: %s", err)
	} else if !l.IsNear(expected, nearDistance) {
		t.Errorf("ParseDecimal got %f, expected %f", l, expected)
	}

	value, expected = "-151.234532", LatLong(-151.234532)
	if l, err = ParseDecimal(value); err != nil {
		t.Errorf("ParseDecimal error: %s", err)
	} else if !l.IsNear(expected, nearDistance) {
		t.Errorf("ParseDecimal got %f, expected %f", l, expected)
	}
}

func TestLatLongPrint(t *testing.T) {
	l, _ := ParseDecimal("151.434367")
	exp := "15126.0620"
	if s := l.PrintGPS(); s != exp {
		t.Errorf("PrintGPS() got %s expected %s", s, exp)
	}

	l, _ = ParseGPS("3356.4343 N")
	exp = "3356.4343"
	if s := l.PrintGPS(); s != exp {
		t.Errorf("PrintGPS() got %s expected %s", s, exp)
	}

	exp = "33° 56' 26.058000\""
	if s := l.PrintDMS(); s != exp {
		t.Errorf("PrintDMS() got %s expected %s", s, exp)
	}
}

func TestTimeParse(t *testing.T) {
	timetests := []struct {
		value    string
		expected Time
		ok       bool
	}{
		{"123456", Time{true, 12, 34, 56, 0}, true},
		{"", Time{}, true},
		{"112233.123", Time{true, 11, 22, 33, 123}, true},
		{"010203.04", Time{true, 1, 2, 3, 4}, true},
		{"10203.04", Time{}, false},
		{"x0u2xd", Time{}, false},
		{"xx2233.123", Time{}, false},
		{"11xx33.123", Time{}, false},
		{"1122xx.123", Time{}, false},
		{"112233.xxx", Time{}, false},
	}
	for _, tt := range timetests {
		actual, err := ParseTime(tt.value)
		if !tt.ok {
			if err == nil {
				t.Errorf("ParseTime(%s) expected error", tt.value)
			}
		} else {
			if err != nil {
				t.Errorf("ParseTime(%s) %s", tt.value, err)
			}
			if actual != tt.expected {
				t.Errorf("ParseTime(%s) got %s expected %s", tt.value, actual, tt.expected)
			}
		}
	}
}

func TestTimeString(t *testing.T) {
	d := Time{
		Hour:        1,
		Minute:      2,
		Second:      3,
		Millisecond: 4,
	}
	expected := "01:02:03.0004"
	if s := d.String(); s != expected {
		t.Fatalf("got %s, expected %s", s, expected)
	}
}

func TestDateParse(t *testing.T) {
	datetests := []struct {
		value    string
		expected Date
		ok       bool
	}{
		{"010203", Date{true, 1, 2, 3}, true},
		{"01003", Date{}, false},
		{"", Date{}, true},
		{"xx0203", Date{}, false},
		{"01xx03", Date{}, false},
		{"0102xx", Date{}, false},
	}
	for _, tt := range datetests {
		actual, err := ParseDate(tt.value)
		if !tt.ok {
			if err == nil {
				t.Errorf("ParseDate(%s) expected error", tt.value)
			}
		} else {
			if err != nil {
				t.Errorf("ParseDate(%s) %s", tt.value, err)
			}
			if actual != tt.expected {
				t.Errorf("ParseDate(%s) got %s expected %s", tt.value, actual, tt.expected)
			}
		}
	}
}

func TestDateString(t *testing.T) {
	d := Date{
		DD: 1,
		MM: 2,
		YY: 3,
	}
	expected := "01/02/03"
	if s := d.String(); s != expected {
		t.Fatalf("got %s expected %s", s, expected)
	}
}
