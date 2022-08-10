package session

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"

	"github.com/bettercap/readline"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func (s *Session) generalHelp() {
	fmt.Println()

	maxLen := 0
	for _, h := range s.CoreHandlers {
		len := len(h.Name)
		if len > maxLen {
			maxLen = len
		}
	}
	pad := "%" + strconv.Itoa(maxLen) + "s"

	for _, h := range s.CoreHandlers {
		s.Events.Printf("  "+tui.Yellow(pad)+" : %s\n", h.Name, h.Description)
	}

	s.Events.Printf("%s\n", tui.Bold("\nModules\n"))

	maxLen = 0
	for _, m := range s.Modules {
		len := len(m.Name())
		if len > maxLen {
			maxLen = len
		}
	}
	pad = "%" + strconv.Itoa(maxLen) + "s"

	for _, m := range s.Modules {
		status := ""
		if m.Running() {
			status = tui.Green("running")
		} else {
			status = tui.Red("not running")
		}
		s.Events.Printf("  "+tui.Yellow(pad)+" > %s\n", m.Name(), status)
	}

	fmt.Println()
}

func (s *Session) moduleHelp(filter string) error {
	err, m := s.Module(filter)
	if err != nil {
		return err
	}

	fmt.Println()
	status := ""
	if m.Running() {
		status = tui.Green("running")
	} else {
		status = tui.Red("not running")
	}
	s.Events.Printf("%s (%s): %s\n\n", tui.Yellow(m.Name()), status, tui.Dim(m.Description()))

	maxLen := 0
	handlers := m.Handlers()
	for _, h := range handlers {
		len := len(h.Name)
		if len > maxLen {
			maxLen = len
		}
	}

	for _, h := range handlers {
		s.Events.Printf("%s", h.Help(maxLen))
	}
	fmt.Println()

	params := m.Parameters()
	if len(params) > 0 {
		v := make([]*ModuleParam, 0)
		maxLen := 0
		for _, h := range params {
			len := len(h.Name)
			if len > maxLen {
				maxLen = len
			}
			v = append(v, h)
		}

		sort.Slice(v, func(i, j int) bool {
			return v[i].Name < v[j].Name
		})

		s.Events.Printf("  Parameters\n\n")
		for _, p := range v {
			s.Events.Printf("%s", p.Help(maxLen))
		}
		fmt.Println()
	}

	return nil
}

func (s *Session) helpHandler(args []string, sess *Session) error {
	filter := ""
	if len(args) == 2 {
		filter = str.Trim(args[1])
	}

	if filter == "" {
		s.generalHelp()
	} else {
		if err := s.moduleHelp(filter); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) activeHandler(args []string, sess *Session) error {
	for _, m := range s.Modules {
		if !m.Running() {
			continue
		}

		s.Events.Printf("%s (%s)\n", tui.Bold(m.Name()), tui.Dim(m.Description()))
		params := m.Parameters()
		if len(params) > 0 {
			fmt.Println()
			for _, p := range params {
				_, val := s.Env.Get(p.Name)
				s.Events.Printf("  %s : %s\n", tui.Yellow(p.Name), val)
			}
		}

		fmt.Println()
	}

	return nil
}

func (s *Session) exitHandler(args []string, sess *Session) error {
	// notify any listener that the session is about to end
	s.Events.Add("session.stopped", nil)

	for _, mod := range s.Modules {
		if mod.Running() {
			mod.Stop()
		}
	}

	s.Active = false
	s.Input.Close()
	return nil
}

func (s *Session) sleepHandler(args []string, sess *Session) error {
	if secs, err := strconv.Atoi(args[0]); err == nil {
		time.Sleep(time.Duration(secs) * time.Second)
		return nil
	} else {
		return err
	}
}

func (s *Session) getHandler(args []string, sess *Session) error {
	key := args[0]
	if strings.Contains(key, "*") {
		prev_ns := ""

		fmt.Println()
		last := len(key) - 1
		prefix := key[:last]
		sortedKeys := s.Env.Sorted()
		padding := 0

		for _, k := range sortedKeys {
			l := len(k)
			if l > padding {
				padding = l
			}
		}

		for _, k := range sortedKeys {
			if strings.HasPrefix(k, prefix) {
				ns := ""
				toks := strings.Split(k, ".")
				if len(toks) > 0 {
					ns = toks[0]
				}

				if ns != prev_ns {
					fmt.Println()
					prev_ns = ns
				}

				s.Events.Printf("  %"+strconv.Itoa(padding)+"s: '%s'\n", k, s.Env.Data[k])
			}
		}
		fmt.Println()
	} else if found, value := s.Env.Get(key); found {
		fmt.Println()
		s.Events.Printf("  %s: '%s'\n", key, value)
		fmt.Println()
	} else {
		return fmt.Errorf("%s not found", key)
	}

	return nil
}

func (s *Session) setHandler(args []string, sess *Session) error {
	key := args[0]
	value := args[1]

	if value == "\"\"" || value == "''" {
		value = ""
	}

	s.Env.Set(key, value)
	return nil
}

func (s *Session) readHandler(args []string, sess *Session) error {
	key := args[0]
	prompt := args[1]

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", prompt)

	value, _ := reader.ReadString('\n')
	value = str.Trim(value)
	if value == "\"\"" || value == "''" {
		value = ""
	}

	s.Env.Set(key, value)
	return nil
}

func (s *Session) clsHandler(args []string, sess *Session) error {
	cmd := "clear"
	if runtime.GOOS == "windows" {
		cmd = "cls"
	}

	c := exec.Command(cmd)
	c.Stdout = os.Stdout
	c.Run()
	return nil
}

func (s *Session) includeHandler(args []string, sess *Session) error {
	return s.RunCaplet(args[0])
}

func (s *Session) shHandler(args []string, sess *Session) error {
	out, err := core.Shell(args[0])
	if err == nil {
		s.Events.Printf("%s\n", out)
	}
	return err
}

func normalizeMac(mac string) string {
	var parts []string
	if strings.ContainsRune(mac, '-') {
		parts = strings.Split(mac, "-")
	} else {
		parts = strings.Split(mac, ":")
	}

	for i, p := range parts {
		if len(p) < 2 {
			parts[i] = "0" + p
		}
	}
	return strings.ToLower(strings.Join(parts, ":"))
}

func (s *Session) propagateAlias(mac, alias string) {
	mac = normalizeMac(mac)

	s.Aliases.Set(mac, alias)

	if dev, found := s.BLE.Get(mac); found {
		dev.Alias = alias
	}

	if dev, found := s.HID.Get(mac); found {
		dev.Alias = alias
	}

	if ap, found := s.WiFi.Get(mac); found {
		ap.Alias = alias
	}

	if sta, found := s.WiFi.GetClient(mac); found {
		sta.Alias = alias
	}

	if host, found := s.Lan.Get(mac); found {
		host.Alias = alias
	}
}

func (s *Session) aliasHandler(args []string, sess *Session) error {
	mac := args[0]
	alias := str.Trim(args[1])
	if alias == "\"\"" || alias == "''" {
		alias = ""
	}
	s.propagateAlias(mac, alias)
	return nil
}

func (s *Session) addHandler(h CommandHandler, c *readline.PrefixCompleter) {
	h.Completer = c
	s.CoreHandlers = append(s.CoreHandlers, h)
}

func (s *Session) registerCoreHandlers() {
	s.addHandler(NewCommandHandler("help MODULE",
		"^(help|\\?)(.*)$",
		"List available commands or show module specific help if no module name is provided.",
		s.helpHandler),
		readline.PcItem("help", readline.PcItemDynamic(func(prefix string) []string {
			prefix = str.Trim(prefix[4:])
			modNames := []string{""}
			for _, m := range s.Modules {
				if prefix == "" || strings.HasPrefix(m.Name(), prefix) {
					modNames = append(modNames, m.Name())
				}
			}
			return modNames
		})))

	s.addHandler(NewCommandHandler("active",
		"^active$",
		"Show information about active modules.",
		s.activeHandler),
		readline.PcItem("active"))

	s.addHandler(NewCommandHandler("quit",
		"^(q|quit|e|exit)$",
		"Close the session and exit.",
		s.exitHandler),
		readline.PcItem("quit"))

	s.addHandler(NewCommandHandler("sleep SECONDS",
		"^sleep\\s+(\\d+)$",
		"Sleep for the given amount of seconds.",
		s.sleepHandler),
		readline.PcItem("sleep"))

	s.addHandler(NewCommandHandler("get NAME",
		"^get\\s+(.+)",
		"Get the value of variable NAME, use * alone for all, or NAME* as a wildcard.",
		s.getHandler),
		readline.PcItem("get", readline.PcItemDynamic(func(prefix string) []string {
			prefix = str.Trim(prefix[3:])
			varNames := []string{""}
			for key := range s.Env.Data {
				if prefix == "" || strings.HasPrefix(key, prefix) {
					varNames = append(varNames, key)
				}
			}
			return varNames
		})))

	s.addHandler(NewCommandHandler("set NAME VALUE",
		"^set\\s+([^\\s]+)\\s+(.+)",
		"Set the VALUE of variable NAME.",
		s.setHandler),
		readline.PcItem("set", readline.PcItemDynamic(func(prefix string) []string {
			prefix = str.Trim(prefix[3:])
			varNames := []string{""}
			for key := range s.Env.Data {
				if prefix == "" || strings.HasPrefix(key, prefix) {
					varNames = append(varNames, key)
				}
			}
			return varNames
		})))

	s.addHandler(NewCommandHandler("read VARIABLE PROMPT",
		`^read\s+([^\s]+)\s+(.+)$`,
		"Show a PROMPT to ask the user for input that will be saved inside VARIABLE.",
		s.readHandler),
		readline.PcItem("read"))

	s.addHandler(NewCommandHandler("clear",
		"^(clear|cls)$",
		"Clear the screen.",
		s.clsHandler),
		readline.PcItem("clear"))

	s.addHandler(NewCommandHandler("include CAPLET",
		"^include\\s+(.+)",
		"Load and run this caplet in the current session.",
		s.includeHandler),
		readline.PcItem("include", readline.PcItemDynamic(func(prefix string) []string {
			prefix = str.Trim(prefix[8:])
			if prefix == "" {
				prefix = "."
			}

			files, _ := filepath.Glob(prefix + "*")
			return files
		})))

	s.addHandler(NewCommandHandler("! COMMAND",
		"^!\\s*(.+)$",
		"Execute a shell command and print its output.",
		s.shHandler),
		readline.PcItem("!"))

	s.addHandler(NewCommandHandler("alias MAC NAME",
		"^alias\\s+([a-fA-F0-9:]{14,17})\\s*(.*)",
		"Assign an alias to a given endpoint given its MAC address.",
		s.aliasHandler),
		readline.PcItem("alias", readline.PcItemDynamic(func(prefix string) []string {
			prefix = str.Trim(prefix[5:])
			macs := []string{""}
			s.Lan.EachHost(func(mac string, e *network.Endpoint) {
				if prefix == "" || strings.HasPrefix(mac, prefix) {
					macs = append(macs, mac)
				}
			})
			return macs
		})))

}
