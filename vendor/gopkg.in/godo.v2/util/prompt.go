package util

import (
	"bufio"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
)

// Prompt prompts user for input with default value.
func Prompt(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return text
}

// PromptPassword prompts user for password input.
func PromptPassword(prompt string) string {
	fmt.Printf(prompt)
	b, err := gopass.GetPasswd()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(b)
}
