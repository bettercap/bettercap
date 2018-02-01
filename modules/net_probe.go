package modules

import (
	"net"
	"sync"
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/malfunkt/iprange"
)

type Prober struct {
	session.SessionModule
	throttle int
}

func NewProber(s *session.Session) *Prober {
	p := &Prober{
		SessionModule: session.NewSessionModule("net.probe", s),
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

func (p *Prober) shouldProbe(ip net.IP) bool {
	addr := ip.String()
	if ip.IsLoopback() == true {
		return false
	} else if addr == p.Session.Interface.IpAddress {
		return false
	} else if addr == p.Session.Gateway.IpAddress {
		return false
	}
	return true
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
	if p.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := p.Configure(); err != nil {
		return err
	}

	p.SetRunning(true)

	go func() {
		list, err := iprange.Parse(p.Session.Interface.CIDR())
		if err != nil {
			log.Fatal("%s", err)
		}

		from := p.Session.Interface.IP
		from_hw := p.Session.Interface.HW
		addresses := list.Expand()

		for p.Running() {
			for _, ip := range addresses {
				if p.shouldProbe(ip) == false {
					log.Debug("Skipping address %s from UDP probing.", ip)
					continue
				}

				p.sendProbe(from, from_hw, ip)

				if p.throttle > 0 {
					time.Sleep(time.Duration(p.throttle) * time.Millisecond)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (p *Prober) Stop() error {
	if p.Running() == false {
		return session.ErrAlreadyStopped
	}
	p.SetRunning(false)
	return nil
}
