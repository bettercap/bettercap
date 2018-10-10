package log

import (
	"strconv"
	"time"

	"github.com/evilsocket/islazy/tui"
)

var (
	// Tokens is a map of the tokens that can be used in Format
	// to insert values returned by the execution of a callback.
	Tokens = map[string]func() string{
		"{date}": func() string {
			return time.Now().Format(DateFormat)
		},
		"{time}": func() string {
			return time.Now().Format(TimeFormat)
		},
		"{datetime}": func() string {
			return time.Now().Format(DateTimeFormat)
		},
		"{level:value}": func() string {
			return strconv.Itoa(int(currLevel))
		},
		"{level:name}": func() string {
			return LevelNames[currLevel]
		},
		"{level:color}": func() string {
			return LevelColors[currLevel]
		},
		"{message}": func() string {
			return currMessage
		},
	}
	// Effects is a map of the tokens that can be used in Format to
	// change the properties of the text.
	Effects = map[string]string{
		"{bold}":        tui.BOLD,
		"{dim}":         tui.DIM,
		"{red}":         tui.RED,
		"{green}":       tui.GREEN,
		"{blue}":        tui.BLUE,
		"{yellow}":      tui.YELLOW,
		"{f:black}":     tui.FOREBLACK,
		"{f:white}":     tui.FOREWHITE,
		"{b:darkgray}":  tui.BACKDARKGRAY,
		"{b:red}":       tui.BACKRED,
		"{b:green}":     tui.BACKGREEN,
		"{b:yellow}":    tui.BACKYELLOW,
		"{b:lightblue}": tui.BACKLIGHTBLUE,
		"{reset}":       tui.RESET,
	}
	// DateFormat is the default date format being used when filling the {date} log token.
	DateFormat = "06-Jan-02"
	// TimeFormat is the default time format being used when filling the {time} or {datetime} log tokens.
	TimeFormat = "15:04:05"
	// DateTimeFormat is the default date and time format being used when filling the {datetime} log token.
	DateTimeFormat = "2006-01-02 15:04:05"
	// Format is the default format being used when logging.
	Format = "{datetime} {level:color}{level:name}{reset} {message}"
)
