package packets

import (
	"bytes"
	"net"

	"github.com/bettercap/bettercap/network"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	openFlags      = 1057
	wpaFlags       = 1041
	durationID     = uint16(0x013a)
	capabilityInfo = uint16(0x0411)
	listenInterval = uint16(3)
	//1-54 Mbit
	fakeApRates  = []byte{0x82, 0x84, 0x8b, 0x96, 0x24, 0x30, 0x48, 0x6c, 0x03, 0x01}
	fakeApWpaRSN = []byte{
		0x01, 0x00, // RSN Version 1
		0x00, 0x0f, 0xac, 0x02, // Group Cipher Suite : 00-0f-ac TKIP
		0x02, 0x00, // 2 Pairwise Cipher Suites (next two lines)
		0x00, 0x0f, 0xac, 0x04, // AES Cipher / CCMP
		0x00, 0x0f, 0xac, 0x02, // TKIP Cipher
		0x01, 0x00, // 1 Authentication Key Management Suite (line below)
		0x00, 0x0f, 0xac, 0x02, // Pre-Shared Key
		0x00, 0x00,
	}
	wpaSignatureBytes = []byte{0, 0x50, 0xf2, 1}

	assocRates        = []byte{0x82, 0x84, 0x8b, 0x96, 0x24, 0x30, 0x48, 0x6c}
	assocESRates      = []byte{0x0C, 0x12, 0x18, 0x60}
	assocRSNInfo      = []byte{0x01, 0x00, 0x00, 0x0F, 0xAC, 0x04, 0x01, 0x00, 0x00, 0x0F, 0xAC, 0x04, 0x01, 0x00, 0x00, 0x0F, 0xAC, 0x02, 0x8C, 0x00}
	assocCapabilities = []byte{0x2C, 0x01, 0x03, 0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type Dot11ApConfig struct {
	SSID       string
	BSSID      net.HardwareAddr
	Channel    int
	Encryption bool
}

func Dot11Info(id layers.Dot11InformationElementID, info []byte) *layers.Dot11InformationElement {
	return &layers.Dot11InformationElement{
		ID:     id,
		Length: uint8(len(info) & 0xff),
		Info:   info,
	}
}

func NewDot11Beacon(conf Dot11ApConfig, seq uint16) (error, []byte) {
	flags := openFlags
	if conf.Encryption {
		flags = wpaFlags
	}

	stack := []gopacket.SerializableLayer{
		&layers.RadioTap{
			DBMAntennaSignal: int8(-10),
			ChannelFrequency: layers.RadioTapChannelFrequency(network.Dot11Chan2Freq(conf.Channel)),
		},
		&layers.Dot11{
			Address1:       network.BroadcastHw,
			Address2:       conf.BSSID,
			Address3:       conf.BSSID,
			Type:           layers.Dot11TypeMgmtBeacon,
			SequenceNumber: seq,
		},
		&layers.Dot11MgmtBeacon{
			Flags:    uint16(flags),
			Interval: 100,
		},
		Dot11Info(layers.Dot11InformationElementIDSSID, []byte(conf.SSID)),
		Dot11Info(layers.Dot11InformationElementIDRates, fakeApRates),
		Dot11Info(layers.Dot11InformationElementIDDSSet, []byte{byte(conf.Channel & 0xff)}),
	}

	if conf.Encryption {
		stack = append(stack, &layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDRSNInfo,
			Length: uint8(len(fakeApWpaRSN) & 0xff),
			Info:   fakeApWpaRSN,
		})
	}

	return Serialize(stack...)
}

func NewDot11BeaconWithCSA(seq uint16, channel int8, ssid string, bssid net.HardwareAddr) (error, []byte) {
	return Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       network.BroadcastHw,
			Address2:       bssid,
			Address3:       bssid,
			SequenceNumber: seq,
			FragmentNumber: 0,
			Type:           layers.Dot11TypeMgmtBeacon,
		},
		&layers.Dot11MgmtBeacon{
			Timestamp: 0,
			Interval:  0x64,
			Flags:     65535,
		},
		Dot11Info(layers.Dot11InformationElementIDSSID, []byte(ssid)),
		Dot11Info(layers.Dot11InformationElementIDRates, fakeApRates),
		Dot11Info(layers.Dot11InformationElementIDSwitchChannelAnnounce, []byte{0,byte(channel),0}),

	)

}

func NewDot11Deauth(a1 net.HardwareAddr, a2 net.HardwareAddr, a3 net.HardwareAddr, seq uint16) (error, []byte) {
	return Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       a1,
			Address2:       a2,
			Address3:       a3,
			Type:           layers.Dot11TypeMgmtDeauthentication,
			SequenceNumber: seq,
		},
		&layers.Dot11MgmtDeauthentication{
			Reason: layers.Dot11ReasonClass2FromNonAuth,
		},
	)
}

func NewDot11Auth(sta net.HardwareAddr, apBSSID net.HardwareAddr, seq uint16) (error, []byte) {
	return Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       apBSSID,
			Address2:       sta,
			Address3:       apBSSID,
			Type:           layers.Dot11TypeMgmtAuthentication,
			SequenceNumber: seq,
			FragmentNumber: 0,
			DurationID:     durationID,
		},
		&layers.Dot11MgmtAuthentication{
			Algorithm: layers.Dot11AlgorithmOpen,
			Sequence:  1,
			Status:    layers.Dot11StatusSuccess,
		},
	)
}

func NewDot11AssociationRequest(sta net.HardwareAddr, apBSSID net.HardwareAddr, apESSID string, seq uint16) (error, []byte) {
	return Serialize(
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       apBSSID,
			Address2:       sta,
			Address3:       apBSSID,
			Type:           layers.Dot11TypeMgmtAssociationReq,
			SequenceNumber: seq,
			FragmentNumber: 0,
			DurationID:     durationID,
		},
		// as seen on wireshark ...
		&layers.Dot11MgmtAssociationReq{
			CapabilityInfo: capabilityInfo,
			ListenInterval: listenInterval,
		},
		Dot11Info(layers.Dot11InformationElementIDSSID, []byte(apESSID)),
		Dot11Info(layers.Dot11InformationElementIDRates, assocRates),
		Dot11Info(layers.Dot11InformationElementIDESRates, assocESRates),
		Dot11Info(layers.Dot11InformationElementIDRSNInfo, assocRSNInfo),
		Dot11Info(layers.Dot11InformationElementIDHTCapabilities, assocCapabilities),
		&layers.Dot11InformationElement{
			ID:     layers.Dot11InformationElementIDVendor,
			Length: 7,
			OUI:    []byte{0, 0x50, 0xf2, 0x02},
			Info:   []byte{0, 0x01, 0},
		},
	)
}

func Dot11Parse(packet gopacket.Packet) (ok bool, radiotap *layers.RadioTap, dot11 *layers.Dot11) {
	ok = false
	radiotap = nil
	dot11 = nil

	radiotapLayer := packet.Layer(layers.LayerTypeRadioTap)
	if radiotapLayer == nil {
		return
	}
	radiotap, ok = radiotapLayer.(*layers.RadioTap)
	if !ok || radiotap == nil {
		return
	}

	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer == nil {
		ok = false
		return
	}

	dot11, ok = dot11Layer.(*layers.Dot11)
	return
}

func Dot11ParseIDSSID(packet gopacket.Packet) (bool, string) {
	for _, layer := range packet.Layers() {
		if layer.LayerType() == layers.LayerTypeDot11InformationElement {
			dot11info, ok := layer.(*layers.Dot11InformationElement)
			if ok && dot11info.ID == layers.Dot11InformationElementIDSSID {
				if len(dot11info.Info) == 0 {
					return true, "<hidden>"
				}
				return true, string(dot11info.Info)
			}
		}
	}

	return false, ""
}

func Dot11ParseEncryption(packet gopacket.Packet, dot11 *layers.Dot11) (bool, string, string, string) {
	var i uint16
	enc := ""
	cipher := ""
	auth := ""
	found := false

	if dot11.Flags.WEP() {
		found = true
		enc = "WEP"
	}

	for _, layer := range packet.Layers() {
		if layer.LayerType() == layers.LayerTypeDot11InformationElement {
			info, ok := layer.(*layers.Dot11InformationElement)
			if ok {
				found = true
				if info.ID == layers.Dot11InformationElementIDRSNInfo {
					enc = "WPA2"
					rsn, err := Dot11InformationElementRSNInfoDecode(info.Info)
					if err == nil {
						for i = 0; i < rsn.Pairwise.Count; i++ {
							cipher = rsn.Pairwise.Suites[i].Type.String()
						}
						for i = 0; i < rsn.AuthKey.Count; i++ {
							auth = rsn.AuthKey.Suites[i].Type.String()
						}
					}
				} else if enc == "" && info.ID == layers.Dot11InformationElementIDVendor && info.Length >= 8 && bytes.Equal(info.OUI, wpaSignatureBytes) && bytes.HasPrefix(info.Info, []byte{1, 0}) {
					enc = "WPA"
					vendor, err := Dot11InformationElementVendorInfoDecode(info.Info)
					if err == nil {
						for i = 0; i < vendor.Unicast.Count; i++ {
							cipher = vendor.Unicast.Suites[i].Type.String()
						}
						for i = 0; i < vendor.AuthKey.Count; i++ {
							auth = vendor.AuthKey.Suites[i].Type.String()
						}
					}
				}
			}
		}
	}

	if found && enc == "" {
		enc = "OPEN"
	}

	return found, enc, cipher, auth

}

func Dot11IsDataFor(dot11 *layers.Dot11, station net.HardwareAddr) bool {
	// only check data packets of connected stations
	if dot11.Type.MainType() != layers.Dot11TypeData {
		return false
	}
	// packet going to this specific BSSID?
	return bytes.Equal(dot11.Address1, station)
}

func Dot11ParseDSSet(packet gopacket.Packet) (bool, int) {
	channel := 0
	found := false
	for _, layer := range packet.Layers() {
		info, ok := layer.(*layers.Dot11InformationElement)
		if ok {
			if info.ID == layers.Dot11InformationElementIDDSSet {
				channel, _ = Dot11InformationElementIDDSSetDecode(info.Info)
				found = true
				break
			}
		}
	}

	return found, channel
}

func Dot11ParseEAPOL(packet gopacket.Packet, dot11 *layers.Dot11) (ok bool, key *layers.EAPOLKey, apMac net.HardwareAddr, staMac net.HardwareAddr) {
	ok = false
	// ref. https://wlan1nde.wordpress.com/2014/10/27/4-way-handshake/
	if keyLayer := packet.Layer(layers.LayerTypeEAPOLKey); keyLayer != nil {
		if key = keyLayer.(*layers.EAPOLKey); key.KeyType == layers.EAPOLKeyTypePairwise {
			ok = true
			if dot11.Flags.FromDS() {
				staMac = dot11.Address1
				apMac = dot11.Address2
			} else if dot11.Flags.ToDS() {
				staMac = dot11.Address2
				apMac = dot11.Address1
			}
		}
	}
	return
}
