package update

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/session"

	"github.com/google/go-github/github"

	"github.com/evilsocket/islazy/tui"
)

type UpdateModule struct {
	session.SessionModule
	client *github.Client
}

func NewUpdateModule(s *session.Session) *UpdateModule {
	mod := &UpdateModule{
		SessionModule: session.NewSessionModule("update", s),
		client:        github.NewClient(nil),
	}

	mod.AddHandler(session.NewModuleHandler("update.check on", "",
		"Check latest available stable version and compare it with the one being used.",
		func(args []string) error {
			return mod.Start()
		}))

	return mod
}

func (mod *UpdateModule) Name() string {
	return "update"
}

func (mod *UpdateModule) Description() string {
	return "A module to check for bettercap's updates."
}

func (mod *UpdateModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *UpdateModule) Configure() error {
	return nil
}

func (mod *UpdateModule) Stop() error {
	return nil
}

func (mod *UpdateModule) versionToNum(ver string) float64 {
	if ver[0] == 'v' {
		ver = ver[1:]
	}

	n := 0.0
	parts := strings.Split(ver, ".")
	nparts := len(parts)

	// reverse
	for i := nparts/2 - 1; i >= 0; i-- {
		opp := nparts - 1 - i
		parts[i], parts[opp] = parts[opp], parts[i]
	}

	for i, e := range parts {
		ev, _ := strconv.Atoi(e)
		n += float64(ev) * math.Pow10(i)
	}

	return n
}

func (mod *UpdateModule) Start() error {
	return mod.SetRunning(true, func() {
		defer mod.SetRunning(false, nil)

		mod.Info("checking latest stable release ...")

		if releases, _, err := mod.client.Repositories.ListReleases(context.Background(), "bettercap", "bettercap", nil); err == nil {
			latest := releases[0]
			if mod.versionToNum(core.Version) < mod.versionToNum(*latest.TagName) {
				mod.Session.Events.Add("update.available", latest)
			} else {
				mod.Info("you are running %s which is the latest stable version.", tui.Bold(core.Version))
			}
		} else {
			mod.Error("error while fetching latest release info from GitHub: %s", err)
		}
	})
}
