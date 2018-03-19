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
)

var maxStationTTL = 5 * time.Minute

type WiFiProbe struct {
	FromAddr   net.HardwareAddr
	FromVendor string
	FromAlias  string
	SSID       string
	RSSI       int8
}

func (w *WiFiModule) stationPruner() {
	w.reads.Add(1)
	defer w.reads.Done()

	log.Debug("WiFi stations pruner started.")
	for w.Running() == true {
		for _, s := range w.Session.WiFi.List() {
			sinceLastSeen := time.Since(s.LastSeen)
			if sinceLastSeen > maxStationTTL {
				log.Debug("Station %s not seen in %s, removing.", s.BSSID(), sinceLastSeen)
				w.Session.WiFi.Remove(s.BSSID())
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *WiFiModule) discoverAccessPoints(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	// search for Dot11InformationElementIDSSID
	if ok, ssid := packets.Dot11ParseIDSSID(packet); ok == true {
		from := dot11.Address3

		// skip stuff we're sending
		if w.apRunning && bytes.Compare(from, w.apConfig.BSSID) == 0 {
			return
		}

		if network.IsZeroMac(from) == false && network.IsBroadcastMac(from) == false {
			var frequency int
			bssid := from.String()

			if found, channel := packets.Dot11ParseDSSet(packet); found {
				frequency = network.Dot11Chan2Freq(channel)
			} else {
				frequency = int(radiotap.ChannelFrequency)
			}

			w.Session.WiFi.AddIfNew(ssid, bssid, frequency, radiotap.DBMAntennaSignal)
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
	if ok == false {
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

	w.Session.Events.Add("wifi.client.probe", WiFiProbe{
		FromAddr:   dot11.Address2,
		FromVendor: network.OuiLookup(dot11.Address2.String()),
		FromAlias:  w.Session.Lan.GetAlias(dot11.Address2.String()),
		SSID:       string(req.Contents[2 : 2+size]),
		RSSI:       radiotap.DBMAntennaSignal,
	})
}

func (w *WiFiModule) discoverClients(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	w.Session.WiFi.EachAccessPoint(func(bssid string, ap *network.AccessPoint) {
		// packet going to this specific BSSID?
		if packets.Dot11IsDataFor(dot11, ap.HW) == true {
			ap.AddClient(dot11.Address2.String(), int(radiotap.ChannelFrequency), radiotap.DBMAntennaSignal)
		}
	})
}
