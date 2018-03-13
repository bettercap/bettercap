package modules

import (
	"crypto/rand"
	"fmt"
	"net"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	openFlags = 1057
	wpaFlags  = 1041
	//1-54 Mbit
	supportedRates = []byte{0x82, 0x84, 0x8b, 0x96, 0x24, 0x30, 0x48, 0x6c, 0x03, 0x01}
	wpaRSN         = []byte{
		0x01, 0x00, // RSN Version 1
		0x00, 0x0f, 0xac, 0x02, // Group Cipher Suite : 00-0f-ac TKIP
		0x02, 0x00, // 2 Pairwise Cipher Suites (next two lines)
		0x00, 0x0f, 0xac, 0x04, // AES Cipher / CCMP
		0x00, 0x0f, 0xac, 0x02, // TKIP Cipher
		0x01, 0x00, // 1 Authentication Key Managment Suite (line below)
		0x00, 0x0f, 0xac, 0x02, // Pre-Shared Key
		0x00, 0x00,
	}
)

type Dot11BeaconConfig struct {
	SSID       string
	BSSID      net.HardwareAddr
	Channel    int
	Encryption bool
}

func NewDot11Beacon(conf Dot11BeaconConfig) (error, []byte) {
	flags := openFlags
	if conf.Encryption == true {
		flags = wpaFlags
	}

	stack := []gopacket.SerializableLayer{
		&layers.RadioTap{},
		&layers.Dot11{
			Address1: network.BroadcastHw,
			Address2: conf.BSSID,
			Address3: conf.BSSID,
			Type:     layers.Dot11TypeMgmtBeacon,
		},
		&layers.Dot11MgmtBeacon{
			Flags:    uint16(flags),
			Interval: 100,
		},
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDSSID,
			Length: uint8(len(conf.SSID) & 0xff),
			Info:   []byte(conf.SSID),
		},
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDRates,
			Length: uint8(len(supportedRates) & 0xff),
			Info:   supportedRates,
		},
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDDSSet,
			Length: 1,
			Info:   []byte{byte(conf.Channel & 0xff)},
		},
	}

	if conf.Encryption == true {
		stack = append(stack, &layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDRSNInfo,
			Length: uint8(len(wpaRSN) & 0xff),
			Info:   wpaRSN,
		})
	}

	return packets.Serialize(stack...)
}

func (w *WiFiModule) sendBeaconPacket(counter int) {
	w.writes.Add(1)
	defer w.writes.Done()

	hw := make([]byte, 6)
	rand.Read(hw)

	n := counter % len(w.frequencies)

	conf := Dot11BeaconConfig{
		SSID:       fmt.Sprintf("Prova_%d", n),
		BSSID:      w.Session.Interface.HW,
		Channel:    network.Dot11Freq2Chan(w.frequencies[n]),
		Encryption: true,
	}

	if err, pkt := NewDot11Beacon(conf); err != nil {
		log.Error("Could not create beacon packet: %s", err)
	} else {
		w.injectPacket(pkt)
	}

	time.Sleep(100 * time.Millisecond)
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
