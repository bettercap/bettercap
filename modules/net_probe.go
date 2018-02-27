package modules

import (
	"net"
	"sync"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/malfunkt/iprange"
)

type Prober struct {
	session.SessionModule
	throttle  int
	waitGroup *sync.WaitGroup
}

func NewProber(s *session.Session) *Prober {
	p := &Prober{
		SessionModule: session.NewSessionModule("net.probe", s),
		waitGroup:     &sync.WaitGroup{},
	}

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
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *Prober) sendProbe(from net.IP, from_hw net.HardwareAddr, ip net.IP) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		p.sendProbeUDP(from, from_hw, ip)
		w.Done()
	}(&wg)

	wg.Wait()
}

func (p *Prober) Configure() error {
	var err error
	if err, p.throttle = p.IntParam("net.probe.throttle"); err != nil {
		return err
	} else {
		log.Debug("Throttling packets of %d ms.", p.throttle)
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
			log.Info("Interface is in monitor mode, skipping net.probe")
			return
		}

		list, err := iprange.Parse(p.Session.Interface.CIDR())
		if err != nil {
			log.Fatal("%s", err)
		}

		from := p.Session.Interface.IP
		from_hw := p.Session.Interface.HW
		addresses := list.Expand()
		throttle := time.Duration(p.throttle) * time.Millisecond

		for p.Running() {
			for _, ip := range addresses {
				if p.Running() == false {
					return
				} else if p.Session.Skip(ip) == true {
					log.Debug("Skipping address %s from UDP probing.", ip)
					continue
				}

				p.sendProbe(from, from_hw, ip)

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
