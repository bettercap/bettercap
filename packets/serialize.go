package packets

import (
	"github.com/google/gopacket"
)

var SerializationOptions = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

func Serialize(layers ...gopacket.SerializableLayer) (error, []byte) {
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, SerializationOptions, layers...); err != nil {
		return err, nil
	}
	return nil, buf.Bytes()
}
