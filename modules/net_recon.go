package modules

import (
	"fmt"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/session"
	"time"
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
		"Show current hosts list.",
		func(args []string) error {
			return d.Show()
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

func (d Discovery) OnSessionEnded(s *session.Session) {
	if d.Running() {
		d.Stop()
	}
}

func (d *Discovery) checkShared(new net.ArpTable) {
	n_gw_shared := 0
	for ip, mac := range new {
		if ip != d.Session.Gateway.IpAddress && mac == d.Session.Gateway.HwAddress {
			n_gw_shared++
		}
	}

	if n_gw_shared > 0 {
		a := ""
		b := ""
		if n_gw_shared == 1 {
			a = ""
			b = "s"
		} else {
			a = "s"
			b = ""
		}

		log.Warning("Found %d endpoint%s which share%s the same MAC of the gateway (%s), there're might be some IP isolation going on, skipping.", n_gw_shared, a, b, d.Session.Gateway.HwAddress)
	}
}

func (d *Discovery) runDiff() {
	var new net.ArpTable = make(net.ArpTable)
	var rem net.ArpTable = make(net.ArpTable)

	if d.before != nil {
		new = net.ArpDiff(d.current, d.before)
		rem = net.ArpDiff(d.before, d.current)
	} else {
		new = d.current
	}

	if len(new) > 0 || len(rem) > 0 {
		d.checkShared(new)

		// refresh target pool
		for ip, mac := range rem {
			d.Session.Targets.Remove(ip, mac)
		}

		for ip, mac := range new {
			d.Session.Targets.AddIfNotExist(ip, mac)
		}
	}
}

func (d *Discovery) Start() error {
	if d.Running() == false {
		d.SetRunning(true)

		go func() {
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
		}()

		return nil
	} else {
		return fmt.Errorf("Network discovery already started.")
	}
}

func (d *Discovery) Show() error {
	d.Session.Targets.Dump()
	return nil
}

func (d *Discovery) Stop() error {
	if d.Running() == true {
		d.SetRunning(false)
		d.quit <- true
		return nil
	} else {
		return fmt.Errorf("Network discovery already stopped.")
	}
}
