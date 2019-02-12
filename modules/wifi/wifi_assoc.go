package wifi

import (
	"bytes"
	"fmt"
	"net"
	"sort"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
)

func (w *WiFiModule) sendAssocPacket(ap *network.AccessPoint) {
	if err, pkt := packets.NewDot11Auth(w.Session.Interface.HW, ap.HW, 1); err != nil {
		w.Error("cloud not create auth packet: %s", err)
	} else {
		w.injectPacket(pkt)
	}

	if err, pkt := packets.NewDot11AssociationRequest(w.Session.Interface.HW, ap.HW, ap.ESSID(), 1); err != nil {
		w.Error("cloud not create association request packet: %s", err)
	} else {
		w.injectPacket(pkt)
	}
}

func (w *WiFiModule) skipAssoc(to net.HardwareAddr) bool {
	for _, mac := range w.assocSkip {
		if bytes.Equal(to, mac) {
			return true
		}
	}
	return false
}

func (w *WiFiModule) isAssocSilent() bool {
	if err, is := w.BoolParam("wifi.assoc.silent"); err != nil {
		w.Warning("%v", err)
	} else {
		w.assocSilent = is
	}
	return w.assocSilent
}

func (w *WiFiModule) doAssocOpen() bool {
	if err, is := w.BoolParam("wifi.assoc.open"); err != nil {
		w.Warning("%v", err)
	} else {
		w.assocOpen = is
	}
	return w.assocOpen
}

func (w *WiFiModule) startAssoc(to net.HardwareAddr) error {
	// parse skip list
	if err, assocSkip := w.StringParam("wifi.assoc.skip"); err != nil {
		return err
	} else if macs, err := network.ParseMACs(assocSkip); err != nil {
		return err
	} else {
		w.assocSkip = macs
	}

	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !w.Running() {
		if err := w.Configure(); err != nil {
			return err
		}
		defer w.handle.Close()
	}

	toAssoc := make([]*network.AccessPoint, 0)
	isBcast := network.IsBroadcastMac(to)
	for _, ap := range w.Session.WiFi.List() {
		if isBcast || bytes.Equal(ap.HW, to) {
			if !w.skipAssoc(ap.HW) {
				toAssoc = append(toAssoc, ap)
			} else {
				w.Debug("skipping ap:%v because skip list %v", ap, w.assocSkip)
			}
		}
	}

	if len(toAssoc) == 0 {
		if isBcast {
			return nil
		}
		return fmt.Errorf("%s is an unknown BSSID or it is in the association skip list.", to.String())
	}

	go func() {
		w.writes.Add(1)
		defer w.writes.Done()

		// since we need to change the wifi adapter channel for each
		// association request, let's sort by channel so we do the minimum
		// amount of hops possible
		sort.Slice(toAssoc, func(i, j int) bool {
			return toAssoc[i].Channel() < toAssoc[j].Channel()
		})

		// send the association request frames
		for _, ap := range toAssoc {
			if w.Running() {
				logger := w.Info
				if w.isAssocSilent() {
					logger = w.Debug
				}

				if ap.IsOpen() && !w.doAssocOpen() {
					w.Debug("skipping association for open network %s (wifi.assoc.open is false)", ap.ESSID())
				} else {
					logger("sending association request to AP %s (channel:%d encryption:%s)", ap.ESSID(), ap.Channel(), ap.Encryption)

					w.onChannel(ap.Channel(), func() {
						w.sendAssocPacket(ap)
					})
				}
			}
		}
	}()

	return nil
}
