package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Dot11CipherType uint8

const (
	Dot11CipherWep    Dot11CipherType = 1
	Dot11CipherTkip   Dot11CipherType = 2
	Dot11CipherWrap   Dot11CipherType = 3
	Dot11CipherCcmp   Dot11CipherType = 4
	Dot11CipherWep104 Dot11CipherType = 5
)

func (a Dot11CipherType) String() string {
	switch a {
	case Dot11CipherWep:
		return "WEP"
	case Dot11CipherTkip:
		return "TKIP"
	case Dot11CipherWrap:
		return "WRAP"
	case Dot11CipherCcmp:
		return "CCMP"
	case Dot11CipherWep104:
		return "WEP104"
	default:
		return "UNK"
	}
}

type Dot11AuthType uint8

const (
	Dot11AuthMgt Dot11AuthType = 1
	Dot11AuthPsk Dot11AuthType = 2
)

func (a Dot11AuthType) String() string {
	switch a {
	case Dot11AuthMgt:
		return "MGT"
	case Dot11AuthPsk:
		return "PSK"
	default:
		return "UNK"
	}
}

type CipherSuite struct {
	OUI  []byte // 3 bytes
	Type Dot11CipherType
}

type AuthSuite struct {
	OUI  []byte // 3 bytes
	Type Dot11AuthType
}

type CipherSuiteSelector struct {
	Count  uint16
	Suites []CipherSuite
}
type AuthSuiteSelector struct {
	Count  uint16
	Suites []AuthSuite
}

type VendorInfo struct {
	WPAVersion uint16
	Multicast  CipherSuite
	Unicast    CipherSuiteSelector
	AuthKey    AuthSuiteSelector
}

type RSNInfo struct {
	Version  uint16
	Group    CipherSuite
	Pairwise CipherSuiteSelector
	AuthKey  AuthSuiteSelector
}

func Dot11InformationElementVendorInfoDecode(vendorInfo []byte) (VendorInfo, error) {
	var v VendorInfo
	var i uint16
	if len(vendorInfo) < 15 {
		return v, fmt.Errorf("VendorInfo packet length %v too short, %v required", len(vendorInfo), 15)
	}

	v.WPAVersion = binary.LittleEndian.Uint16(vendorInfo[0:2])
	v.Multicast.OUI = vendorInfo[2:5]
	v.Multicast.Type = Dot11CipherType(vendorInfo[5])

	v.Unicast.Count = binary.LittleEndian.Uint16(vendorInfo[6:8])

	p := 8
	for i = 0; i < v.Unicast.Count && p < len(vendorInfo); i++ {
		var suite CipherSuite
		suite.OUI = vendorInfo[p : p+3]
		suite.Type = Dot11CipherType(vendorInfo[p+3])
		v.Unicast.Suites = append(v.Unicast.Suites, suite)
		p = p + 4
	}

	v.AuthKey.Count = binary.LittleEndian.Uint16(vendorInfo[p : p+2])
	p = p + 2
	for i = 0; i < v.AuthKey.Count && p < len(vendorInfo); i++ {
		var suite AuthSuite
		suite.OUI = vendorInfo[p : p+3]
		suite.Type = Dot11AuthType(vendorInfo[p+3])
		v.AuthKey.Suites = append(v.AuthKey.Suites, suite)
		p = p + 4
	}

	return v, nil
}

func Dot11InformationElementRSNInfoDecode(info []byte) (RSNInfo, error) {
	var rsn RSNInfo
	if len(info) < 20 {
		return rsn, fmt.Errorf("RSNInfo packet length %v too short, %v required", len(info), 20)
	}

	rsn.Version = binary.LittleEndian.Uint16(info[0:2])
	rsn.Group.OUI = info[2:5]
	rsn.Group.Type = Dot11CipherType(info[5])
	rsn.Pairwise.Count = binary.LittleEndian.Uint16(info[6:8])

	p := 8
	for i := uint16(0); i < rsn.Pairwise.Count && p < len(info); i++ {
		var suite CipherSuite
		suite.OUI = info[p : p+3]
		suite.Type = Dot11CipherType(info[p+3])
		rsn.Pairwise.Suites = append(rsn.Pairwise.Suites, suite)
		p = p + 4
	}

	rsn.AuthKey.Count = binary.LittleEndian.Uint16(info[p : p+2])
	p = p + 2
	for i := uint16(0); i < rsn.AuthKey.Count && p < len(info); i++ {
		var suite AuthSuite
		suite.OUI = info[p : p+3]
		suite.Type = Dot11AuthType(info[p+3])
		rsn.AuthKey.Suites = append(rsn.AuthKey.Suites, suite)
		p = p + 4
	}

	return rsn, nil
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

func Dot11Parse(packet gopacket.Packet) (ok bool, radiotap *layers.RadioTap, dot11 *layers.Dot11) {
	ok = false
	radiotap = nil
	dot11 = nil

	radiotapLayer := packet.Layer(layers.LayerTypeRadioTap)
	if radiotapLayer == nil {
		return
	}
	radiotap, ok = radiotapLayer.(*layers.RadioTap)
	if ok == false || radiotap == nil {
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
			if ok == true && dot11info.ID == layers.Dot11InformationElementIDSSID {
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
			if ok == true {
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
				} else if enc == "" && info.ID == layers.Dot11InformationElementIDVendor && info.Length >= 8 && bytes.Compare(info.OUI, []byte{0, 0x50, 0xf2, 1}) == 0 && bytes.HasPrefix(info.Info, []byte{1, 0}) {
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
	return bytes.Compare(dot11.Address1, station) == 0
}
