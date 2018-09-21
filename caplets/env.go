package caplets

import (
	"os"
	"path/filepath"

	"github.com/bettercap/bettercap/core"
)

const (
	Suffix      = ".cap"
	InstallPath = "/usr/local/share/bettercap/caplets/"
)

var (
	LoadPaths = []string{
		"./caplets/",
		InstallPath,
	}
)

func init() {
	for _, path := range core.SepSplit(core.Trim(os.Getenv("CAPSPATH")), ":") {
		if path = core.Trim(path); len(path) > 0 {
			LoadPaths = append(LoadPaths, path)
		}
	}

	for i, path := range LoadPaths {
		LoadPaths[i], _ = filepath.Abs(path)
	}
}
