package watcher

import "fmt"

//"log"

const (
	// NONE means no event, initial state.
	NONE = iota
	// CREATED means file was created.
	CREATED
	// DELETED means file was deleted.
	DELETED
	// MODIFIED means file was modified.
	MODIFIED
	// PERM means changed permissions
	PERM
	// NOEXIST means file does not exist.
	NOEXIST
	// NOPERM means no permissions for the file (see const block comment).
	NOPERM
	// INVALID means any type of error not represented above.
	INVALID
)

// FileEvent is a wrapper around github.com/howeyc/fsnotify.FileEvent
type FileEvent struct {
	Event    int
	Path     string
	UnixNano int64
}

// newFileEvent creates a new file event.
func newFileEvent(op int, path string, unixNano int64) *FileEvent {
	//log.Printf("to channel %+v\n", originEvent)
	return &FileEvent{Event: op, Path: path, UnixNano: unixNano}
}

// String returns an eye friendly version of this event.
func (fe *FileEvent) String() string {
	var status string
	switch fe.Event {
	case CREATED:
		status = "was created"
	case DELETED:
		status = "was deleted"
	case MODIFIED:
		status = "was modified"
	case PERM:
		status = "permissions changed"
	case NOEXIST:
		status = "does not exist"
	case NOPERM:
		status = "is not accessible (permission)"
	case INVALID:
		status = "is invalid"
	}
	return fmt.Sprintf("%s %s", fe.Path, status)
}
