package caplets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/fs"
)

var (
	cache     = make(map[string]*Caplet)
	cacheLock = sync.Mutex{}
)

func List() []*Caplet {
	caplets := make([]*Caplet, 0)
	for _, searchPath := range LoadPaths {
		files, _ := filepath.Glob(searchPath + "/*" + Suffix)
		files2, _ := filepath.Glob(searchPath + "/*/*" + Suffix)

		for _, fileName := range append(files, files2...) {
			if _, err := os.Stat(fileName); err == nil {
				base := strings.Replace(fileName, searchPath+string(os.PathSeparator), "", -1)
				base = strings.Replace(base, Suffix, "", -1)

				if caplet, err := Load(base); err != nil {
					fmt.Fprintf(os.Stderr, "wtf: %v\n", err)
				} else {
					caplets = append(caplets, caplet)
				}
			}
		}
	}

	sort.Slice(caplets, func(i, j int) bool {
		return strings.Compare(caplets[i].Name, caplets[j].Name) == -1
	})

	return caplets
}

func Load(name string) (*Caplet, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if caplet, found := cache[name]; found {
		return caplet, nil
	}

	baseName := name
	names := []string{}
	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	if !filepath.IsAbs(name) {
		for _, path := range LoadPaths {
			names = append(names, filepath.Join(path, name))
		}
	} else {
		names = append(names, name)
	}

	for _, fileName := range names {
		if stats, err := os.Stat(fileName); err == nil {
			cap := &Caplet{
				Script:  newScript(fileName, stats.Size()),
				Name:    baseName,
				Scripts: make([]Script, 0),
			}
			cache[name] = cap

			if reader, err := fs.LineReader(fileName); err != nil {
				return nil, fmt.Errorf("error reading caplet %s: %v", fileName, err)
			} else {
				for line := range reader {
					cap.Code = append(cap.Code, line)
				}

				// the caplet has a dedicated folder
				if strings.Contains(baseName, "/") || strings.Contains(baseName, "\\") {
					dir := filepath.Dir(fileName)
					// get all secondary .cap and .js files
					if files, err := ioutil.ReadDir(dir); err == nil && len(files) > 0 {
						for _, f := range files {
							subFileName := filepath.Join(dir, f.Name())
							if subFileName != fileName && (strings.HasSuffix(subFileName, ".cap") || strings.HasSuffix(subFileName, ".js")) {
								if reader, err := fs.LineReader(subFileName); err == nil {
									script := newScript(subFileName, f.Size())
									for line := range reader {
										script.Code = append(script.Code, line)
									}
									cap.Scripts = append(cap.Scripts, script)
								}
							}
						}
					}
				}
			}

			return cap, nil
		}
	}
	return nil, fmt.Errorf("caplet %s not found", name)
}
