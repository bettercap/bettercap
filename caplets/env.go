package caplets

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/evilsocket/islazy/str"
	"github.com/mitchellh/go-homedir"
)

const (
	EnvVarName     = "CAPSPATH"
	Suffix         = ".cap"
	InstallArchive = "https://github.com/bettercap/caplets/archive/master.zip"
)

func getInstallBase() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ALLUSERSPROFILE"), "bettercap")
	}
	return "/usr/local/share/bettercap/"
}

func getUserHomeDir() string {
	usr, _ := homedir.Dir()
	return usr
}

var (
	UserHomePath       = getUserHomeDir()
	InstallBase        = getInstallBase()
	InstallPathArchive = filepath.Join(InstallBase, "caplets-master")
	InstallPath        = filepath.Join(InstallBase, "caplets")
	ArchivePath        = filepath.Join(os.TempDir(), "caplets.zip")

	LoadPaths = []string{
		"./",
		"./caplets/",
		InstallPath,
		filepath.Join(UserHomePath, "caplets"),
	}
)

func init() {
	for _, path := range str.SplitBy(str.Trim(os.Getenv(EnvVarName)), string(os.PathListSeparator)) {
		if path = str.Trim(path); len(path) > 0 {
			LoadPaths = append(LoadPaths, path)
		}
	}

	for i, path := range LoadPaths {
		LoadPaths[i], _ = filepath.Abs(path)
	}
}
