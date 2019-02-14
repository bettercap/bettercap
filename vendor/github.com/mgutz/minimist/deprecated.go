package minimist

import "github.com/mgutz/to"

// Leftover is an alias for Other and is deprecated. USE Others() instead.
func (am ArgMap) Leftover() []string {
	return am["_"].([]string)
}

//// USE As* functions instead. eg AsBool, AsInt

// ZeroBool tries to convert any of related aliases to bool
func (am ArgMap) ZeroBool(aliases ...string) bool {
	for _, key := range aliases {
		b, err := to.Bool(am[key])
		if err == nil {
			return b
		}
	}
	return false
}

// ZeroString should get value from path or return val.
func (am ArgMap) ZeroString(aliases ...string) string {
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

// ZeroInt should get value from path or return val.
func (am ArgMap) ZeroInt(aliases ...string) int {
	for _, key := range aliases {
		i, err := to.Int64(am[key])
		if err == nil {
			return int(i)
		}
	}
	return 0
}

// ZeroFloat should get value from path or return val.
func (am ArgMap) ZeroFloat(aliases ...string) float64 {
	for _, key := range aliases {
		f, err := to.Float64(am[key])
		if err == nil {
			return f
		}
	}
	return 0
}
