package to

import "time"

// MustBool converts v to bool or panics.
func MustBool(v interface{}) bool {
	b, err := Bool(v)
	if err != nil {
		panic(err)
	}
	return b
}

// MustDuration converts v to Duration or panics.
func MustDuration(v interface{}) time.Duration {
	d, err := Duration(v)
	if err != nil {
		panic(err)
	}
	return d
}

// MustInt convert v to int64 or panics.
func MustInt(v interface{}) int {
	i, err := Int64(v)
	if err != nil {
		panic(err)
	}
	return int(i)
}

// MustInt64 convert v to int64 or panics.
func MustInt64(v interface{}) int64 {
	i, err := Int64(v)
	if err != nil {
		panic(err)
	}
	return i
}

// MustFloat converts v to float64 or panics.
func MustFloat(v interface{}) float64 {
	f, err := Float64(v)
	if err != nil {
		panic(err)
	}
	return f
}

// MustMap converts v to map[string]interface{} or returns val.
func MustMap(v interface{}, val map[string]interface{}) map[string]interface{} {
	m, err := Map(v)
	if err != nil {
		panic(err)
	}
	return m
}

// MustSlice converts v to []interface{} or returns val.
func MustSlice(v interface{}, val []interface{}) []interface{} {
	sli, err := Slice(v)
	if err != nil {
		panic(err)
	}
	return sli
}

// MustString converts v to iterface{}
func MustString(v interface{}) string {
	return String(v)
}

// MustTime converts v to Time or panics.
func MustTime(v interface{}) time.Time {
	t, err := Time(v)
	if err != nil {
		panic(err)
	}
	return t
}
