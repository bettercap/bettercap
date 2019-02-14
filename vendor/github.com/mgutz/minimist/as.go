package minimist

import (
	"time"

	"github.com/mgutz/to"
)

// AsBool tries to convert any of related aliases to bool
func (am ArgMap) AsBool(aliases ...string) bool {
	for _, key := range aliases {
		b, err := to.Bool(am[key])
		if err == nil {
			return b
		}
	}
	return false
}

// AsDuration tries to convert any of related aliases to bool
func (am ArgMap) AsDuration(aliases ...string) time.Duration {
	for _, key := range aliases {
		d, err := to.Duration(am[key])
		if err == nil {
			return d
		}
	}
	return 0
}

// AsFloat should get value from path or return val.
func (am ArgMap) AsFloat(aliases ...string) float64 {
	for _, key := range aliases {
		f, err := to.Float64(am[key])
		if err == nil {
			return f
		}
	}
	return 0
}

// AsInt should get value from path or return val.
func (am ArgMap) AsInt(aliases ...string) int {
	for _, key := range aliases {
		i, err := to.Int64(am[key])
		if err == nil {
			return int(i)
		}
	}
	return 0
}

// AsString should get value from path or return val.
func (am ArgMap) AsString(aliases ...string) string {
	if len(aliases) == 0 {
		panic("Alias key(s) required")
	}
	for _, key := range aliases {
		s := to.String(am[key])
		if len(s) > 0 {
			return s
		}
	}
	return ""
}

// AsTime should get value from path or return val.
func (am ArgMap) AsTime(aliases ...string) time.Time {
	for _, key := range aliases {
		t, err := to.Time(am[key])
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
