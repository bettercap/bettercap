package packets

import (
	"github.com/google/gopacket"
)

func Serialize(layers ...gopacket.SerializableLayer) (error, []byte) {
	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err := gopacket.SerializeLayers(buf, opts, layers...); err != nil {
		return err, nil
	}

	return nil, buf.Bytes()
}
