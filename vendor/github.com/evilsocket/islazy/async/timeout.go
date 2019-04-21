package async

import (
	"time"
)

// TimedCallback represents a generic function with a return value.
type TimedCallback func() interface{}

// WithTimeout will execute the callback and return its value or a
// ErrTimeout if its execution will exceed the provided duration.
func WithTimeout(tm time.Duration, cb TimedCallback) (interface{}, error) {
	timeout := time.After(tm)
	done := make(chan interface{})
	go func() {
		done <- cb()
	}()

	select {
	case <-timeout:
		return nil, ErrTimeout
	case res := <-done:
		if res != nil {
			if e, ok := res.(error); ok {
				return nil, e
			}
		}
		return res, nil
	}
}
