package minimist

import (
	"fmt"

	"github.com/mgutz/to"
)

// MustBool tries to convert any of related aliases to bool. If no keys
// it panics.
func (am ArgMap) MustBool(aliases ...string) bool {
	for _, key := range aliases {
		if am[key] == nil {
			continue
		}

		b, err := to.Bool(am[key])
		if err != nil {
			panic(err)
		}
		return b
	}
	panic(fmt.Sprintf("None of these bool flags were found: %v", aliases))
}

// MustInt should get value from path or return val.
func (am ArgMap) MustInt(aliases ...string) int {
	for _, key := range aliases {
		if am[key] == nil {
			continue
		}

		i64, err := to.Int64(am[key])
		if err != nil {
			continue
		}
		return int(i64)
	}
	panic(fmt.Sprintf("None of these int flags were found or convertable to int: %v", aliases))
}

// MustFloat should get value from path or return val.
func (am ArgMap) MustFloat(aliases ...string) float64 {
	for _, key := range aliases {
		f, err := to.Float64(am[key])
		if err == nil {
			return f
		}
	}
	panic(fmt.Sprintf("None of these flags were found or convertable to float64: %v", aliases))
}

// MustString should get value from path or return val.
func (am ArgMap) MustString(aliases ...string) string {
	for _, key := range aliases {
		if am[key] == nil {
			continue
		}

		s := to.String(am[key])
		if s == "" {
			continue
		}
		return s
	}
	panic(fmt.Sprintf("None of these string flags were found: %v", aliases))
}
