package prober

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
	p := &Prober{
		SessionModule: session.NewSessionModule("net.probe", s),
		waitGroup:     &sync.WaitGroup{},
	}

	p.AddParam(session.NewBoolParameter("net.probe.nbns",
		"true",
		"Enable NetBIOS name service discovery probes."))

	p.AddParam(session.NewBoolParameter("net.probe.mdns",
		"true",
		"Enable mDNS discovery probes."))

	p.AddParam(session.NewBoolParameter("net.probe.upnp",
		"true",
		"Enable UPNP discovery probes."))

	p.AddParam(session.NewBoolParameter("net.probe.wsd",
		"true",
		"Enable WSD discovery probes."))

	p.AddParam(session.NewIntParameter("net.probe.throttle",
		"10",
		"If greater than 0, probe packets will be throttled by this value in milliseconds."))

	p.AddHandler(session.NewModuleHandler("net.probe on", "",
		"Start network hosts probing in background.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("net.probe off", "",
		"Stop network hosts probing in background.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p Prober) Name() string {
	return "net.probe"
}

func (p Prober) Description() string {
	return "Keep probing for new hosts on the network by sending dummy UDP packets to every possible IP on the subnet."
}

func (p Prober) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (p *Prober) Configure() error {
	var err error
	if err, p.throttle = p.IntParam("net.probe.throttle"); err != nil {
		return err
	} else if err, p.probes.NBNS = p.BoolParam("net.probe.nbns"); err != nil {
		return err
	} else if err, p.probes.MDNS = p.BoolParam("net.probe.mdns"); err != nil {
		return err
	} else if err, p.probes.UPNP = p.BoolParam("net.probe.upnp"); err != nil {
		return err
	} else if err, p.probes.WSD = p.BoolParam("net.probe.wsd"); err != nil {
		return err
	} else {
		p.Debug("Throttling packets of %d ms.", p.throttle)
	}
	return nil
}

func (p *Prober) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		if p.Session.Interface.IpAddress == network.MonitorModeAddress {
			p.Info("Interface is in monitor mode, skipping net.probe")
			return
		}

		list, err := iprange.Parse(p.Session.Interface.CIDR())
		if err != nil {
			p.Fatal("%s", err)
		}

		from := p.Session.Interface.IP
		from_hw := p.Session.Interface.HW
		addresses := list.Expand()
		throttle := time.Duration(p.throttle) * time.Millisecond

		for p.Running() {
			if p.probes.MDNS {
				p.sendProbeMDNS(from, from_hw)
			}

			if p.probes.UPNP {
				p.sendProbeUPNP(from, from_hw)
			}

			if p.probes.WSD {
				p.sendProbeWSD(from, from_hw)
			}

			for _, ip := range addresses {
				if !p.Running() {
					return
				} else if p.Session.Skip(ip) {
					p.Debug("skipping address %s from probing.", ip)
					continue
				} else if p.probes.NBNS {
					p.sendProbeNBNS(from, from_hw, ip)
				}
				time.Sleep(throttle)
			}

			time.Sleep(5 * time.Second)
		}
	})
}

func (p *Prober) Stop() error {
	return p.SetRunning(false, func() {
		p.waitGroup.Wait()
	})
}
