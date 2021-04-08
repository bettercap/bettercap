package net_probe

import (
	"sync"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/malfunkt/iprange"
)

type Probes struct {
	NBNS bool
	MDNS bool
	UPNP bool
	WSD  bool
}

type Prober struct {
	session.SessionModule
	throttle  int
	probes    Probes
	waitGroup *sync.WaitGroup
}

func NewProber(s *session.Session) *Prober {
	mod := &Prober{
		SessionModule: session.NewSessionModule("net.probe", s),
		waitGroup:     &sync.WaitGroup{},
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddParam(session.NewBoolParameter("net.probe.nbns",
		"true",
		"Enable NetBIOS name service discovery probes."))

	mod.AddParam(session.NewBoolParameter("net.probe.mdns",
		"true",
		"Enable mDNS discovery probes."))

	mod.AddParam(session.NewBoolParameter("net.probe.upnp",
		"true",
		"Enable UPNP discovery probes."))

	mod.AddParam(session.NewBoolParameter("net.probe.wsd",
		"true",
		"Enable WSD discovery probes."))

	mod.AddParam(session.NewIntParameter("net.probe.throttle",
		"10",
		"If greater than 0, probe packets will be throttled by this value in milliseconds."))

	mod.AddHandler(session.NewModuleHandler("net.probe on", "",
		"Start network hosts probing in background.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("net.probe off", "",
		"Stop network hosts probing in background.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod Prober) Name() string {
	return "net.probe"
}

func (mod Prober) Description() string {
	return "Keep probing for new hosts on the network by sending dummy UDP packets to every possible IP on the subnet."
}

func (mod Prober) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *Prober) Configure() error {
	var err error
	if err, mod.throttle = mod.IntParam("net.probe.throttle"); err != nil {
		return err
	} else if err, mod.probes.NBNS = mod.BoolParam("net.probe.nbns"); err != nil {
		return err
	} else if err, mod.probes.MDNS = mod.BoolParam("net.probe.mdns"); err != nil {
		return err
	} else if err, mod.probes.UPNP = mod.BoolParam("net.probe.upnp"); err != nil {
		return err
	} else if err, mod.probes.WSD = mod.BoolParam("net.probe.wsd"); err != nil {
		return err
	} else {
		mod.Debug("Throttling packets of %d ms.", mod.throttle)
	}
	return nil
}

func (mod *Prober) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		if mod.Session.Interface.IpAddress == network.MonitorModeAddress {
			mod.Info("Interface is in monitor mode, skipping net.probe")
			return
		}

		cidr := mod.Session.Interface.CIDR()
		list, err := iprange.Parse(cidr)
		if err != nil {
			mod.Fatal("%s", err)
		}

		if mod.probes.MDNS {
			go mod.mdnsProber()
		}

		fromIP := mod.Session.Interface.IP
		fromHW := mod.Session.Interface.HW
		addresses := list.Expand()
		throttle := time.Duration(mod.throttle) * time.Millisecond

		mod.Info("probing %d addresses on %s", len(addresses), cidr)

		for mod.Running() {
			if mod.probes.MDNS {
				mod.sendProbeMDNS(fromIP, fromHW)
			}

			if mod.probes.UPNP {
				mod.sendProbeUPNP(fromIP, fromHW)
			}

			if mod.probes.WSD {
				mod.sendProbeWSD(fromIP, fromHW)
			}

			for _, ip := range addresses {
				if !mod.Running() {
					return
				} else if mod.Session.Skip(ip) {
					mod.Debug("skipping address %s from probing.", ip)
					continue
				} else if mod.probes.NBNS {
					mod.sendProbeNBNS(fromIP, fromHW, ip)
				}
				time.Sleep(throttle)
			}

			time.Sleep(5 * time.Second)
		}
	})
}

func (mod *Prober) Stop() error {
	return mod.SetRunning(false, func() {
		mod.waitGroup.Wait()
	})
}
