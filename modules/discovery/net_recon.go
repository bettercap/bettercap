package discovery

import (
	"github.com/bettercap/bettercap/modules/utils"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
)

type Discovery struct {
	session.SessionModule
	selector *utils.ViewSelector
}

func NewDiscovery(s *session.Session) *Discovery {
	d := &Discovery{
		SessionModule: session.NewSessionModule("net.recon", s),
	}

	d.AddHandler(session.NewModuleHandler("net.recon on", "",
		"Start network hosts discovery.",
		func(args []string) error {
			return d.Start()
		}))

	d.AddHandler(session.NewModuleHandler("net.recon off", "",
		"Stop network hosts discovery.",
		func(args []string) error {
			return d.Stop()
		}))

	d.AddParam(session.NewBoolParameter("net.show.meta",
		"false",
		"If true, the net.show command will show all metadata collected about each endpoint."))

	d.AddHandler(session.NewModuleHandler("net.show", "",
		"Show cache hosts list (default sorting by ip).",
		func(args []string) error {
			return d.Show("")
		}))

	d.AddHandler(session.NewModuleHandler("net.show ADDRESS1, ADDRESS2", `net.show (.+)`,
		"Show information about a specific list of addresses (by IP or MAC).",
		func(args []string) error {
			return d.Show(args[0])
		}))

	d.AddHandler(session.NewModuleHandler("net.show.meta ADDRESS1, ADDRESS2", `net\.show\.meta (.+)`,
		"Show meta information about a specific list of addresses (by IP or MAC).",
		func(args []string) error {
			return d.showMeta(args[0])
		}))

	d.selector = utils.ViewSelectorFor(&d.SessionModule, "net.show", []string{"ip", "mac", "seen", "sent", "rcvd"},
		"ip asc")

	return d
}

func (d Discovery) Name() string {
	return "net.recon"
}

func (d Discovery) Description() string {
	return "Read periodically the ARP cache in order to monitor for new hosts on the network."
}

func (d Discovery) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (d *Discovery) runDiff(cache network.ArpTable) {
	// check for endpoints who disappeared
	var rem network.ArpTable = make(network.ArpTable)

	d.Session.Lan.EachHost(func(mac string, e *network.Endpoint) {
		if _, found := cache[mac]; !found {
			rem[mac] = e.IpAddress
		}
	})

	for mac, ip := range rem {
		d.Session.Lan.Remove(ip, mac)
	}

	// now check for new friends ^_^
	for ip, mac := range cache {
		d.Session.Lan.AddIfNew(ip, mac)
	}
}

func (d *Discovery) Configure() error {
	return nil
}

func (d *Discovery) Start() error {
	if err := d.Configure(); err != nil {
		return err
	}

	return d.SetRunning(true, func() {
		every := time.Duration(1) * time.Second
		iface := d.Session.Interface.Name()
		for d.Running() {
			if table, err := network.ArpUpdate(iface); err != nil {
				log.Error("%s", err)
			} else {
				d.runDiff(table)
			}
			time.Sleep(every)
		}
	})
}

func (d *Discovery) Stop() error {
	return d.SetRunning(false, nil)
}
