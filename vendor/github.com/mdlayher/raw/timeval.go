// +build !darwin,!arm,!windows,!mipsle,!mips,!386

package raw

import (
	"time"

	"golang.org/x/sys/unix"
)

// newTimeval transforms a duration into a unix.Timeval struct.
// An error is returned in case of zero time value.
func newTimeval(timeout time.Duration) (*unix.Timeval, error) {
	if timeout < time.Microsecond {
		return nil, &timeoutError{}
	}
	return &unix.Timeval{
		Sec:  int64(timeout / time.Second),
		Usec: int64(timeout % time.Second / time.Microsecond),
	}, nil
}
