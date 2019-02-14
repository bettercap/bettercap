package to

import "time"

// AsBool converts v to bool or returns false.
func AsBool(v interface{}) bool {
	b, err := Bool(v)
	if err != nil {
		return false
	}
	return b
}

// AsDuration converts v to Duration or returns Duration(0).
func AsDuration(v interface{}) time.Duration {
	d, err := Duration(v)
	if err != nil {
		return time.Duration(0)
	}
	return d
}

// AsInt converts v to int or returns 0.
func AsInt(v interface{}) int {
	i, err := Int64(v)
	if err != nil {
		return 0
	}
	return int(i)
}

// AsInt64 converts v to int64 or returns 0.
func AsInt64(v interface{}) int64 {
	i, err := Int64(v)
	if err != nil {
		return 0
	}
	return i
}

// AsFloat converts v to float64 or returns float64(0)
func AsFloat(v interface{}) float64 {
	f, err := Float64(v)
	if err != nil {
		return 0
	}
	return f
}

// AsMap converts v to map[string]interface{} or returns val.
func AsMap(v interface{}) map[string]interface{} {
	m, err := Map(v)
	if err != nil {
		return map[string]interface{}{}
	}
	return m
}

// AsSlice converts v to []interface{} or returns val.
func AsSlice(v interface{}, val []interface{}) []interface{} {
	sli, err := Slice(v)
	if err != nil {
		return []interface{}{}
	}
	return sli
}

// AsString converts v to string or returns ""
func AsString(v interface{}) string {
	return String(v)
}

// AsTime converts v to Time or returns Time{}
func AsTime(v interface{}) time.Time {
	t, err := Time(v)
	if err != nil {
		return time.Time{}
	}
	return t
}
