package to

import "time"

// ZeroBool converts v to bool or returns false.
func ZeroBool(v interface{}) bool {
	b, err := Bool(v)
	if err != nil {
		return false
	}
	return b
}

// ZeroDuration converts v to Duration or returns Duration(0).
func ZeroDuration(v interface{}) time.Duration {
	d, err := Duration(v)
	if err != nil {
		return time.Duration(0)
	}
	return d
}

// ZeroInt converts v to int64 or returns 0.
func ZeroInt(v interface{}) int {
	i, err := Int64(v)
	if err != nil {
		return 0
	}
	return int(i)
}

// ZeroInt64 converts v to int64 or returns 0.
func ZeroInt64(v interface{}) int64 {
	i, err := Int64(v)
	if err != nil {
		return 0
	}
	return i
}

// ZeroFloat converts v to float64 or returns float64(0)
func ZeroFloat(v interface{}) float64 {
	f, err := Float64(v)
	if err != nil {
		return 0
	}
	return f
}

// ZeroMap converts v to map[string]interface{} or returns val.
func ZeroMap(v interface{}) map[string]interface{} {
	m, err := Map(v)
	if err != nil {
		return map[string]interface{}{}
	}
	return m
}

// ZeroSlice converts v to []interface{} or returns val.
func ZeroSlice(v interface{}, val []interface{}) []interface{} {
	sli, err := Slice(v)
	if err != nil {
		return []interface{}{}
	}
	return sli
}

// ZeroString converts v to string or returns ""
func ZeroString(v interface{}) string {
	return String(v)
}

// ZeroTime converts v to Time or returns Time{}
func ZeroTime(v interface{}) time.Time {
	t, err := Time(v)
	if err != nil {
		return time.Time{}
	}
	return t
}
