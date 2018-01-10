package session

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evilsocket/bettercap-ng/core"

	"github.com/chzyer/readline"
)

func (s *Session) helpHandler(args []string, sess *Session) error {
	filter := ""
	if len(args) == 2 {
		filter = args[1]
	}

	if filter == "" {
		fmt.Println()
		fmt.Printf(core.Bold("MAIN COMMANDS\n\n"))
		for _, h := range s.CoreHandlers {
			fmt.Printf("  "+core.Yellow("%"+strconv.Itoa(s.HelpPadding)+"s")+" : %s\n", h.Name, h.Description)
		}

		fmt.Printf(core.Bold("\nMODULES\n"))

		for _, m := range s.Modules {
			status := ""
			if m.Running() {
				status = core.Green("running")
			} else {
				status = core.Red("not running")
			}
			fmt.Printf("  "+core.Yellow("%"+strconv.Itoa(s.HelpPadding)+"s")+" > %s\n", m.Name(), status)
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
		for _, h := range m.Handlers() {
			fmt.Printf(h.Help(s.HelpPadding))
		}

		params := m.Parameters()
		if len(params) > 0 {
			fmt.Printf("\n  Parameters\n\n")
			for _, p := range params {
				fmt.Printf(p.Help(s.HelpPadding))
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
		fmt.Printf("[%s] %s (%s)\n", core.Green("active"), m.Name(), core.Dim(m.Description()))
		params := m.Parameters()
		if len(params) > 0 {
			for _, p := range params {
				_, p.Value = s.Env.Get(p.Name)
				fmt.Printf("  %s: '%s'\n", p.Name, core.Yellow(p.Value))
			}
			fmt.Println()
		}
	}

	return nil
}

func (s *Session) exitHandler(args []string, sess *Session) error {
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
	if key == "*" {
		prev_ns := ""

		fmt.Println()
		for _, k := range s.Env.Sorted() {
			ns := ""
			toks := strings.Split(k, ".")
			if len(toks) > 0 {
				ns = toks[0]
			}

			if ns != prev_ns {
				fmt.Println()
				prev_ns = ns
			}

			fmt.Printf("  %"+strconv.Itoa(s.Env.Padding)+"s: '%s'\n", k, s.Env.Storage[k])
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
	readline.ClearScreen(s.Input.Stdout())
	return nil
}

func (s *Session) registerCoreHandlers() {
	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("help",
		"^(help|\\?)$",
		"Display list of available commands.",
		s.helpHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("help MODULE",
		"^(help|\\?) (.+)$",
		"Show module specific help.",
		s.helpHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("active",
		"^active$",
		"Show information about active modules.",
		s.activeHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("exit",
		"^(q|quit|e|exit)$",
		"Close the session and exit.",
		s.exitHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("sleep SECONDS",
		"^sleep\\s+(\\d+)$",
		"Sleep for the given amount of seconds.",
		s.sleepHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("get NAME",
		"^get\\s+(.+)",
		"Get the value of variable NAME, use * for all.",
		s.getHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("set NAME VALUE",
		"^set\\s+([^\\s]+)\\s+(.+)",
		"Set the VALUE of variable NAME.",
		s.setHandler))

	s.CoreHandlers = append(s.CoreHandlers, NewCommandHandler("clear",
		"^(clear|cls)$",
		"Clear the screen.",
		s.clsHandler))
}
