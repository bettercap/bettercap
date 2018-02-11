package modules

import (
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/session"
)

type Discovery struct {
	session.SessionModule

	refresh int
	before  net.ArpTable
	current net.ArpTable
	quit    chan bool
}

func NewDiscovery(s *session.Session) *Discovery {
	d := &Discovery{
		SessionModule: session.NewSessionModule("net.recon", s),

		refresh: 1,
		before:  nil,
		current: nil,
		quit:    make(chan bool),
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
		"Show current hosts list (default sorting by ip).",
		func(args []string) error {
			return d.Show("address")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by seen", "",
		"Show current hosts list (sort by last seen).",
		func(args []string) error {
			return d.Show("seen")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by sent", "",
		"Show current hosts list (sort by sent packets).",
		func(args []string) error {
			return d.Show("sent")
		}))

	d.AddHandler(session.NewModuleHandler("net.show by rcvd", "",
		"Show current hosts list (sort by received packets).",
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

func (d *Discovery) runDiff() {
	// check for endpoints who disappeared
	var rem net.ArpTable = make(net.ArpTable)
	for mac, t := range d.Session.Targets.Targets {
		if _, found := d.current[mac]; found == false {
			rem[mac] = t.IpAddress
		}
	}

	for mac, ip := range rem {
		d.Session.Targets.Remove(ip, mac)
	}

	// now check for new friends ^_^
	for ip, mac := range d.current {
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
		for {
			select {
			case <-time.After(time.Duration(d.refresh) * time.Second):
				var err error

				if d.current, err = net.ArpUpdate(d.Session.Interface.Name()); err != nil {
					log.Error("%s", err)
					continue
				}

				d.runDiff()

				d.before = d.current

			case <-d.quit:
				return
			}
		}
	})
}

func (d *Discovery) Stop() error {
	return d.SetRunning(false, func() {
		d.quit <- true
	})
}
