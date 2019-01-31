package session

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bettercap/bettercap/caplets"

	"github.com/bettercap/readline"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
)

func containsCapitals(s string) bool {
	for _, ch := range s {
		if ch < 133 && ch > 101 {
			return false
		}
	}
	return true
}

func (s *Session) setupReadline() (err error) {
	prefixCompleters := make([]readline.PrefixCompleterInterface, 0)
	for _, h := range s.CoreHandlers {
		if h.Completer == nil {
			prefixCompleters = append(prefixCompleters, readline.PcItem(h.Name))
		} else {
			prefixCompleters = append(prefixCompleters, h.Completer)
		}
	}

	tree := make(map[string][]string)
	for _, m := range s.Modules {
		for _, h := range m.Handlers() {
			parts := strings.Split(h.Name, " ")
			name := parts[0]

			if _, found := tree[name]; !found {
				tree[name] = []string{}
			}

			var appendedOption = strings.Join(parts[1:], " ")

			if len(appendedOption) > 0 && !containsCapitals(appendedOption) {
				tree[name] = append(tree[name], appendedOption)
			}
		}
	}

	for _, caplet := range caplets.List() {
		tree[caplet.Name] = []string{}
	}

	for root, subElems := range tree {
		item := readline.PcItem(root)
		item.Children = []readline.PrefixCompleterInterface{}
		for _, child := range subElems {
			item.Children = append(item.Children, readline.PcItem(child))
		}
		prefixCompleters = append(prefixCompleters, item)
	}

	history := ""
	if !*s.Options.NoHistory {
		history, _ = fs.Expand(HistoryFile)
	}

	cfg := readline.Config{
		HistoryFile:     history,
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
		AutoComplete:    readline.NewPrefixCompleter(prefixCompleters...),
	}

	s.Input, err = readline.NewEx(&cfg)
	return err
}

func (s *Session) startNetMon() {
	// keep reading network events in order to add / update endpoints
	go func() {
		for event := range s.Queue.Activities {
			if !s.Active {
				return
			}

			if s.IsOn("net.recon") && event.Source {
				addr := event.IP.String()
				mac := event.MAC.String()

				existing := s.Lan.AddIfNew(addr, mac)
				if existing != nil {
					existing.LastSeen = time.Now()
				} else {
					existing, _ = s.Lan.Get(mac)
				}

				if existing != nil && event.Meta != nil {
					existing.OnMeta(event.Meta)
				}
			}
		}
	}()
}

func (s *Session) setupSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		s.Events.Log(log.WARNING, "Got SIGTERM")
		s.Close()
		os.Exit(0)
	}()
}

func (s *Session) setupEnv() {
	s.Env.Set("iface.index", fmt.Sprintf("%d", s.Interface.Index))
	s.Env.Set("iface.name", s.Interface.Name())
	s.Env.Set("iface.ipv4", s.Interface.IpAddress)
	s.Env.Set("iface.ipv6", s.Interface.Ip6Address)
	s.Env.Set("iface.mac", s.Interface.HwAddress)
	s.Env.Set("gateway.address", s.Gateway.IpAddress)
	s.Env.Set("gateway.mac", s.Gateway.HwAddress)

	if found, v := s.Env.Get(PromptVariable); !found || v == "" {
		s.Env.Set(PromptVariable, DefaultPrompt)
	}

	dbg := "false"
	if *s.Options.Debug {
		dbg = "true"
	}
	s.Env.WithCallback("log.debug", dbg, func(newValue string) {
		newDbg := false
		if newValue == "true" {
			newDbg = true
		}
		s.Events.SetDebug(newDbg)
	})

	silent := "false"
	if *s.Options.Silent {
		silent = "true"
	}
	s.Env.WithCallback("log.silent", silent, func(newValue string) {
		newSilent := false
		if newValue == "true" {
			newSilent = true
		}
		s.Events.SetSilent(newSilent)
	})
}
