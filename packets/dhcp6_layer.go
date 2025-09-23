package packets

import (
	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

type DHCPv6Layer struct {
	Raw []byte
}

func (l *DHCPv6Layer) LayerType() gopacket.LayerType { return layers.LayerTypeDHCPv6 }

func (l DHCPv6Layer) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(len(l.Raw))
	if err != nil {
		return err
	}

	copy(bytes, l.Raw)
	return nil
}
