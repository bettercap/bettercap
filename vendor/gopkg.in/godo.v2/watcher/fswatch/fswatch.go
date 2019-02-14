package fswatch

import (
	"time"
)

// These values represent the events fswatch knows about. fswatch uses a
// stat(2) call to look up file information; a file will only have a NOPERM
// event if the parent directory has no search permission (i.e. parent
// directory doesn't have executable permissions for the current user).
const (
	NONE     = iota // No event, initial state.
	CREATED         // File was created.
	DELETED         // File was deleted.
	MODIFIED        // File was modified.
	PERM            // Changed permissions
	NOEXIST         // File does not exist.
	NOPERM          // No permissions for the file (see const block comment).
	INVALID         // Any type of error not represented above.
)

// NotificationBufLen is the number of notifications that should be buffered
// in the channel.
var NotificationBufLen = 16

// WatchDelay is the duration between path scans. It defaults to 100ms.
var WatchDelay time.Duration

func init() {
	del, err := time.ParseDuration("100ms")
	if err != nil {
		panic("couldn't set up fswatch: " + err.Error())
	}
	WatchDelay = del
}
