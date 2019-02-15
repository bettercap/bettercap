package wifi

import (
	"bytes"

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
	if ok, key, apMac, staMac := packets.Dot11ParseEAPOL(packet, dot11); ok {
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
		staIsUs := bytes.Equal(staMac, mod.Session.Interface.HW)
		station, found := ap.Get(staMac.String())
		if !found {
			station, _ = ap.AddClientIfNew(staMac.String(), ap.Frequency, ap.RSSI)
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
		doSave := numUnsaved > 0
		if doSave && mod.shakesFile != "" {
			mod.Debug("saving handshake frames to %s", mod.shakesFile)
			if err := mod.Session.WiFi.SaveHandshakesTo(mod.shakesFile, mod.handle.LinkType()); err != nil {
				mod.Error("error while saving handshake frames to %s: %s", mod.shakesFile, err)
			}
		}

		// if we had unsaved packets and either the handshake is complete
		// or it contains the PMKID, generate a new event.
		if doSave && (rawPMKID != nil || station.Handshake.Complete()) {
			mod.Session.Events.Add("wifi.client.handshake", HandshakeEvent{
				File:       mod.shakesFile,
				NewPackets: numUnsaved,
				AP:         apMac,
				Station:    staMac,
				PMKID:      rawPMKID,
			})
			// make sure the info that we have key material for this AP
			// is persisted even after stations are pruned due to inactivity
			ap.WithKeyMaterial(true)
		}

		// if we added ourselves as a client station but we didn't get any
		// PMKID, just remove it from the list of clients of this AP.
		if staIsUs && rawPMKID == nil {
			ap.RemoveClient(staMac.String())
		}
	}
}
