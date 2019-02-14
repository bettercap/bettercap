package to

import "time"

// OrBool converts v to bool or returns val.
func OrBool(v interface{}, val bool) bool {
	b, err := Bool(v)
	if err != nil {
		return val
	}
	return b
}

// OrDuration converts v to Duration or returns val.
func OrDuration(v interface{}, val time.Duration) time.Duration {
	d, err := Duration(v)
	if err != nil {
		return val
	}
	return d
}

// OrInt converts v to int64 or returns val.
func OrInt(v interface{}, val int) int {
	i, err := Int64(v)
	if err != nil {
		return val
	}
	return int(i)
}

// OrInt64 converts v to int64 or returns val.
func OrInt64(v interface{}, val int64) int64 {
	i, err := Int64(v)
	if err != nil {
		return val
	}
	return i
}

// OrFloat converts v to float64 or returns val.
func OrFloat(v interface{}, val float64) float64 {
	f, err := Float64(v)
	if err != nil {
		return val
	}
	return f
}

// OrMap converts v to map[string]interface{} or returns val.
func OrMap(v interface{}, val map[string]interface{}) map[string]interface{} {
	m, err := Map(v)
	if err != nil {
		return val
	}
	return m
}

// OrSlice converts v to []interface{} or returns val.
func OrSlice(v interface{}, val []interface{}) []interface{} {
	sli, err := Slice(v)
	if err != nil {
		return val
	}
	return sli
}

// OrString converts v to string or returns ""
func OrString(v interface{}) string {
	return String(v)
}

// OrTime converts v to Time or returns Time{}
func OrTime(v interface{}, val time.Time) time.Time {
	t, err := Time(v)
	if err != nil {
		return val
	}
	return t
}
