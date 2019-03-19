package caplets

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evilsocket/islazy/fs"
)

type Caplet struct {
	Name string   `json:"name"`
	Path string   `json:"path"`
	Size int64    `json:"size"`
	Code []string `json:"code"`
}

func NewCaplet(name string, path string, size int64) Caplet {
	return Caplet{
		Name: name,
		Path: path,
		Size: size,
		Code: make([]string, 0),
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
