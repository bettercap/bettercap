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

var (
	wpsSignatureBytes = []byte{0x00, 0x50, 0xf2, 0x04}
)

func wpsUint16At(data []byte, size int, offset *int) (bool, uint16) {
	if *offset <= size-2 {
		off := *offset
		*offset += 2
		return true, binary.BigEndian.Uint16(data[off:])
	}
	// fmt.Printf("uint16At( data(%d), off=%d )\n", size, *offset)
	return false, 0
}

func wpsDataAt(data []byte, size int, offset *int, num int) (bool, []byte) {
	max := size - num
	if *offset <= max {
		off := *offset
		*offset += num
		return true, data[off : off+num]
	}
	// fmt.Printf("dataAt( data(%d), off=%d, num=%d )\n", size, *offset, num)
	return false, nil
}

func dot11ParseWPSTag(id uint16, size uint16, data []byte, info *map[string]string) {
	name := ""
	val := ""

	if attr, found := wpsAttributes[id]; found {
		name = attr.Name
		if attr.Type == wpsStr {
			val = string(data)
		} else {
			val = hex.EncodeToString(data)
		}

		if attr.Desc != nil {
			if desc, found := attr.Desc[val]; found {
				val = desc
			}
		}

		if attr.Func != nil {
			val = attr.Func(data, info)
		}
	} else {
		name = fmt.Sprintf("0x%X", id)
		val = hex.EncodeToString(data)
	}

	if val != "" {
		(*info)[name] = val
	}
}

func dot11ParseWPSData(data []byte) (ok bool, info map[string]string) {
	info = map[string]string{}
	size := len(data)

	for offset := 0; offset < size; {
		ok := false
		tagId := uint16(0)
		tagLen := uint16(0)
		tagData := []byte(nil)

		if ok, tagId = wpsUint16At(data, size, &offset); !ok {
			break
		} else if ok, tagLen = wpsUint16At(data, size, &offset); !ok {
			break
		} else if ok, tagData = wpsDataAt(data, size, &offset, int(tagLen)); !ok {
			break
		} else {
			dot11ParseWPSTag(tagId, tagLen, tagData, &info)
		}
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
