package wifi

import (
	"bytes"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (mod *WiFiModule) stationPruner() {
	mod.reads.Add(1)
	defer mod.reads.Done()

	maxApTTL := time.Duration(mod.apTTL) * time.Second
	maxStaTTL := time.Duration(mod.staTTL) * time.Second

	mod.Debug("wifi stations pruner started (ap.ttl:%v sta.ttl:%v).", maxApTTL, maxStaTTL)
	for mod.Running() {
		// loop every AP
		for _, ap := range mod.Session.WiFi.List() {
			sinceLastSeen := time.Since(ap.LastSeen)
			if sinceLastSeen > maxApTTL {
				mod.Debug("station %s not seen in %s, removing.", ap.BSSID(), sinceLastSeen)
				mod.Session.WiFi.Remove(ap.BSSID())
				continue
			}
			// loop every AP client
			for _, c := range ap.Clients() {
				sinceLastSeen := time.Since(c.LastSeen)
				if sinceLastSeen > maxStaTTL {
					mod.Debug("client %s of station %s not seen in %s, removing.", c.String(), ap.BSSID(), sinceLastSeen)
					ap.RemoveClient(c.BSSID())

					mod.Session.Events.Add("wifi.client.lost", ClientEvent{
						AP:     ap,
						Client: c,
					})
				}
			}
		}
		time.Sleep(1 * time.Second)
		// refresh
		maxApTTL = time.Duration(mod.apTTL) * time.Second
		maxStaTTL = time.Duration(mod.staTTL) * time.Second
	}
}

func (mod *WiFiModule) discoverAccessPoints(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	// search for Dot11InformationElementIDSSID
	if ok, ssid := packets.Dot11ParseIDSSID(packet); ok {
		from := dot11.Address3

		// skip stuff we're sending
		if mod.apRunning && bytes.Equal(from, mod.apConfig.BSSID) {
			return
		}

		if !network.IsZeroMac(from) && !network.IsBroadcastMac(from) {
			if int(radiotap.DBMAntennaSignal) >= mod.minRSSI {
				var frequency int
				bssid := from.String()

				if found, channel := packets.Dot11ParseDSSet(packet); found {
					frequency = network.Dot11Chan2Freq(channel)
				} else {
					frequency = int(radiotap.ChannelFrequency)
				}

				if ap, isNew := mod.Session.WiFi.AddIfNew(ssid, bssid, frequency, radiotap.DBMAntennaSignal); !isNew {
					//set beacon packet on the access point station.
					//This is for it to be included in the saved handshake file for wifi.assoc
					ap.Station.Handshake.Beacon = packet
					ap.EachClient(func(mac string, station *network.Station) {
						station.Handshake.SetBeacon(packet)
					})
				}
			} else {
				mod.Debug("skipping %s with %d dBm", from.String(), radiotap.DBMAntennaSignal)
			}
		}
	}
}

func (mod *WiFiModule) discoverProbes(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
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

	clientSTA := network.NormalizeMac(dot11.Address2.String())
	if mod.filterProbeSTA != nil && !mod.filterProbeSTA.MatchString(clientSTA) {
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

	apSSID := string(req.Contents[2 : 2+size])
	if mod.filterProbeAP != nil && !mod.filterProbeAP.MatchString(apSSID) {
		return
	}

	mod.Session.Events.Add("wifi.client.probe", ProbeEvent{
		FromAddr:   clientSTA,
		FromVendor: network.ManufLookup(clientSTA),
		FromAlias:  mod.Session.Lan.GetAlias(clientSTA),
		SSID:       apSSID,
		RSSI:       radiotap.DBMAntennaSignal,
	})
}

func (mod *WiFiModule) discoverClients(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	mod.Session.WiFi.EachAccessPoint(func(bssid string, ap *network.AccessPoint) {
		// packet going to this specific BSSID?
		if packets.Dot11IsDataFor(dot11, ap.HW) {
			bssid := dot11.Address2.String()
			freq := int(radiotap.ChannelFrequency)
			rssi := radiotap.DBMAntennaSignal

			if station, isNew := ap.AddClientIfNew(bssid, freq, rssi); isNew {
				mod.Session.Events.Add("wifi.client.new", ClientEvent{
					AP:     ap,
					Client: station,
				})
			}
		}
	})
}
