package session

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"

	"github.com/bettercap/readline"
)

const (
	IPv4Validator = `^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`
	IPv6Validator = `^[:a-fA-F0-9]{6,}$`
	IPValidator   = `^[\.:a-fA-F0-9]{6,}$`
)

type ModuleHandler struct {
	*sync.Mutex

	Name        string
	Description string
	Parser      *regexp.Regexp
	Completer   *readline.PrefixCompleter
	exec        func(args []string) error
}

func NewModuleHandler(name string, expr string, desc string, exec func(args []string) error) ModuleHandler {
	h := ModuleHandler{
		Mutex:       &sync.Mutex{},
		Name:        name,
		Description: desc,
		Parser:      nil,
		exec:        exec,
	}

	if expr != "" {
		h.Parser = regexp.MustCompile(expr)
	}

	return h
}

func (h *ModuleHandler) Complete(name string, cb func(prefix string) []string) {
	h.Completer = readline.PcItem(name, readline.PcItemDynamic(func(prefix string) []string {
		prefix = str.Trim(prefix[len(name):])
		return cb(prefix)
	}))
}

func (h *ModuleHandler) Help(padding int) string {
	return fmt.Sprintf("  "+tui.Bold("%"+strconv.Itoa(padding)+"s")+" : %s\n", h.Name, h.Description)
}

func (h *ModuleHandler) Parse(line string) (bool, []string) {
	if h.Parser == nil {
		if line == h.Name {
			return true, nil
		}
		return false, nil
	}
	result := h.Parser.FindStringSubmatch(line)
	if len(result) == h.Parser.NumSubexp()+1 {
		return true, result[1:]
	}
	return false, nil
}

func (h *ModuleHandler) Exec(args []string) error {
	h.Lock()
	defer h.Unlock()
	return h.exec(args)
}

type JSONModuleHandler struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parser      string `json:"parser"`
}

func (h ModuleHandler) MarshalJSON() ([]byte, error) {
	j := JSONModuleHandler{
		Name:        h.Name,
		Description: h.Description,
	}
	if h.Parser != nil {
		j.Parser = h.Parser.String()
	}
	return json.Marshal(j)
}
