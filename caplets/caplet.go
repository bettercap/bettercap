package caplets

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evilsocket/islazy/fs"
)

type Caplet struct {
	Name string
	Path string
	Size int64
	Code []string
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
