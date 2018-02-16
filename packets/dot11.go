package packets

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func NewDot11Deauth(a1 net.HardwareAddr, a2 net.HardwareAddr, a3 net.HardwareAddr, t layers.Dot11Type, reason layers.Dot11Reason, seq uint16) (error, []byte) {
	var (
		deauth        layers.Dot11MgmtDeauthentication
		dot11Layer    layers.Dot11
		radioTapLayer layers.RadioTap
	)

	deauth.Reason = reason

	dot11Layer.Address1 = a1
	dot11Layer.Address2 = a2
	dot11Layer.Address3 = a3
	dot11Layer.Type = t
	dot11Layer.SequenceNumber = seq

	return Serialize(
		&radioTapLayer,
		&dot11Layer,
		&deauth,
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
