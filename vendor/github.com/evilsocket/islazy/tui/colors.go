package tui

import (
	"os"
	"strings"
)

// https://misc.flogisoft.com/bash/tip_colors_and_formatting
var (
	// effects
	BOLD  = "\033[1m"
	DIM   = "\033[2m"
	RESET = "\033[0m"
	// colors
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	BLUE   = "\033[34m"
	YELLOW = "\033[33m"
	// foreground colors
	FOREBLACK = "\033[30m"
	FOREWHITE = "\033[97m"
	// background colors
	BACKDARKGRAY  = "\033[100m"
	BACKRED       = "\033[41m"
	BACKGREEN     = "\033[42m"
	BACKYELLOW    = "\033[43m"
	BACKLIGHTBLUE = "\033[104m"

	ctrl = []string{"\x033", "\\e", "\x1b"}
)

// Effects returns true if colors and effects are supported
// on the current terminal.
func Effects() bool {
	if term := os.Getenv("TERM"); term == "" {
		return false
	} else if term == "dumb" {
		return false
	}
	return true
}

// Disable will disable all colors and effects.
func Disable() {
	BOLD = ""
	DIM = ""
	RESET = ""
	RED = ""
	GREEN = ""
	BLUE = ""
	YELLOW = ""
	FOREBLACK = ""
	FOREWHITE = ""
	BACKDARKGRAY = ""
	BACKRED = ""
	BACKGREEN = ""
	BACKYELLOW = ""
	BACKLIGHTBLUE = ""
}

// HasEffect returns true if the string has any shell control codes in it.
func HasEffect(s string) bool {
	for _, ch := range ctrl {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}

// Wrap wraps a string with an effect or color and appends a reset control code.
func Wrap(e, s string) string {
	return e + s + RESET
}

// Bold makes the string Bold.
func Bold(s string) string {
	return Wrap(BOLD, s)
}

// Dim makes the string Diminished.
func Dim(s string) string {
	return Wrap(DIM, s)
}

// Red makes the string Red.
func Red(s string) string {
	return Wrap(RED, s)
}

// Green makes the string Green.
func Green(s string) string {
	return Wrap(GREEN, s)
}

// Blue makes the string Green.
func Blue(s string) string {
	return Wrap(BLUE, s)
}

// Yellow makes the string Green.
func Yellow(s string) string {
	return Wrap(YELLOW, s)
}
