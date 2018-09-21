package caplets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Caplet struct {
	Name string
	Path string
	Size int64
	Code []string
}

func (cap *Caplet) Eval(argv []string, lineCb func(line string) error) error {
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
			fmt.Printf("error while restoring working directory: %v\n", err)
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

		if err = lineCb(line); err != nil {
			return err
		}
	}

	return nil
}
