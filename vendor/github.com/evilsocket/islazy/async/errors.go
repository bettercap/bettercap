package async

import (
	"errors"
)

var (
	// ErrTimeout happens when there's a timeout ... doh.
	ErrTimeout = errors.New("timeout")
)
