package net_probe

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/bettercap/bettercap/packets"

	"github.com/hashicorp/mdns"
)

var services = []string{
	"_hap._tcp.local",
	"_homekit._tcp.local",
	"_airplay._tcp.local",
	"_raop._tcp.local",
	"_sleep-proxy._udp.local",
	"_companion-link._tcp.local",
	"_googlezone._tcp.local",
	"_googlerpc._tcp.local",
	"_googlecast._tcp.local",
	"local",
}

func (mod *Prober) sendProbeMDNS(from net.IP, from_hw net.HardwareAddr) {
	err, raw := packets.NewMDNSProbe(from, from_hw)
	if err != nil {
		mod.Error("error while sending mdns probe: %v", err)
		return
	} else if err := mod.Session.Queue.Send(raw); err != nil {
		mod.Error("error sending mdns packet: %s", err)
	} else {
		mod.Debug("sent %d bytes of MDNS probe", len(raw))
	}
}

func (mod *Prober) mdnsListener(c chan *mdns.ServiceEntry) {
	mod.Debug("mdns listener started")
	defer mod.Debug("mdns listener stopped")

	for entry := range c {
		addrs := []string{}
		if entry.AddrV4 != nil {
			addrs = append(addrs, entry.AddrV4.String())
		}
		if entry.AddrV6 != nil {
			addrs = append(addrs, entry.AddrV6.String())
		}

		for _, addr := range addrs {
			if host := mod.Session.Lan.GetByIp(addr); host != nil {
				meta := make(map[string]string)

				meta["mdns:name"] = entry.Name
				meta["mdns:hostname"] = entry.Host

				if entry.AddrV4 != nil {
					meta["mdns:ipv4"] = entry.AddrV4.String()
				}

				if entry.AddrV6 != nil {
					meta["mdns:ipv6"] = entry.AddrV6.String()
				}

				meta["mdns:port"] = fmt.Sprintf("%d", entry.Port)

				mod.Debug("meta for %s: %v", addr, meta)

				host.OnMeta(meta)
			} else {
				mod.Debug("got mdns entry for unknown ip %s", entry.AddrV4)
			}
		}
	}
}

func (mod *Prober) mdnsProber() {
	mod.Debug("mdns prober started")
	defer mod.Debug("mdns.prober stopped")

	mod.waitGroup.Add(1)
	defer mod.waitGroup.Done()

	log.SetOutput(ioutil.Discard)

	ch := make(chan *mdns.ServiceEntry)
	defer close(ch)

	go mod.mdnsListener(ch)

	for mod.Running() {
		for _, svc := range services {
			if mod.Running() {
				mdns.Lookup(svc, ch)
			}
		}
	}
}
