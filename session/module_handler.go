package session

import (
	"fmt"
	"github.com/evilsocket/bettercap/core"
	"regexp"
	"strconv"
)

type ModuleHandler struct {
	Name        string
	Description string
	Parser      *regexp.Regexp
	Exec        func(args []string) error
}

func NewModuleHandler(name string, expr string, desc string, exec func(args []string) error) ModuleHandler {
	h := ModuleHandler{
		Name:        name,
		Description: desc,
		Parser:      nil,
		Exec:        exec,
	}

	if expr != "" {
		h.Parser = regexp.MustCompile(expr)
	}

	return h
}

func (h *ModuleHandler) Help(padding int) string {
	return fmt.Sprintf("  "+core.Bold("%"+strconv.Itoa(padding)+"s")+" : %s\n", h.Name, h.Description)
}

func (h *ModuleHandler) Parse(line string) (bool, []string) {
	if h.Parser == nil {
		if line == h.Name {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		result := h.Parser.FindStringSubmatch(line)
		if len(result) == h.Parser.NumSubexp()+1 {
			return true, result[1:len(result)]
		} else {
			return false, nil
		}
	}
}
