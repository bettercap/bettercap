package caplets

import (
	"os"
	"path/filepath"

	"github.com/evilsocket/islazy/str"
)

const (
	EnvVarName     = "CAPSPATH"
	Suffix         = ".cap"
	InstallArchive = "https://github.com/bettercap/caplets/archive/master.zip"
	InstallBase    = "/usr/local/share/bettercap/"
)

var (
	InstallPathArchive = filepath.Join(InstallBase, "caplets-master")
	InstallPath        = filepath.Join(InstallBase, "caplets")

	LoadPaths = []string{
		"./",
		"./caplets/",
		InstallPath,
	}
)

func init() {
	for _, path := range str.SplitBy(str.Trim(os.Getenv(EnvVarName)), ":") {
		if path = str.Trim(path); len(path) > 0 {
			LoadPaths = append(LoadPaths, path)
		}
	}

	for i, path := range LoadPaths {
		LoadPaths[i], _ = filepath.Abs(path)
	}
}
