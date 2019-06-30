package packets

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"reflect"
	"testing"
)

func TestDot11Vars(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{openFlags, 1057},
		{wpaFlags, 1041},
		{fakeApRates, []byte{0x82, 0x84, 0x8b, 0x96, 0x24, 0x30, 0x48, 0x6c, 0x03, 0x01}},
		{fakeApWpaRSN, []byte{
			0x01, 0x00, // RSN Version 1
			0x00, 0x0f, 0xac, 0x02, // Group Cipher Suite : 00-0f-ac TKIP
			0x02, 0x00, // 2 Pairwise Cipher Suites (next two lines)
			0x00, 0x0f, 0xac, 0x04, // AES Cipher / CCMP
			0x00, 0x0f, 0xac, 0x02, // TKIP Cipher
			0x01, 0x00, // 1 Authentication Key Management Suite (line below)
			0x00, 0x0f, 0xac, 0x02, // Pre-Shared Key
			0x00, 0x00,
		}},
		{wpaSignatureBytes, []byte{0, 0x50, 0xf2, 1}},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func BuildDot11ApConfig() Dot11ApConfig {
	ssid := "I still love Ruby, don't worry!"
	bssid, _ := net.ParseMAC("pi:ca:tw:as:he:re")
	channel := 1
	encryption := false

	config := Dot11ApConfig{
		SSID:       ssid,
		BSSID:      bssid,
		Channel:    channel,
		Encryption: encryption,
	}

	return config
}

func TestDot11ApConfig(t *testing.T) {
	ssid := "I still love Ruby, don't worry!"
	bssid, _ := net.ParseMAC("pi:ca:tw:as:he:re")
	channel := 1
	encryption := false

	config := Dot11ApConfig{
		SSID:       ssid,
		BSSID:      bssid,
		Channel:    channel,
		Encryption: encryption,
	}

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{config.SSID, ssid},
		{config.BSSID, bssid},
		{config.Channel, channel},
		{config.Encryption, encryption},
	}

	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11Info(t *testing.T) {
	id := layers.Dot11InformationElementIDSSID
	info := []byte{}

	dot11InfoElement := Dot11Info(id, info)

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{dot11InfoElement.ID, id},
		{dot11InfoElement.Length, uint8(len(info))},
		{dot11InfoElement.Info, info},
	}

	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestNewDot11Beacon(t *testing.T) {
	conf := BuildDot11ApConfig()
	seq := uint16(0)

	err, bytes := NewDot11Beacon(conf, seq)

	if err != nil {
		t.Error(err)
	}

	if len(bytes) <= 0 {
		t.Error("unable to create new dot11 beacon")
	}
}

func TestNewDot11Deauth(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:00:00:00:00")
	seq := uint16(0)

	err, bytes := NewDot11Deauth(mac, mac, mac, seq)

	if err != nil {
		t.Error(err)
	}

	if len(bytes) <= 0 {
		t.Error("unable to create new dot11 beacon")
	}
}

func BuildDot11Packet() gopacket.Packet {
	mac, _ := net.ParseMAC("00:00:00:00:00:00")
	seq := uint16(0)
	_, bytes := Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       mac,
			Address2:       mac,
			Address3:       mac,
			Type:           layers.Dot11TypeMgmtDeauthentication,
			SequenceNumber: seq,
		},
		&layers.Dot11MgmtDeauthentication{
			Reason: layers.Dot11ReasonClass2FromNonAuth,
		},
	)

	return gopacket.NewPacket(bytes, layers.LayerTypeRadioTap, gopacket.Default)
}

func TestDot11Parse(t *testing.T) {
	packet := BuildDot11Packet()

	ok, radiotap, dot11 := Dot11Parse(packet)

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		// testing for the known bad cases
		{ok, false},
		{radiotap, nil},
		{dot11, nil},
	}

	for _, u := range units {
		if reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11ParseIDSSID(t *testing.T) {
	conf := BuildDot11ApConfig()
	seq := uint16(0)

	err, bytes := NewDot11Beacon(conf, seq)

	if err != nil {
		t.Error(err)
	}

	packet := gopacket.NewPacket(bytes, layers.LayerTypeRadioTap, gopacket.Default)

	ok, id := Dot11ParseIDSSID(packet)

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{ok, true},
		{id, "I still love Ruby, don't worry!"},
	}

	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11ParseEncryption(t *testing.T) {
	ssid := "I still love Ruby, don't worry!"
	bssid, _ := net.ParseMAC("pi:ca:tw:as:he:re")
	channel := 1
	encryption := true

	config := Dot11ApConfig{
		SSID:       ssid,
		BSSID:      bssid,
		Channel:    channel,
		Encryption: encryption,
	}

	seq := uint16(0)

	err, bytes := NewDot11Beacon(config, seq)

	if err != nil {
		t.Error(err)
	}

	packet := gopacket.NewPacket(bytes, layers.LayerTypeRadioTap, gopacket.Default)
	_, _, dot11 := Dot11Parse(packet)

	found, enc, cipher, auth := Dot11ParseEncryption(packet, dot11)

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{found, true},
		{enc, "WPA2"},
		{cipher, "TKIP"},
		{auth, "PSK"},
	}

	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestDot11IsDataFor(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:00:00:00:00")
	seq := uint16(0)
	_, bytes := Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       mac,
			Address2:       mac,
			Address3:       mac,
			Type:           layers.Dot11TypeData,
			SequenceNumber: seq,
		},
	)
	packet := gopacket.NewPacket(bytes, layers.LayerTypeRadioTap, gopacket.Default)
	station, _ := net.ParseMAC("00:00:00:00:00:00")
	_, _, dot11 := Dot11Parse(packet)
	if !Dot11IsDataFor(dot11, station) {
		t.Error("unable to determine dot11 packet is for a given station")
	}
}

// TODO: add Dot11ParseDSSet test. Not sure how to build proper
// example packet to complete this test, for now. <3
//func TestDot11ParseDSSet(t *testing.T) {
//}
