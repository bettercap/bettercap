package net_probe

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

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

func (mod *Prober) mdnsProber() {
	mod.waitGroup.Add(1)
	defer mod.waitGroup.Done()

	mod.Debug("mdns prober started")
	defer mod.Debug("mdns.prober stopped")

	log.SetOutput(ioutil.Discard)

	ch := make(chan *mdns.ServiceEntry)
	wait := sync.WaitGroup{}

	defer close(ch)

	go func(c chan *mdns.ServiceEntry) {
		mod.Debug("mdns channel read started")
		defer mod.Debug("mdns channel read stopped")

		for entry := range c {
			if host := mod.Session.Lan.GetByIp(entry.AddrV4.String()); host != nil {
				meta := make(map[string]string)

				meta["mdns:name"] = entry.Name
				meta["mdns:hostname"] = entry.Host
				meta["mdns:ipv4"] = entry.AddrV4.String()

				if entry.AddrV6 != nil {
					meta["mdns:ipv6"] = entry.AddrV6.String()
				}

				meta["mdns:port"] = fmt.Sprintf("%d", entry.Port)

				host.OnMeta(meta)
			} else {
				mod.Debug("got mdns entry for known ip %s", entry.AddrV4)
			}
		}
	}(ch)

	for mod.Running() {
		for _, svc := range services {
			go func(svc string, w *sync.WaitGroup) {
				w.Add(1)
				defer w.Done()

				params := mdns.DefaultParams(svc)
				params.Entries = ch
				params.Timeout = time.Duration(5) * time.Second

				mdns.Query(params)
			}(svc, &wait)
		}
		wait.Wait()
	}
}
