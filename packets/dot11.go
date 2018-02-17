package packets

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

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
