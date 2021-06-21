package ui

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/bettercap/bettercap/session"

	"github.com/google/go-github/github"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/islazy/zip"
)

var versionParser = regexp.MustCompile(`name:"ui",version:"([^"]+)"`)

type UIModule struct {
	session.SessionModule
	client   *github.Client
	tmpFile  string
	basePath string
	uiPath   string
}

func getDefaultInstallBase() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ALLUSERSPROFILE"), "bettercap")
	}
	return "/usr/local/share/bettercap/"
}

func NewUIModule(s *session.Session) *UIModule {
	mod := &UIModule{
		SessionModule: session.NewSessionModule("ui", s),
		client:        github.NewClient(nil),
	}

	mod.AddParam(session.NewStringParameter("ui.basepath",
		getDefaultInstallBase(),
		"",
		"UI base installation path."))

	mod.AddParam(session.NewStringParameter("ui.tmpfile",
		filepath.Join(os.TempDir(), "ui.zip"),
		"",
		"Temporary file to use while downloading UI updates."))

	mod.AddHandler(session.NewModuleHandler("ui.version", "",
		"Print the currently installed UI version.",
		func(args []string) error {
			return mod.showVersion()
		}))

	mod.AddHandler(session.NewModuleHandler("ui.update", "",
		"Download the latest available version of the UI and install it.",
		func(args []string) error {
			return mod.Start()
		}))

	return mod
}

func (mod *UIModule) Name() string {
	return "ui"
}

func (mod *UIModule) Description() string {
	return "A module to manage bettercap's UI updates and installed version."
}

func (mod *UIModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *UIModule) Configure() (err error) {
	if err, mod.basePath = mod.StringParam("ui.basepath"); err != nil {
		return err
	} else {
		mod.uiPath = filepath.Join(mod.basePath, "ui")
	}

	if err, mod.tmpFile = mod.StringParam("ui.tmpfile"); err != nil {
		return err
	}

	return nil
}

func (mod *UIModule) Stop() error {
	return nil
}

func (mod *UIModule) download(version, url string) error {
	if !fs.Exists(mod.basePath) {
		mod.Warning("creating ui install path %s ...", mod.basePath)
		if err := os.MkdirAll(mod.basePath, os.ModePerm); err != nil {
			return err
		}
	}

	out, err := os.Create(mod.tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(mod.tmpFile)

	mod.Info("downloading ui %s from %s ...", tui.Bold(version), url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if fs.Exists(mod.uiPath) {
		mod.Warning("removing previously installed UI from %s ...", mod.uiPath)
		if err := os.RemoveAll(mod.uiPath); err != nil {
			return err
		}
	}

	mod.Info("installing to %s ...", mod.uiPath)

	if _, err = zip.Unzip(mod.tmpFile, mod.basePath); err != nil {
		return err
	}

	mod.Info("installation complete, you can now run the %s (or https-ui) caplet to start the UI.", tui.Bold("http-ui"))

	return nil
}

func (mod *UIModule) showVersion() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	if !fs.Exists(mod.uiPath) {
		return fmt.Errorf("path %s does not exist, ui not installed", mod.uiPath)
	}

	search := filepath.Join(mod.uiPath, "/main.*.js")
	matches, err := filepath.Glob(search)
	if err != nil {
		return err
	} else if len(matches) == 0 {
		return fmt.Errorf("can't find any main.*.js files in %s", mod.uiPath)
	}

	for _, filename := range matches {
		if raw, err := ioutil.ReadFile(filename); err != nil {
			return err
		} else if m := versionParser.FindStringSubmatch(string(raw)); m != nil {
			version := m[1]
			mod.Info("v%s", version)
			return nil
		}
	}

	return fmt.Errorf("can't parse version from %s", search)
}

func (mod *UIModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	} else if err := mod.SetRunning(true, nil); err != nil {
		return err
	}
	defer mod.SetRunning(false, nil)

	mod.Info("checking latest stable release ...")

	if releases, _, err := mod.client.Repositories.ListReleases(context.Background(), "bettercap", "ui", nil); err == nil {
		latest := releases[0]
		for _, a := range latest.Assets {
			if *a.Name == "ui.zip" {
				return mod.download(*latest.TagName, *a.BrowserDownloadURL)
			}
		}
	} else {
		mod.Error("error while fetching latest release info from GitHub: %s", err)
	}

	return nil
}
