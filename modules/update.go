package modules

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/google/go-github/github"
)

type UpdateModule struct {
	session.SessionModule
	client *github.Client
}

func NewUpdateModule(s *session.Session) *UpdateModule {
	u := &UpdateModule{
		SessionModule: session.NewSessionModule("update", s),
		client:        github.NewClient(nil),
	}

	u.AddHandler(session.NewModuleHandler("update.check on", "",
		"Check latest available stable version and compare it with the one being used.",
		func(args []string) error {
			return u.Start()
		}))

	return u
}

func (u *UpdateModule) Name() string {
	return "update"
}

func (u *UpdateModule) Description() string {
	return "A module to check for bettercap's updates."
}

func (u *UpdateModule) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (u *UpdateModule) Configure() error {
	return nil
}

func (u *UpdateModule) Stop() error {
	return nil
}

func (u *UpdateModule) versionToNum(ver string) float64 {
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

func (u *UpdateModule) Start() error {
	return u.SetRunning(true, func() {
		defer u.SetRunning(false, nil)

		log.Info("Checking latest stable release ...")

		if releases, _, err := u.client.Repositories.ListReleases(context.Background(), "bettercap", "bettercap", nil); err == nil {
			latest := releases[0]
			if u.versionToNum(core.Version) < u.versionToNum(*latest.TagName) {
				u.Session.Events.Add("update.available", latest)
			} else {
				log.Info("You are running %s which is the latest stable version.", core.Bold(core.Version))
			}
		} else {
			log.Error("Error while fetching latest release info from GitHub: %s", err)
		}
	})
}
