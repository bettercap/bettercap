package core

// https://misc.flogisoft.com/bash/tip_colors_and_formatting
const (
	BOLD = "\033[1m"
	DIM  = "\033[2m"

	RED    = "\033[31m"
	GREEN  = "\033[32m"
	BLUE   = "\033[34m"
	YELLOW = "\033[33m"

	FG_BLACK = "\033[30m"
	FG_WHITE = "\033[97m"

	BG_DGRAY  = "\033[100m"
	BG_RED    = "\033[41m"
	BG_GREEN  = "\033[42m"
	BG_YELLOW = "\033[43m"
	BG_LBLUE  = "\033[104m"

	RESET = "\033[0m"
)

const ON = GREEN + "✔" + RESET
const OFF = RED + "✘" + RESET

// W for Wrap
func W(e, s string) string {
	return e + s + RESET
}

func Bold(s string) string {
	return W(BOLD, s)
}

func Dim(s string) string {
	return W(DIM, s)
}

func Red(s string) string {
	return W(RED, s)
}

func Green(s string) string {
	return W(GREEN, s)
}

func Yellow(s string) string {
	return W(YELLOW, s)
}
