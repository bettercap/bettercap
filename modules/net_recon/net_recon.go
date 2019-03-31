package net_recon

import (
	"github.com/bettercap/bettercap/modules/utils"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
)

type Discovery struct {
	session.SessionModule
	selector *utils.ViewSelector
}

func NewDiscovery(s *session.Session) *Discovery {
	mod := &Discovery{
		SessionModule: session.NewSessionModule("net.recon", s),
	}

	mod.AddHandler(session.NewModuleHandler("net.recon on", "",
		"Start network hosts discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("net.recon off", "",
		"Stop network hosts discovery.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("net.clear", "",
		"Clear all endpoints collected by the hosts discovery module.",
		func(args []string) error {
			mod.Session.Lan.Clear()
			return nil
		}))

	mod.AddParam(session.NewBoolParameter("net.show.meta",
		"false",
		"If true, the net.show command will show all metadata collected about each endpoint."))

	mod.AddHandler(session.NewModuleHandler("net.show", "",
		"Show cache hosts list (default sorting by ip).",
		func(args []string) error {
			return mod.Show("")
		}))

	mod.AddHandler(session.NewModuleHandler("net.show ADDRESS1, ADDRESS2", `net.show (.+)`,
		"Show information about a specific comma separated list of addresses (by IP or MAC).",
		func(args []string) error {
			return mod.Show(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("net.show.meta ADDRESS1, ADDRESS2", `net\.show\.meta (.+)`,
		"Show meta information about a specific comma separated list of addresses (by IP or MAC).",
		func(args []string) error {
			return mod.showMeta(args[0])
		}))

	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "net.show", []string{"ip", "mac", "seen", "sent", "rcvd"},
		"ip asc")

	return mod
}

func (mod Discovery) Name() string {
	return "net.recon"
}

func (mod Discovery) Description() string {
	return "Read periodically the ARP cache in order to monitor for new hosts on the network."
}

func (mod Discovery) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *Discovery) runDiff(cache network.ArpTable) {
	// check for endpoints who disappeared
	var rem network.ArpTable = make(network.ArpTable)

	mod.Session.Lan.EachHost(func(mac string, e *network.Endpoint) {
		if _, found := cache[mac]; !found {
			rem[mac] = e.IpAddress
		}
	})

	for mac, ip := range rem {
		mod.Session.Lan.Remove(ip, mac)
	}

	// now check for new friends ^_^
	for ip, mac := range cache {
		mod.Session.Lan.AddIfNew(ip, mac)
	}
}

func (mod *Discovery) Configure() error {
	return nil
}

func (mod *Discovery) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		every := time.Duration(1) * time.Second
		iface := mod.Session.Interface.Name()
		for mod.Running() {
			if table, err := network.ArpUpdate(iface); err != nil {
				mod.Error("%s", err)
			} else {
				mod.runDiff(table)
			}
			time.Sleep(every)
		}
	})
}

func (mod *Discovery) Stop() error {
	return mod.SetRunning(false, nil)
}
