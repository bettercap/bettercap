package modules

import (
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket/layers"
)

func NewDot11Beacon(bssid net.HardwareAddr, ssid string, seq uint16) (error, []byte) {
	// TODO: still very incomplete
	return packets.Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       network.BroadcastHw,
			Address2:       bssid,
			Address3:       bssid,
			Type:           layers.Dot11TypeMgmtBeacon,
			SequenceNumber: seq, // not sure this needs to be a specific value
		},
		&layers.Dot11MgmtBeacon{
			Timestamp: uint64(time.Now().Second()), // not sure
			Interval:  1041,                        // ?
			Flags:     100,                         // ?
		},
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDSSID,
			Length: uint8(len(ssid) & 0xff),
			Info:   []byte(ssid),
		},
		// TODO: Rates n stuff ...
		&layers.Dot11InformationElement{
			BaseLayer: layers.BaseLayer{
				Contents: []byte{0x01, 0x08, 0x82, 0x84, 0x8b, 0x96, 0x24, 0x30, 0x48, 0x6c},
			},
		},
		&layers.Dot11InformationElement{
			BaseLayer: layers.BaseLayer{
				Contents: []byte{0x03, 0x01, 0x0b},
			},
		},
	)
}

func (w *WiFiModule) sendBeaconPacket(counter int) {
	w.writes.Add(1)
	defer w.writes.Done()

	if err, pkt := NewDot11Beacon(w.Session.Interface.HW, "Prova", uint16(counter)); err != nil {
		log.Error("Could not create beacon packet: %s", err)
	} else {
		w.injectPacket(pkt)
	}

	time.Sleep(10 * time.Millisecond)
}

func (w *WiFiModule) startBeaconFlood() error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if w.Running() == false {
		if err := w.Configure(); err != nil {
			return err
		}
	}

	go func() {
		defer w.handle.Close()
		for counter := 0; w.Running(); counter++ {
			w.sendBeaconPacket(counter)
		}
	}()

	return nil
}
