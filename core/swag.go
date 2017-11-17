package core

// https://misc.flogisoft.com/bash/tip_colors_and_formatting
const (
	BOLD = "\033[1m"
	DIM  = "\033[2m"

	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"

	RESET = "\033[0m"
)

const ON = GREEN + "✔" + RESET
const OFF = RED + "✘" + RESET

func Bold(s string) string {
	return BOLD + s + RESET
}

func Dim(s string) string {
	return DIM + s + RESET
}

func Red(s string) string {
	return RED + s + RESET
}

func Green(s string) string {
	return GREEN + s + RESET
}

func Yellow(s string) string {
	return YELLOW + s + RESET
}
