package caplets

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/caplets"
	"github.com/bettercap/bettercap/session"
	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/islazy/zip"
)

type CapletsModule struct {
	session.SessionModule
}

func NewCapletsModule(s *session.Session) *CapletsModule {
	mod := &CapletsModule{
		SessionModule: session.NewSessionModule("caplets", s),
	}

	mod.AddHandler(session.NewModuleHandler("caplets.show", "",
		"Show a list of installed caplets.",
		func(args []string) error {
			return mod.Show()
		}))

	mod.AddHandler(session.NewModuleHandler("caplets.paths", "",
		"Show a list caplet search paths.",
		func(args []string) error {
			return mod.Paths()
		}))

	mod.AddHandler(session.NewModuleHandler("caplets.update", "",
		"Install/updates the caplets.",
		func(args []string) error {
			return mod.Update()
		}))

	mod.AddHandler(session.NewModuleHandler("caplets.fixpaths", "",
		"Fix the absolute caplets path.",
		func(args []string) error {
			re := regexp.MustCompile(`(~/|/usr/local/share/bettercap/)`)
			return filepath.Walk(caplets.InstallPath,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() || (filepath.Ext(path) != ".js" && filepath.Ext(path) != ".cap") {
						return nil
					}

					raw, err := ioutil.ReadFile(path)
					if err != nil {
						return err
					}

					res := re.ReplaceAllStringFunc(string(raw), func(value string) string {
						switch value {
						case "~/":
							return strings.Replace(caplets.UserHomePath, `\`, `/`, -1) + "/"
						case "/usr/local/share/bettercap/":
							return strings.Replace(caplets.InstallBase, `\`, `/`, -1) + "/"
						default:
							return value
						}
					})
					err = ioutil.WriteFile(path, []byte(res), info.Mode())
					if err != nil {
						return err
					}
					return nil
				})
		}))

	return mod
}

func (mod *CapletsModule) Name() string {
	return "caplets"
}

func (mod *CapletsModule) Description() string {
	return "A module to list and update caplets."
}

func (mod *CapletsModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *CapletsModule) Configure() error {
	return nil
}

func (mod *CapletsModule) Stop() error {
	return nil
}

func (mod *CapletsModule) Start() error {
	return nil
}

func (mod *CapletsModule) Show() error {
	caplets := caplets.List()
	if len(caplets) == 0 {
		return fmt.Errorf("no installed caplets on this system, use the caplets.update command to download them")
	}

	colNames := []string{
		"Name",
		"Path",
		"Size",
	}
	rows := [][]string{}

	for _, caplet := range caplets {
		rows = append(rows, []string{
			tui.Bold(caplet.Name),
			caplet.Path,
			tui.Dim(humanize.Bytes(uint64(caplet.Size))),
		})
	}

	tui.Table(os.Stdout, colNames, rows)

	return nil
}

func (mod *CapletsModule) Paths() error {
	colNames := []string{
		"Path",
	}
	rows := [][]string{}

	for _, path := range caplets.LoadPaths {
		rows = append(rows, []string{path})
	}

	tui.Table(os.Stdout, colNames, rows)
	fmt.Printf("(paths can be customized by defining the %s environment variable)\n", tui.Bold(caplets.EnvVarName))

	return nil
}

func (mod *CapletsModule) Update() error {
	if !fs.Exists(caplets.InstallBase) {
		mod.Info("creating caplets install path %s ...", caplets.InstallBase)
		if err := os.MkdirAll(caplets.InstallBase, os.ModePerm); err != nil {
			return err
		}
	}

	out, err := os.Create(caplets.ArchivePath)
	if err != nil {
		return err
	}
	defer out.Close()

	mod.Info("downloading caplets from %s ...", caplets.InstallArchive)

	resp, err := http.Get(caplets.InstallArchive)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	mod.Info("installing caplets to %s ...", caplets.InstallPath)

	if _, err = zip.Unzip(caplets.ArchivePath, caplets.InstallBase); err != nil {
		return err
	}

	os.RemoveAll(caplets.InstallPath)

	return os.Rename(caplets.InstallPathArchive, caplets.InstallPath)
}
