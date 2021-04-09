package wifi

import (
	"bytes"
	"fmt"
	"github.com/bettercap/bettercap/network"
	"net"
	"path"

	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func allZeros(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

func (mod *WiFiModule) discoverHandshakes(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	isEAPOL := false

	if ok, key, apMac, staMac := packets.Dot11ParseEAPOL(packet, dot11); ok {
		isEAPOL = true

		// first, locate the AP in our list by its BSSID
		ap, found := mod.Session.WiFi.Get(apMac.String())
		if !found {
			mod.Warning("could not find AP with BSSID %s", apMac.String())
			return
		}

		// locate the client station, if its BSSID is ours, it means we sent
		// an association request via wifi.assoc because we're trying to capture
		// the PMKID from the first EAPOL sent by the AP.
		// (Reference about PMKID https://hashcat.net/forum/thread-7717.html)
		// In this case, we need to add ourselves as a client station of the AP
		// in order to have a consistent association of AP, client and handshakes.
		staIsUs := bytes.Equal(staMac, mod.iface.HW)
		station, found := ap.Get(staMac.String())
		staAdded := false
		if !found {
			station, staAdded = ap.AddClientIfNew(staMac.String(), ap.Frequency, ap.RSSI)
		}

		rawPMKID := []byte(nil)
		if !key.Install && key.KeyACK && !key.KeyMIC {
			// [1] (ACK) AP is sending ANonce to the client
			rawPMKID = station.Handshake.AddAndGetPMKID(packet)
			PMKID := "without PMKID"
			if rawPMKID != nil {
				PMKID = "with PMKID"
			}

			mod.Debug("got frame 1/4 of the %s <-> %s handshake (%s) (anonce:%x)",
				apMac,
				staMac,
				PMKID,
				key.Nonce)

			//add the ap's station's beacon packet to be saved as part of the handshake cap file
			//https://github.com/ZerBea/hcxtools/issues/92
			//https://github.com/bettercap/bettercap/issues/592

			if ap.Station.Handshake.Beacon != nil {
				mod.Debug("adding beacon frame to handshake for %s", apMac)
				station.Handshake.AddFrame(1, ap.Station.Handshake.Beacon)
			}

		} else if !key.Install && !key.KeyACK && key.KeyMIC && !allZeros(key.Nonce) {
			// [2] (MIC) client is sending SNonce+MIC to the API
			station.Handshake.AddFrame(1, packet)

			mod.Debug("got frame 2/4 of the %s <-> %s handshake (snonce:%x mic:%x)",
				apMac,
				staMac,
				key.Nonce,
				key.MIC)
		} else if key.Install && key.KeyACK && key.KeyMIC {
			// [3]: (INSTALL+ACK+MIC) AP informs the client that the PTK is installed
			station.Handshake.AddFrame(2, packet)

			mod.Debug("got frame 3/4 of the %s <-> %s handshake (mic:%x)",
				apMac,
				staMac,
				key.MIC)
		}

		// if we have unsaved packets as part of the handshake, save them.
		numUnsaved := station.Handshake.NumUnsaved()
		shakesFileName := mod.shakesFile
		if mod.shakesAggregate == false {
			shakesFileName = path.Join(shakesFileName, fmt.Sprintf("%s.pcap", ap.PathFriendlyName()))
		}
		doSave := numUnsaved > 0
		if doSave && shakesFileName != "" {
			mod.Debug("(aggregate %v) saving handshake frames to %s", mod.shakesAggregate, shakesFileName)
			if err := mod.Session.WiFi.SaveHandshakesTo(shakesFileName, mod.handle.LinkType()); err != nil {
				mod.Error("error while saving handshake frames to %s: %s", shakesFileName, err)
			}
		}

		validPMKID := rawPMKID != nil
		validHalfHandshake := !staIsUs && station.Handshake.Half()
		validFullHandshake := station.Handshake.Complete()
		// if we have unsaved packets AND
		//   if we captured a PMKID OR
		//   if we captured am half handshake which is not ours OR
		//   if we captured a full handshake
		if doSave && (validPMKID || validHalfHandshake || validFullHandshake) {
			mod.Session.Events.Add("wifi.client.handshake", HandshakeEvent{
				File:       shakesFileName,
				NewPackets: numUnsaved,
				AP:         apMac.String(),
				Station:    staMac.String(),
				PMKID:      rawPMKID,
				Half:       station.Handshake.Half(),
				Full:       station.Handshake.Complete(),
			})
			// make sure the info that we have key material for this AP
			// is persisted even after stations are pruned due to inactivity
			ap.WithKeyMaterial(true)
		}
		// if we added ourselves as a client station but we didn't get any
		// PMKID, just remove it from the list of clients of this AP.
		if staAdded || (staIsUs && rawPMKID == nil) {
			ap.RemoveClient(staMac.String())
		}
	}

	// quick and dirty heuristic, see thread here https://github.com/bettercap/bettercap/issues/810#issuecomment-805145392
	if isEAPOL || (dot11.Type.MainType() != layers.Dot11TypeData && dot11.Type.MainType() != layers.Dot11TypeCtrl) {
		target := (* network.Station)(nil)
		targetAP := (* network.AccessPoint)(nil)

		// collect target bssids
		bssids := make([]net.HardwareAddr, 0)
		for _, addr := range []net.HardwareAddr{dot11.Address1, dot11.Address2, dot11.Address3, dot11.Address4} {
			if bytes.Equal(addr, network.BroadcastHw) == false {
				bssids = append(bssids, addr)
			}
		}

		// for each AP
		mod.Session.WiFi.EachAccessPoint(func(mac string, ap *network.AccessPoint) {
			// only check APs we captured handshakes of
			if target == nil && ap.HasKeyMaterial() {
				// search client station
				ap.EachClient(func(mac string, station *network.Station) {
					// any valid key material for this station?
					if station.Handshake.Any() {
						// check if target
						for _, a := range bssids {
							if bytes.Equal(a, station.HW) {
								target = station
								targetAP = ap
								break
							}
						}
					}
				})
			}
		})

		if target != nil {
			mod.Debug("saving extra %s frame (%d bytes) for %s",
				dot11.Type.String(),
				len(packet.Data()),
				target.String())

			target.Handshake.AddExtra(packet)

			shakesFileName := mod.shakesFile
			if mod.shakesAggregate == false {
				shakesFileName = path.Join(shakesFileName, fmt.Sprintf("%s.pcap", targetAP.PathFriendlyName()))
			}
			if shakesFileName != "" {
				mod.Debug("(aggregate %v) saving handshake frames to %s", mod.shakesAggregate, shakesFileName)
				if err := mod.Session.WiFi.SaveHandshakesTo(shakesFileName, mod.handle.LinkType()); err != nil {
					mod.Error("error while saving handshake frames to %s: %s", shakesFileName, err)
				}
			}
		}
	}

}
