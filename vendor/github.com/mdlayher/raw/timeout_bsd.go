// +build darwin dragonfly freebsd netbsd openbsd

package raw

import (
	"time"
)

const (
	// Maximum read timeout per syscall.
	// It is required because read/recvfrom won't be interrupted on closing of the file descriptor.
	readTimeout = 200 * time.Millisecond
)

// Copyright (c) 2012 The Go Authors. All rights reserved.
// Source code in this file is based on src/net/interface_linux.go,
// from the Go standard library.  The Go license can be found here:
// https://golang.org/LICENSE.

// Taken from:
// https://github.com/golang/go/blob/master/src/net/net.go#L417-L421.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
