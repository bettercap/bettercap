package wifi

import (
	"fmt"
	"net"
	"sort"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/packets"
)

func (mod *WiFiModule) sendAssocPacket(ap network.StationSnapshot) {
	if err, pkt := packets.NewDot11Auth(mod.iface.HW, ap.HW, 1); err != nil {
		mod.Error("cloud not create auth packet: %s", err)
	} else {
		mod.injectPacket(pkt)
	}

	if err, pkt := packets.NewDot11AssociationRequest(mod.iface.HW, ap.HW, ap.Hostname, 1); err != nil {
		mod.Error("cloud not create association request packet: %s", err)
	} else {
		mod.injectPacket(pkt)
	}
}

func (mod *WiFiModule) skipAssoc(to net.HardwareAddr) bool {
	return hardwareAddrIn(mod.assocSkip, to)
}

func (mod *WiFiModule) isAssocSilent() bool {
	if err, is := mod.BoolParam("wifi.assoc.silent"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.assocSilent = is
	}
	return mod.assocSilent
}

func (mod *WiFiModule) doAssocOpen() bool {
	if err, is := mod.BoolParam("wifi.assoc.open"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.assocOpen = is
	}
	return mod.assocOpen
}

func (mod *WiFiModule) doAssocAcquired() bool {
	if err, is := mod.BoolParam("wifi.assoc.acquired"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.assocAcquired = is
	}
	return mod.assocAcquired
}

func (mod *WiFiModule) assocTargetsFor(selector wifiSelector) []wifiTarget {
	toAssoc := make([]wifiTarget, 0)
	for _, target := range mod.resolveWiFiTargets(selector, false) {
		if !mod.skipAssoc(target.apSnapshot.HW) {
			toAssoc = append(toAssoc, target)
		} else {
			mod.Debug("skipping ap:%v because skip list %v", target.ap, mod.assocSkip)
		}
	}
	return toAssoc
}

func (mod *WiFiModule) assocCompleter(prefix string) []string {
	return mod.wifiTargetCompleter(prefix, false)
}

func (mod *WiFiModule) startAssoc(target string) error {
	selector, err := newWiFiSelector(target)
	if err != nil {
		return err
	}

	// parse skip list
	if assocSkip, err := mod.parseWiFiSkipList("wifi.assoc.skip"); err != nil {
		return err
	} else {
		mod.assocSkip = assocSkip
	}

	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !mod.Running() {
		if err := mod.Configure(); err != nil {
			return err
		}
		defer mod.handle.Close()
	}

	toAssoc := mod.assocTargetsFor(selector)

	if len(toAssoc) == 0 {
		if selector.all {
			return nil
		}
		return fmt.Errorf("%q is an unknown BSSID or ESSID, or it is in the association skip list", selector.raw)
	}
	mod.writes.Add(1)
	go func() {
		defer mod.writes.Done()

		// since we need to change the wifi adapter channel for each
		// association request, let's sort by channel so we do the minimum
		// amount of hops possible
		sort.Slice(toAssoc, func(i, j int) bool {
			return toAssoc[i].apSnapshot.Channel < toAssoc[j].apSnapshot.Channel
		})

		// send the association request frames
		for _, target := range toAssoc {
			if mod.Running() {
				ap := target.ap
				snapshot := target.apSnapshot
				logger := mod.Info
				if mod.isAssocSilent() {
					logger = mod.Debug
				}

				if (snapshot.Encryption == "" || snapshot.Encryption == "OPEN") && !mod.doAssocOpen() {
					mod.Debug("skipping association for open network %s (wifi.assoc.open is false)", snapshot.Hostname)
				} else if ap.HasKeyMaterial() && !mod.doAssocAcquired() {
					mod.Debug("skipping association for AP %s (key material already acquired)", snapshot.Hostname)
				} else {
					logger("sending association request to AP %s (channel:%d encryption:%s)", snapshot.Hostname, snapshot.Channel, snapshot.Encryption)

					mod.onChannel(snapshot.Channel, func() {
						mod.sendAssocPacket(snapshot)
					})
				}
			}
		}
	}()

	return nil
}
