package log

import (
	"os"
)

// FatalPolicy represents a callback to be executed on Fatal messages.
type FatalPolicy func()

// os.Exit(1) on Fatal messages.
func ExitOnFatal() {
	os.Exit(1)
}

// os.Exit(0) on Fatal messages.
func ExitCleanOnFatal() {
	os.Exit(0)
}

// Do nothing on Fatal messages.
func NoneOnFatal() {

}
