package session

import "regexp"

type CommandHandler struct {
	Name        string
	Description string
	Parser      *regexp.Regexp
	Exec        func(args []string, s *Session) error
}

func NewCommandHandler(name string, expr string, desc string, exec func(args []string, s *Session) error) CommandHandler {
	return CommandHandler{
		Name:        name,
		Description: desc,
		Parser:      regexp.MustCompile(expr),
		Exec:        exec,
	}
}

func (h *CommandHandler) Parse(line string) (bool, []string) {
	result := h.Parser.FindStringSubmatch(line)
	if len(result) == h.Parser.NumSubexp()+1 {
		return true, result[1:]
	} else {
		return false, nil
	}
}
