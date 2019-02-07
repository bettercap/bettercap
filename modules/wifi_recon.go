package modules

import (
	"bytes"
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/evilsocket/islazy/tui"
)

var maxStationTTL = 5 * time.Minute

func (w *WiFiModule) stationPruner() {
	w.reads.Add(1)
	defer w.reads.Done()

	log.Debug("wifi stations pruner started.")
	for w.Running() {
		// loop every AP
		for _, ap := range w.Session.WiFi.List() {
			sinceLastSeen := time.Since(ap.LastSeen)
			if sinceLastSeen > maxStationTTL {
				log.Debug("station %s not seen in %s, removing.", ap.BSSID(), sinceLastSeen)
				w.Session.WiFi.Remove(ap.BSSID())
				continue
			}
			// loop every AP client
			for _, c := range ap.Clients() {
				sinceLastSeen := time.Since(c.LastSeen)
				if sinceLastSeen > maxStationTTL {
					log.Debug("client %s of station %s not seen in %s, removing.", c.String(), ap.BSSID(), sinceLastSeen)
					ap.RemoveClient(c.BSSID())

					w.Session.Events.Add("wifi.client.lost", WiFiClientEvent{
						AP:     ap,
						Client: c,
					})
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *WiFiModule) discoverAccessPoints(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	// search for Dot11InformationElementIDSSID
	if ok, ssid := packets.Dot11ParseIDSSID(packet); ok {
		from := dot11.Address3

		// skip stuff we're sending
		if w.apRunning && bytes.Equal(from, w.apConfig.BSSID) {
			return
		}

		if !network.IsZeroMac(from) && !network.IsBroadcastMac(from) {
			var frequency int
			bssid := from.String()

			if found, channel := packets.Dot11ParseDSSet(packet); found {
				frequency = network.Dot11Chan2Freq(channel)
			} else {
				frequency = int(radiotap.ChannelFrequency)
			}

			if ap, isNew := w.Session.WiFi.AddIfNew(ssid, bssid, frequency, radiotap.DBMAntennaSignal); !isNew {
				ap.EachClient(func(mac string, station *network.Station) {
					station.Handshake.SetBeacon(packet)
				})
			}
		}
	}
}

func (w *WiFiModule) discoverProbes(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	if dot11.Type != layers.Dot11TypeMgmtProbeReq {
		return
	}

	reqLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeReq)
	if reqLayer == nil {
		return
	}

	req, ok := reqLayer.(*layers.Dot11MgmtProbeReq)
	if !ok {
		return
	}

	tot := len(req.Contents)
	if tot < 3 {
		return
	}

	avail := uint32(tot - 2)
	if avail < 2 {
		return
	}
	size := uint32(req.Contents[1])
	if size == 0 || size > avail {
		return
	}

	w.Session.Events.Add("wifi.client.probe", WiFiProbeEvent{
		FromAddr:   dot11.Address2,
		FromVendor: network.ManufLookup(dot11.Address2.String()),
		FromAlias:  w.Session.Lan.GetAlias(dot11.Address2.String()),
		SSID:       string(req.Contents[2 : 2+size]),
		RSSI:       radiotap.DBMAntennaSignal,
	})
}

func (w *WiFiModule) discoverClients(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	w.Session.WiFi.EachAccessPoint(func(bssid string, ap *network.AccessPoint) {
		// packet going to this specific BSSID?
		if packets.Dot11IsDataFor(dot11, ap.HW) {
			bssid := dot11.Address2.String()
			freq := int(radiotap.ChannelFrequency)
			rssi := radiotap.DBMAntennaSignal

			if station, isNew := ap.AddClientIfNew(bssid, freq, rssi); isNew {
				w.Session.Events.Add("wifi.client.new", WiFiClientEvent{
					AP:     ap,
					Client: station,
				})
			}
		}
	})
}

func allZeros(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

func (w *WiFiModule) discoverHandshakes(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	if keyLayer := packet.Layer(layers.LayerTypeEAPOLKey); keyLayer != nil {
		if key := keyLayer.(*layers.EAPOLKey); key.KeyType == layers.EAPOLKeyTypePairwise {
			staMac := net.HardwareAddr{}
			apMac := net.HardwareAddr{}
			if dot11.Flags.FromDS() {
				staMac = dot11.Address1
				apMac = dot11.Address2
			} else if dot11.Flags.ToDS() {
				staMac = dot11.Address2
				apMac = dot11.Address1
			}

			if station, found := w.Session.WiFi.GetClient(staMac.String()); found {
				// ref. https://hashcat.net/forum/thread-7717.html
				rawPMKID := []byte(nil)

				// ref. https://wlan1nde.wordpress.com/2014/10/27/4-way-handshake/
				if !key.Install && key.KeyACK && !key.KeyMIC {
					// [1] (ACK) AP is sending ANonce to the client
					log.Debug("[%s] got frame 1/4 of the %s <-> %s handshake (anonce:%x)",
						tui.Green("wifi"),
						apMac,
						staMac,
						key.Nonce)
					rawPMKID = station.Handshake.AddAndGetPMKID(packet)
				} else if !key.Install && !key.KeyACK && key.KeyMIC && !allZeros(key.Nonce) {
					// [2] (MIC) client is sending SNonce+MIC to the API
					log.Debug("[%s] got frame 2/4 of the %s <-> %s handshake (snonce:%x mic:%x)",
						tui.Green("wifi"),
						apMac,
						staMac,
						key.Nonce,
						key.MIC)
					station.Handshake.AddFrame(1, packet)
				} else if key.Install && key.KeyACK && key.KeyMIC {
					// [3]: (INSTALL+ACK+MIC) AP informs the client that the PTK is installed
					log.Debug("[%s] got frame 3/4 of the %s <-> %s handshake (mic:%x)",
						tui.Green("wifi"),
						apMac,
						staMac,
						key.MIC)
					station.Handshake.AddFrame(2, packet)
				}

				numUnsaved := station.Handshake.NumUnsaved()
				doSave := numUnsaved > 0
				if doSave && w.shakesFile != "" {
					log.Debug("saving handshake frames to %s", w.shakesFile)
					if err := w.Session.WiFi.SaveHandshakesTo(w.shakesFile, w.handle.LinkType()); err != nil {
						log.Error("error while saving handshake frames to %s: %s", w.shakesFile, err)
					}
				}

				if doSave && (rawPMKID != nil || station.Handshake.Complete()) {
					w.Session.Events.Add("wifi.client.handshake", WiFiHandshakeEvent{
						File:       w.shakesFile,
						NewPackets: numUnsaved,
						AP:         apMac,
						Station:    staMac,
						PMKID:      rawPMKID,
					})
				}
			} else {
				log.Warning("EAPOL captured for unknown station %s", staMac.String())
			}
		}
	}
}
