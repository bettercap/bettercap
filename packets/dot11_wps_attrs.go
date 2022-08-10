package packets

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type wpsAttrType int

const (
	wpsHex wpsAttrType = 0
	wpsStr wpsAttrType = 1
)

type wpsAttr struct {
	Name string
	Type wpsAttrType
	Func func([]byte, *map[string]string) string
	Desc map[string]string
}

type wpsDevType struct {
	Category string
	Subcats  map[uint16]string
}

var (
	wfaExtensionBytes = []byte{0x00, 0x37, 0x2a}
	wpsVersion2ID     = uint8(0x00)
	wpsVersionDesc    = map[string]string{
		"10": "1.0",
		"11": "1.1",
		"20": "2.0",
	}

	wpsDeviceTypes = map[uint16]wpsDevType{
		0x0001: {"Computer", map[uint16]string{
			0x0001: "PC",
			0x0002: "Server",
			0x0003: "Media Center",
		}},
		0x0002: {"Input Device", map[uint16]string{}},
		0x0003: {"Printers, Scanners, Faxes and Copiers", map[uint16]string{
			0x0001: "Printer",
			0x0002: "Scanner",
		}},
		0x0004: {"Camera", map[uint16]string{
			0x0001: "Digital Still Camera",
		}},
		0x0005: {"Storage", map[uint16]string{
			0x0001: "NAS",
		}},
		0x0006: {"Network Infra", map[uint16]string{
			0x0001: "AP",
			0x0002: "Router",
			0x0003: "Switch",
		}},

		0x0007: {"Display", map[uint16]string{
			0x0001: "TV",
			0x0002: "Electronic Picture Frame",
			0x0003: "Projector",
		}},

		0x0008: {"Multimedia Device", map[uint16]string{
			0x0001: "DAR",
			0x0002: "PVR",
			0x0003: "MCX",
		}},

		0x0009: {"Gaming Device", map[uint16]string{
			0x0001: "XBox",
			0x0002: "XBox360",
			0x0003: "Playstation",
		}},
		0x000F: {"Telephone", map[uint16]string{
			0x0001: "Windows Mobile",
		}},
	}

	wpsAttributes = map[uint16]wpsAttr{
		0x104A: {Name: "Version", Desc: wpsVersionDesc},
		0x1044: {Name: "State", Desc: map[string]string{
			"01": "Not Configured",
			"02": "Configured",
		}},
		0x1012: {Name: "Device Password ID", Desc: map[string]string{
			"0000": "Pin",
			"0004": "PushButton",
		}},
		0x103B: {Name: "Response Type", Desc: map[string]string{
			"00": "Enrollee Info",
			"01": "Enrollee",
			"02": "Registrar",
			"03": "AP",
		}},

		0x1054: {Name: "Primary Device Type", Func: dot11ParseWPSDeviceType},
		0x1049: {Name: "Vendor Extension", Func: dot11ParseWPSVendorExtension},
		0x1053: {Name: "Selected Registrar Config Methods", Func: dot11ParseWPSConfigMethods},
		0x1008: {Name: "Config Methods", Func: dot11ParseWPSConfigMethods},
		0x103C: {Name: "RF Bands", Func: dott11ParseWPSBands},

		0x1057: {Name: "AP Setup Locked"},
		0x1041: {Name: "Selected Registrar"},
		0x1047: {Name: "UUID-E"},
		0x1021: {Name: "Manufacturer", Type: wpsStr},
		0x1023: {Name: "Model Name", Type: wpsStr},
		0x1024: {Name: "Model Number", Type: wpsStr},
		0x1042: {Name: "Serial Number", Type: wpsStr},
		0x1011: {Name: "Device Name", Type: wpsStr},
		0x1045: {Name: "SSID", Type: wpsStr},
		0x102D: {Name: "OS Version", Type: wpsStr},
	}

	wpsConfigs = map[uint16]string{
		0x0001: "USB",
		0x0002: "Ethernet",
		0x0004: "Label",
		0x0008: "Display",
		0x0010: "External NFC",
		0x0020: "Internal NFC",
		0x0040: "NFC Interface",
		0x0080: "Push Button",
		0x0100: "Keypad",
	}

	wpsBands = map[uint8]string{
		0x01: "2.4Ghz",
		0x02: "5.0Ghz",
	}
)

func dott11ParseWPSBands(data []byte, info *map[string]string) string {
	if len(data) == 1 {
		mask := uint8(data[0])
		bands := []string{}

		for bit, band := range wpsBands {
			if mask&bit != 0 {
				bands = append(bands, band)
			}
		}

		if len(bands) > 0 {
			return strings.Join(bands, ", ")
		}
	}

	return hex.EncodeToString(data)
}

func dot11ParseWPSConfigMethods(data []byte, info *map[string]string) string {
	if len(data) == 2 {
		mask := binary.BigEndian.Uint16(data)
		configs := []string{}

		for bit, conf := range wpsConfigs {
			if mask&bit != 0 {
				configs = append(configs, conf)
			}
		}

		if len(configs) > 0 {
			return strings.Join(configs, ", ")
		}
	}

	return hex.EncodeToString(data)
}

func dot11ParseWPSVendorExtension(data []byte, info *map[string]string) string {
	if len(data) > 3 && bytes.Equal(data[0:3], wfaExtensionBytes) {
		size := len(data)
		for offset := 3; offset < size; {
			idByte := uint8(data[offset])
			if next := offset + 1; next < size {
				sizeByte := uint8(data[next])
				if idByte == wpsVersion2ID {
					if next = offset + 2; next < size {
						verByte := fmt.Sprintf("%x", data[next])
						(*info)["Version"] = wpsVersionDesc[verByte]
						if next = offset + 3; next < size {
							data = data[next:]
						}
						break
					}
				}
				offset += int(sizeByte) + 2
			} else {
				break
			}
		}
	}
	return hex.EncodeToString(data)
}

func dot11ParseWPSDeviceType(data []byte, info *map[string]string) string {
	if len(data) == 8 {
		catId := binary.BigEndian.Uint16(data[0:2])
		oui := data[2:6]
		subCatId := binary.BigEndian.Uint16(data[6:8])
		if cat, found := wpsDeviceTypes[catId]; found {
			if sub, found := cat.Subcats[subCatId]; found {
				return fmt.Sprintf("%s (oui:%x)", sub, oui)
			}
			return fmt.Sprintf("%s (oui:%x)", cat.Category, oui)
		}
		return fmt.Sprintf("cat:%x sub:%x oui:%x %x", catId, subCatId, oui, data)
	}
	return hex.EncodeToString(data)
}
