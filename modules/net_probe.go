package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/malfunkt/iprange"
)

type Prober struct {
	session.SessionModule
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
	} else if p.Session.Targets.Has(addr) == true {
		return false
	}
	return true
}

func (p Prober) OnSessionEnded(s *session.Session) {
	if p.Running() {
		p.Stop()
	}
}

func (p *Prober) sendProbe(from net.IP, from_hw net.HardwareAddr, ip net.IP) {
	name := fmt.Sprintf("%s:137", ip)
	if addr, err := net.ResolveUDPAddr("udp", name); err != nil {
		log.Error("Could not resolve %s.", name)
	} else if con, err := net.DialUDP("udp", nil, addr); err != nil {
		log.Error("Could not dial %s.", name)
	} else {
		log.Debug("UDP connection to %s enstablished.", name)
		defer con.Close()
		con.Write([]byte{0xde, 0xad, 0xbe, 0xef})
	}
}

func (p *Prober) Start() error {
	if p.Running() == false {
		throttle := int(0)
		if err, v := p.Param("net.probe.throttle").Get(p.Session); err != nil {
			return err
		} else {
			throttle = v.(int)
			log.Debug("Throttling packets of %d ms.", throttle)
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

					if throttle > 0 {
						time.Sleep(time.Duration(throttle) * time.Millisecond)
					}
				}

				time.Sleep(5 * time.Second)
			}
		}()

		return nil
	} else {
		return fmt.Errorf("Network prober already started.")
	}
}

func (p *Prober) Stop() error {
	if p.Running() == true {
		p.SetRunning(false)
		return nil
	} else {
		return fmt.Errorf("Network prober already stopped.")
	}
}
