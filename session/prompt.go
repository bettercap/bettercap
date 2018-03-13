package session

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/core"

	"github.com/dustin/go-humanize"
)

const (
	PromptVariable = "$"
	DefaultPrompt  = "{by}{fw}{cidr} {fb}> {env.iface.ipv4} {reset} {bold}» {reset}"
)

var PromptCallbacks = map[string]func(s *Session) string{
	"{cidr}": func(s *Session) string {
		return s.Interface.CIDR()
	},
	"{net.sent}": func(s *Session) string {
		return fmt.Sprintf("%d", s.Queue.Stats.Sent)
	},
	"{net.sent.human}": func(s *Session) string {
		return humanize.Bytes(s.Queue.Stats.Sent)
	},
	"{net.received}": func(s *Session) string {
		return fmt.Sprintf("%d", s.Queue.Stats.Received)
	},
	"{net.received.human}": func(s *Session) string {
		return humanize.Bytes(s.Queue.Stats.Received)
	},
	"{net.packets}": func(s *Session) string {
		return fmt.Sprintf("%d", s.Queue.Stats.PktReceived)
	},
	"{net.errors}": func(s *Session) string {
		return fmt.Sprintf("%d", s.Queue.Stats.Errors)
	},
}

var envRe = regexp.MustCompile("{env\\.([^}]+)}")

type Prompt struct {
}

func NewPrompt() Prompt {
	return Prompt{}
}

func (p Prompt) Render(s *Session) string {
	found, prompt := s.Env.Get(PromptVariable)
	if found == false {
		prompt = DefaultPrompt
	}

	// these are here because if colors are disabled,
	// we need the updated core.* variables
	var effects = map[string]string{
		"{bold}":  core.BOLD,
		"{dim}":   core.DIM,
		"{r}":     core.RED,
		"{g}":     core.GREEN,
		"{b}":     core.BLUE,
		"{y}":     core.YELLOW,
		"{fb}":    core.FG_BLACK,
		"{fw}":    core.FG_WHITE,
		"{bdg}":   core.BG_DGRAY,
		"{br}":    core.BG_RED,
		"{bg}":    core.BG_GREEN,
		"{by}":    core.BG_YELLOW,
		"{blb}":   core.BG_LBLUE, // Ziggy this is for you <3
		"{reset}": core.RESET,
	}

	for tok, effect := range effects {
		prompt = strings.Replace(prompt, tok, effect, -1)
	}

	for tok, cb := range PromptCallbacks {
		prompt = strings.Replace(prompt, tok, cb(s), -1)
	}

	m := envRe.FindAllString(prompt, -1)
	for _, match := range m {
		name := strings.Trim(strings.Replace(match, "env.", "", -1), "{}")
		_, value := s.Env.Get(name)
		prompt = strings.Replace(prompt, match, value, -1)
	}

	// make sure an user error does not screw all terminal
	if strings.HasPrefix(prompt, core.RESET) == false {
		prompt += core.RESET
	}

	return prompt
}
