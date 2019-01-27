package packets

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type wpsAttrType int

const (
	wpsHex wpsAttrType = 0
	wpsStr wpsAttrType = 1
)

type wpsAttr struct {
	Name string
	Type wpsAttrType
	Desc map[string]string
}

var (
	wpsSignatureBytes = []byte{0x00, 0x50, 0xf2, 0x04}
	wpsAttributes     = map[uint16]wpsAttr{
		0x104A: wpsAttr{Name: "Version", Desc: map[string]string{
			"10": "1.0",
			"11": "1.1",
		}},
		0x1044: wpsAttr{Name: "State", Desc: map[string]string{
			"01": "Not Configured",
			"02": "Configured",
		}},
		0x1057: wpsAttr{Name: "AP Setup Locked"},
		0x1041: wpsAttr{Name: "Selected Registrar"},
		0x1012: wpsAttr{Name: "Device Password ID"},
		0x1053: wpsAttr{Name: "Selected Registrar Config Methods", Desc: map[string]string{
			"0001": "USB",
			"0002": "Ethernet",
			"0004": "Label",
			"0008": "Display",
			"0010": "External NFC",
			"0020": "Internal NFC",
			"0040": "NFC Interface",
			"0080": "Push Button",
			"0100": "Keypad",
		}},
		0x103B: wpsAttr{Name: "Response Type"},
		0x1047: wpsAttr{Name: "UUID-E"},
		0x1021: wpsAttr{Name: "Manufacturer", Type: wpsStr},
		0x1023: wpsAttr{Name: "Model Name", Type: wpsStr},
		0x1024: wpsAttr{Name: "Model Number", Type: wpsStr},
		0x1042: wpsAttr{Name: "Serial Number", Type: wpsStr},
		0x1054: wpsAttr{Name: "Primary Device Type"},
		0x1011: wpsAttr{Name: "Device Name", Type: wpsStr},
		0x1008: wpsAttr{Name: "Config Methods"},
		0x103C: wpsAttr{Name: "RF Bands"},
		0x1045: wpsAttr{Name: "SSID", Type: wpsStr},
		0x102D: wpsAttr{Name: "OS Version", Type: wpsStr},
		0x1049: wpsAttr{Name: "Vendor Extension"},
	}
)

func dot11ParseWPSData(data []byte) (ok bool, info map[string]string) {
	info = map[string]string{}
	size := len(data)

	for offset := 0; offset < size; {
		tagId := binary.BigEndian.Uint16(data[offset:])
		offset += 2
		tagLen := binary.BigEndian.Uint16(data[offset:])
		offset += 2
		tagData := data[offset : offset+int(tagLen)]

		if attr, found := wpsAttributes[tagId]; found {
			val := ""
			if attr.Type == wpsStr {
				val = string(tagData)
			} else {
				val = hex.EncodeToString(tagData)
			}

			if attr.Desc != nil {
				if desc, found := attr.Desc[val]; found {
					val = desc
				}
			}

			info[attr.Name] = val
		} else {
			info[fmt.Sprintf("0x%X", tagId)] = hex.EncodeToString(tagData)
		}

		offset += int(tagLen)
	}

	return true, info
}

func Dot11ParseWPS(packet gopacket.Packet, dot11 *layers.Dot11) (ok bool, bssid net.HardwareAddr, info map[string]string) {
	ok = false
	for _, layer := range packet.Layers() {
		if layer.LayerType() == layers.LayerTypeDot11InformationElement {
			if dot11info, infoOk := layer.(*layers.Dot11InformationElement); infoOk && dot11info.ID == layers.Dot11InformationElementIDVendor {
				if bytes.Equal(dot11info.OUI, wpsSignatureBytes) {
					bssid = dot11.Address3
					ok, info = dot11ParseWPSData(dot11info.Info)
					return
				}
			}
		}
	}
	return
}
