package caplets

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evilsocket/islazy/fs"
)

type Script struct {
	Path string   `json:"path"`
	Size int64    `json:"size"`
	Code []string `json:"code"`
}

func newScript(path string, size int64) Script {
	return Script{
		Path: path,
		Size: size,
		Code: make([]string, 0),
	}
}

type Caplet struct {
	Script
	Name    string   `json:"name"`
	Scripts []Script `json:"scripts"`
}

func NewCaplet(name string, path string, size int64) Caplet {
	return Caplet{
		Script:  newScript(path, size),
		Name:    name,
		Scripts: make([]Script, 0),
	}
}

func (cap *Caplet) Eval(argv []string, lineCb func(line string) error) error {
	if argv == nil {
		argv = []string{}
	}
	// the caplet might include other files (include directive, proxy modules, etc),
	// temporarily change the working directory
	return fs.Chdir(filepath.Dir(cap.Path), func() error {
		for _, line := range cap.Code {
			// skip empty lines and comments
			if line == "" || line[0] == '#' {
				continue
			}
			// replace $0 with argv[0], $1 with argv[1] and so on
			for i, arg := range argv {
				what := fmt.Sprintf("$%d", i)
				line = strings.Replace(line, what, arg, -1)
			}

			if err := lineCb(line); err != nil {
				return err
			}
		}
		return nil
	})
}
