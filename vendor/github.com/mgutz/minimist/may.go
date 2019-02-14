package minimist

import "github.com/mgutz/to"

// MayBool tries to convert any of related aliases to bool
func (am ArgMap) MayBool(defaultValue bool, aliases ...string) bool {
	for _, key := range aliases {
		b, err := to.Bool(am[key])
		if err == nil {
			return b
		}
	}
	return defaultValue
}

// MayString should get value from path or return val.
func (am ArgMap) MayString(defaultValue string, aliases ...string) string {
	if len(aliases) == 0 {
		panic("Alias key(s) required")
	}
	for _, key := range aliases {
		s := to.String(am[key])
		if len(s) > 0 {
			return s
		}
	}
	return defaultValue
}

// MayInt should get value from path or return val.
func (am ArgMap) MayInt(defaultValue int, aliases ...string) int {
	for _, key := range aliases {
		i, err := to.Int64(am[key])
		if err == nil {
			return int(i)
		}
	}
	return defaultValue
}

// MayFloat should get value from path or return val.
func (am ArgMap) MayFloat(defaultValue float64, aliases ...string) float64 {
	for _, key := range aliases {
		f, err := to.Float64(am[key])
		if err == nil {
			return f
		}
	}
	return defaultValue
}
