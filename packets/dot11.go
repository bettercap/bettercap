package packets

import (
	"bytes"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var wpaSignatureBytes = []byte{0, 0x50, 0xf2, 1}

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
				} else if enc == "" && info.ID == layers.Dot11InformationElementIDVendor && info.Length >= 8 && bytes.Compare(info.OUI, wpaSignatureBytes) == 0 && bytes.HasPrefix(info.Info, []byte{1, 0}) {
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
