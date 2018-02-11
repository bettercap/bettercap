package modules

import (
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/session"
)

type Discovery struct {
	session.SessionModule
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

	d.AddHandler(session.NewModuleHandler("net.show", "",
		"Show cache hosts list (default sorting by ip).",
		func(args []string) error {
			return d.Show("address")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by seen", "",
		"Show cache hosts list (sort by last seen).",
		func(args []string) error {
			return d.Show("seen")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by sent", "",
		"Show cache hosts list (sort by sent packets).",
		func(args []string) error {
			return d.Show("sent")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by rcvd", "",
		"Show cache hosts list (sort by received packets).",
		func(args []string) error {
			return d.Show("rcvd")
		}))

	return d
}

func (d Discovery) Name() string {
	return "net.recon"
}

func (d Discovery) Description() string {
	return "Read periodically the ARP cache in order to monitor for new hosts on the network."
}

func (d Discovery) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (d *Discovery) runDiff(cache net.ArpTable) {
	// check for endpoints who disappeared
	var rem net.ArpTable = make(net.ArpTable)
	for mac, t := range d.Session.Targets.Targets {
		if _, found := cache[mac]; found == false {
			rem[mac] = t.IpAddress
		}
	}

	for mac, ip := range rem {
		d.Session.Targets.Remove(ip, mac)
	}

	// now check for new friends ^_^
	for ip, mac := range cache {
		d.Session.Targets.AddIfNew(ip, mac)
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
			if table, err := net.ArpUpdate(iface); err != nil {
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
