package log

import (
	"github.com/evilsocket/islazy/tui"
)

// Verbosity represents the verbosity level of the logger.
type Verbosity int

const (
	// Debug messages.
	DEBUG Verbosity = iota
	// Informative messages.
	INFO
	// Informative messages that are important.
	IMPORTANT
	// Warning messages.
	WARNING
	// Recoverable error conditions.
	ERROR
	// Fatal error conditions.
	FATAL
)

var (
	// LevelNames is a map of the names ( {level:name} ) of each verbosity level.
	LevelNames = map[Verbosity]string{
		DEBUG:     "dbg",
		INFO:      "inf",
		IMPORTANT: "imp",
		WARNING:   "war",
		ERROR:     "err",
		FATAL:     "!!!",
	}
	// LevelColors is a map of the colors ( {level:color} ) of each verbosity level.
	LevelColors = map[Verbosity]string{
		DEBUG:     tui.DIM + tui.FOREBLACK + tui.BACKDARKGRAY,
		INFO:      tui.FOREWHITE + tui.BACKGREEN,
		IMPORTANT: tui.FOREWHITE + tui.BACKLIGHTBLUE,
		WARNING:   tui.FOREWHITE + tui.BACKYELLOW,
		ERROR:     tui.FOREWHITE + tui.BACKRED,
		FATAL:     tui.FOREWHITE + tui.BACKRED + tui.BOLD,
	}
)
