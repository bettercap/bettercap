package session

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"

	"github.com/bettercap/readline"
)

func (s *Session) helpHandler(args []string, sess *Session) error {
	filter := ""
	if len(args) == 2 {
		filter = core.Trim(args[1])
	}

	if filter == "" {
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
			fmt.Printf("  "+core.Yellow(pad)+" : %s\n", h.Name, h.Description)
		}

		fmt.Printf(core.Bold("\nModules\n\n"))

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
				status = core.Green("running")
			} else {
				status = core.Red("not running")
			}
			fmt.Printf("  "+core.Yellow(pad)+" > %s\n", m.Name(), status)
		}

		fmt.Println()

	} else {
		err, m := s.Module(filter)
		if err != nil {
			return err
		}

		fmt.Println()
		status := ""
		if m.Running() {
			status = core.Green("running")
		} else {
			status = core.Red("not running")
		}
		fmt.Printf("%s (%s): %s\n\n", core.Yellow(m.Name()), status, core.Dim(m.Description()))

		maxLen := 0
		handlers := m.Handlers()
		for _, h := range handlers {
			len := len(h.Name)
			if len > maxLen {
				maxLen = len
			}
		}

		for _, h := range handlers {
			fmt.Printf(h.Help(maxLen))
		}
		fmt.Println()

		params := m.Parameters()
		if len(params) > 0 {
			fmt.Printf("  Parameters\n\n")
			maxLen := 0
			for _, h := range params {
				len := len(h.Name)
				if len > maxLen {
					maxLen = len
				}
			}

			for _, p := range params {
				fmt.Printf(p.Help(maxLen))
			}
			fmt.Println()
		}
	}

	return nil
}

func (s *Session) activeHandler(args []string, sess *Session) error {
	for _, m := range s.Modules {
		if m.Running() == false {
			continue
		}

		fmt.Printf("%s (%s)\n", core.Bold(m.Name()), core.Dim(m.Description()))
		params := m.Parameters()
		if len(params) > 0 {
			fmt.Println()
			for _, p := range params {
				_, val := s.Env.Get(p.Name)
				fmt.Printf("  %s : %s\n", core.Yellow(p.Name), val)
			}
		}

		fmt.Println()
	}

	return nil
}

func (s *Session) exitHandler(args []string, sess *Session) error {
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
		for _, k := range s.Env.Sorted() {
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

				fmt.Printf("  %"+strconv.Itoa(s.Env.Padding)+"s: '%s'\n", k, s.Env.Data[k])
			}
		}
		fmt.Println()
	} else if found, value := s.Env.Get(key); found == true {
		fmt.Println()
		fmt.Printf("  %s: '%s'\n", key, value)
		fmt.Println()
	} else {
		return fmt.Errorf("%s not found", key)
	}

	return nil
}

func (s *Session) setHandler(args []string, sess *Session) error {
	key := args[0]
	value := args[1]

	if value == "\"\"" {
		value = ""
	}

	s.Env.Set(key, value)
	return nil
}

func (s *Session) clsHandler(args []string, sess *Session) error {
	// fixes a weird bug which causes the screen not to be fully
	// cleared if a "clear; net.show" commands chain is executed
	// in the interactive session.
	for i := 0; i < 180; i++ {
		fmt.Println()
	}
	readline.ClearScreen(s.Input.Stdout())
	return nil
}

func (s *Session) includeHandler(args []string, sess *Session) error {
	return s.RunCaplet(args[0])
}

func (s *Session) shHandler(args []string, sess *Session) error {
	out, err := core.Shell(args[0])
	if err == nil {
		fmt.Printf("%s\n", out)
	}
	return err
}

func (s *Session) aliasHandler(args []string, sess *Session) error {
	mac := args[0]
	alias := core.Trim(args[1])

	if s.Lan.SetAliasFor(mac, alias) == true {
		return nil
	} else {
		return fmt.Errorf("Could not find endpoint %s", mac)
	}
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
			prefix = core.Trim(prefix[4:])
			modNames := []string{""}
			for _, m := range s.Modules {
				if prefix == "" || strings.HasPrefix(m.Name(), prefix) == true {
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
			prefix = core.Trim(prefix[3:])
			varNames := []string{""}
			for key := range s.Env.Data {
				if prefix == "" || strings.HasPrefix(key, prefix) == true {
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
			prefix = core.Trim(prefix[3:])
			varNames := []string{""}
			for key := range s.Env.Data {
				if prefix == "" || strings.HasPrefix(key, prefix) == true {
					varNames = append(varNames, key)
				}
			}
			return varNames
		})))

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
			prefix = core.Trim(prefix[8:])
			if prefix == "" {
				prefix = "."
			}

			files := []string{}
			files, _ = filepath.Glob(prefix + "*")
			return files
		})))

	s.addHandler(NewCommandHandler("! COMMAND",
		"^!\\s*(.+)$",
		"Execute a shell command and print its output.",
		s.shHandler),
		readline.PcItem("!"))

	s.addHandler(NewCommandHandler("alias MAC NAME",
		"^alias\\s+([a-fA-F0-9:]{17})\\s*(.*)",
		"Assign an alias to a given endpoint given its MAC address.",
		s.aliasHandler),
		readline.PcItem("alias", readline.PcItemDynamic(func(prefix string) []string {
			prefix = core.Trim(prefix[5:])
			macs := []string{""}
			s.Lan.EachHost(func(mac string, e *network.Endpoint) {
				if prefix == "" || strings.HasPrefix(mac, prefix) == true {
					macs = append(macs, mac)
				}
			})
			return macs
		})))

}
