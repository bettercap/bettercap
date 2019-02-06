package session

import (
	"fmt"
	"strings"

	"github.com/evilsocket/islazy/tui"

	"github.com/dustin/go-humanize"
)

const (
	PromptVariable       = "$"
	DefaultPrompt        = "{by}{fw}{cidr} {fb}> {env.iface.ipv4} {reset} {bold}» {reset}"
	DefaultPromptMonitor = "{by}{fb} {env.iface.name} {reset} {bold}» {reset}"
)

var (
	effects         = map[string]string{}
	PromptCallbacks = map[string]func(s *Session) string{
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
)

type Prompt struct {
}

func NewPrompt() Prompt {
	// these are here because if colors are disabled,
	// we need the updated tui.* variables
	effects = map[string]string{
		"{bold}":  tui.BOLD,
		"{dim}":   tui.DIM,
		"{r}":     tui.RED,
		"{g}":     tui.GREEN,
		"{b}":     tui.BLUE,
		"{y}":     tui.YELLOW,
		"{fb}":    tui.FOREBLACK,
		"{fw}":    tui.FOREWHITE,
		"{bdg}":   tui.BACKDARKGRAY,
		"{br}":    tui.BACKRED,
		"{bg}":    tui.BACKGREEN,
		"{by}":    tui.BACKYELLOW,
		"{blb}":   tui.BACKLIGHTBLUE, // Ziggy this is for you <3
		"{reset}": tui.RESET,
	}
	return Prompt{}
}

func (p Prompt) Render(s *Session) string {
	found, prompt := s.Env.Get(PromptVariable)
	if !found {
		prompt = DefaultPrompt
	}

	for tok, effect := range effects {
		prompt = strings.Replace(prompt, tok, effect, -1)
	}

	for tok, cb := range PromptCallbacks {
		prompt = strings.Replace(prompt, tok, cb(s), -1)
	}

	// make sure an user error does not screw all terminal
	if !strings.HasPrefix(prompt, tui.RESET) {
		prompt += tui.RESET
	}

	return prompt
}
