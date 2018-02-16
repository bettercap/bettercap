package packets

import (
	"net"

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
