package session

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
	CapletsTree     = make(map[string][]string)
	CapletLoadPaths = []string{
		"./caplets/",
		"/usr/share/bettercap/caplets/",
	}

	cache     = make(map[string]*Caplet)
	cacheLock = sync.Mutex{}
)

func buildCapletsTree(path string, prefix string) {
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, CapletSuffix) {
			base := strings.TrimPrefix(path, prefix)
			name := strings.Replace(filename, CapletSuffix, "", -1)
			CapletsTree[base+name] = []string{}
		} else if file.IsDir() {
			buildCapletsTree(filepath.Join(path, filename)+"/", prefix)
		}
	}
}

func init() {
	for _, path := range core.SepSplit(core.Trim(os.Getenv("CAPSPATH")), ":") {
		if path = core.Trim(path); len(path) > 0 {
			CapletLoadPaths = append(CapletLoadPaths, path)
		}
	}

	for i, path := range CapletLoadPaths {
		CapletLoadPaths[i], _ = filepath.Abs(path)
	}

	for _, path := range CapletLoadPaths {
		buildCapletsTree(path, path)
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
