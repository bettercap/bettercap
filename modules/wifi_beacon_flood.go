package modules

import (
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket/layers"
)

type Dot11EncryptionType int

const (
	Dot11Open Dot11EncryptionType = iota
	Dot11Wep
	Dot11WpaTKIP
	Dot11WpaAES
)

type Dot11BeaconConfig struct {
	SSID       string
	BSSID      net.HardwareAddr
	Channel    int
	Encryption Dot11EncryptionType
}

func NewDot11Beacon(conf Dot11BeaconConfig) (error, []byte) {
	// TODO: still very incomplete
	return packets.Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       network.BroadcastHw,
			Address2:       conf.BSSID,
			Address3:       conf.BSSID,
			Type:           layers.Dot11TypeMgmtBeacon,
			SequenceNumber: 0, // not sure this needs to be a specific value
		},
		&layers.Dot11MgmtBeacon{
			Timestamp: uint64(time.Now().Second()), // not sure
			Interval:  1041,                        // ?
			Flags:     100,                         // ?
		},
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDSSID,
			Length: uint8(len(conf.SSID) & 0xff),
			Info:   []byte(conf.SSID),
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

	conf := Dot11BeaconConfig{
		SSID:       "Prova",
		BSSID:      w.Session.Interface.HW,
		Channel:    1,
		Encryption: Dot11Open,
	}

	if err, pkt := NewDot11Beacon(conf); err != nil {
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
