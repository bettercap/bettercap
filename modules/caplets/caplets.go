package caplets

import (
	"fmt"
	"io"
	"net/http"
	"os"

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
	c := &CapletsModule{
		SessionModule: session.NewSessionModule("caplets", s),
	}

	c.AddHandler(session.NewModuleHandler("caplets.show", "",
		"Show a list of installed caplets.",
		func(args []string) error {
			return c.Show()
		}))

	c.AddHandler(session.NewModuleHandler("caplets.paths", "",
		"Show a list caplet search paths.",
		func(args []string) error {
			return c.Paths()
		}))

	c.AddHandler(session.NewModuleHandler("caplets.update", "",
		"Install/updates the caplets.",
		func(args []string) error {
			return c.Update()
		}))

	return c
}

func (c *CapletsModule) Name() string {
	return "caplets"
}

func (c *CapletsModule) Description() string {
	return "A module to list and update caplets."
}

func (c *CapletsModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (c *CapletsModule) Configure() error {
	return nil
}

func (c *CapletsModule) Stop() error {
	return nil
}

func (c *CapletsModule) Start() error {
	return nil
}

func (c *CapletsModule) Show() error {
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

func (c *CapletsModule) Paths() error {
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

func (c *CapletsModule) Update() error {
	if !fs.Exists(caplets.InstallBase) {
		c.Info("creating caplets install path %s ...", caplets.InstallBase)
		if err := os.MkdirAll(caplets.InstallBase, os.ModePerm); err != nil {
			return err
		}
	}

	out, err := os.Create("/tmp/caplets.zip")
	if err != nil {
		return err
	}
	defer out.Close()

	c.Info("downloading caplets from %s ...", caplets.InstallArchive)

	resp, err := http.Get(caplets.InstallArchive)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	c.Info("installing caplets to %s ...", caplets.InstallPath)

	if _, err = zip.Unzip("/tmp/caplets.zip", caplets.InstallBase); err != nil {
		return err
	}

	os.RemoveAll(caplets.InstallPath)

	return os.Rename(caplets.InstallPathArchive, caplets.InstallPath)
}
