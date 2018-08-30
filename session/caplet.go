package session

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
)

const (
	CapletSuffix = ".cap"
)

type Caplet struct {
	Path string
	Code []string
}

var (
	CapletLoadPaths = []string{
		"./caplets/",
		"/usr/local/share/bettercap/caplets/",
	}

	cache     = make(map[string]*Caplet)
	cacheLock = sync.Mutex{}
)

func init() {
	for _, path := range core.SepSplit(core.Trim(os.Getenv("CAPSPATH")), ":") {
		if path = core.Trim(path); len(path) > 0 {
			CapletLoadPaths = append(CapletLoadPaths, path)
		}
	}

	for i, path := range CapletLoadPaths {
		CapletLoadPaths[i], _ = filepath.Abs(path)
	}
}

func LoadCaplet(name string) (error, *Caplet) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if caplet, found := cache[name]; found {
		return nil, caplet
	}

	names := []string{name}
	if !strings.HasSuffix(name, CapletSuffix) {
		names = append(names, name+CapletSuffix)
	}

	for _, path := range CapletLoadPaths {
		if !strings.HasSuffix(name, CapletSuffix) {
			name += CapletSuffix
		}
		names = append(names, filepath.Join(path, name))
	}

	for _, filename := range names {
		if core.Exists(filename) {
			cap := &Caplet{
				Path: filename,
				Code: make([]string, 0),
			}

			I.Events.Log(core.INFO, "reading from caplet %s ...", filename)
			input, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("error reading caplet %s: %v", filename, err), nil
			}
			defer input.Close()

			scanner := bufio.NewScanner(input)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := core.Trim(scanner.Text())
				if line == "" || line[0] == '#' {
					continue
				}
				cap.Code = append(cap.Code, line)
			}

			cache[name] = cap
			return nil, cap
		}
	}

	return fmt.Errorf("caplet %s not found", name), nil
}

func parseCapletCommand(line string) (is bool, caplet *Caplet, argv []string) {
	file := core.Trim(line)
	parts := strings.Split(file, " ")
	argc := len(parts)
	argv = make([]string, 0)
	// check for any arguments
	if argc > 1 {
		file = core.Trim(parts[0])
		if argc >= 2 {
			argv = parts[1:]
		}
	}

	if err, cap := LoadCaplet(file); err == nil {
		return true, cap, argv
	}

	return false, nil, nil
}

func (cap *Caplet) Eval(s *Session, argv []string) error {
	// the caplet might include other files (include directive, proxy modules, etc),
	// temporarily change the working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error while getting current working directory: %v", err)
	}

	capPath := filepath.Dir(cap.Path)
	if err := os.Chdir(capPath); err != nil {
		return fmt.Errorf("error while changing current working directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			s.Events.Log(core.ERROR, "error while restoring working directory: %v", err)
		}
	}()

	if argv == nil {
		argv = []string{}
	}

	for _, line := range cap.Code {
		// replace $0 with argv[0], $1 with argv[1] and so on
		for i, arg := range argv {
			what := fmt.Sprintf("$%d", i)
			line = strings.Replace(line, what, arg, -1)
		}

		if err = s.Run(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}
