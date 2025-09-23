package packets

import (
	"bytes"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func TestSerializationOptions(t *testing.T) {
	// Verify the global serialization options are set correctly
	if !SerializationOptions.FixLengths {
		t.Error("SerializationOptions.FixLengths should be true")
	}
	if !SerializationOptions.ComputeChecksums {
		t.Error("SerializationOptions.ComputeChecksums should be true")
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name        string
		layers      []gopacket.SerializableLayer
		expectError bool
		minLength   int
	}{
		{
			name: "simple ethernet frame",
			layers: []gopacket.SerializableLayer{
				&layers.Ethernet{
					SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
					DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
					EthernetType: layers.EthernetTypeIPv4,
				},
			},
			expectError: false,
			minLength:   14, // Ethernet header
		},
		{
			name: "ethernet with IPv4",
			layers: []gopacket.SerializableLayer{
				&layers.Ethernet{
					SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
					DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
					EthernetType: layers.EthernetTypeIPv4,
				},
				&layers.IPv4{
					Version:  4,
					Protocol: layers.IPProtocolTCP,
					TTL:      64,
					SrcIP:    []byte{192, 168, 1, 1},
					DstIP:    []byte{192, 168, 1, 2},
				},
			},
			expectError: false,
			minLength:   34, // Ethernet + IPv4 headers
		},
		{
			name: "complete TCP packet",
			layers: func() []gopacket.SerializableLayer {
				ip4 := &layers.IPv4{
					Version:  4,
					Protocol: layers.IPProtocolTCP,
					TTL:      64,
					SrcIP:    []byte{192, 168, 1, 1},
					DstIP:    []byte{192, 168, 1, 2},
				}
				tcp := &layers.TCP{
					SrcPort: 12345,
					DstPort: 80,
					Seq:     1000,
					Ack:     0,
					SYN:     true,
					Window:  65535,
				}
				tcp.SetNetworkLayerForChecksum(ip4)
				return []gopacket.SerializableLayer{
					&layers.Ethernet{
						SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
						DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
						EthernetType: layers.EthernetTypeIPv4,
					},
					ip4,
					tcp,
				}
			}(),
			expectError: false,
			minLength:   54, // Ethernet + IPv4 + TCP headers
		},
		{
			name:        "empty layers",
			layers:      []gopacket.SerializableLayer{},
			expectError: false,
			minLength:   0,
		},
		{
			name: "layer with payload",
			layers: []gopacket.SerializableLayer{
				&layers.Ethernet{
					SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
					DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
					EthernetType: layers.EthernetTypeIPv4,
				},
				gopacket.Payload([]byte("Hello, World!")),
			},
			expectError: false,
			minLength:   27, // Ethernet header + payload
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, data := Serialize(tt.layers...)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				if len(data) < tt.minLength {
					t.Errorf("Data length %d is less than expected minimum %d", len(data), tt.minLength)
				}

				// For non-empty results, verify we can parse it back
				if len(data) > 0 && len(tt.layers) > 0 {
					packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
					if packet == nil {
						t.Error("Failed to parse serialized data")
					}
				}
			}
		})
	}
}

func TestSerializeWithChecksum(t *testing.T) {
	// Test that checksums are computed correctly
	ip4 := &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		TTL:      64,
		SrcIP:    []byte{192, 168, 1, 1},
		DstIP:    []byte{192, 168, 1, 2},
	}

	udp := &layers.UDP{
		SrcPort: 12345,
		DstPort: 53,
	}

	// Set network layer for checksum computation
	udp.SetNetworkLayerForChecksum(ip4)

	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	err, data := Serialize(eth, ip4, udp)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Parse back and verify checksums
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip := ipLayer.(*layers.IPv4)
		// The checksum should be computed (non-zero)
		if ip.Checksum == 0 {
			t.Error("IPv4 checksum was not computed")
		}
	} else {
		t.Error("IPv4 layer not found in packet")
	}

	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		// The checksum should be computed (non-zero for UDP over IPv4)
		if udp.Checksum == 0 {
			t.Error("UDP checksum was not computed")
		}
	} else {
		t.Error("UDP layer not found in packet")
	}
}

func TestSerializeFixLengths(t *testing.T) {
	// Test that lengths are fixed correctly
	ip4 := &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolTCP,
		TTL:      64,
		SrcIP:    []byte{10, 0, 0, 1},
		DstIP:    []byte{10, 0, 0, 2},
		// Don't set Length - it should be computed
	}

	tcp := &layers.TCP{
		SrcPort: 80,
		DstPort: 12345,
		Seq:     1000,
		SYN:     true,
		Window:  65535,
	}

	tcp.SetNetworkLayerForChecksum(ip4)

	payload := gopacket.Payload([]byte("Test payload data"))

	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	err, data := Serialize(eth, ip4, tcp, payload)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Parse back and verify lengths
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip := ipLayer.(*layers.IPv4)
		expectedLen := 20 + 20 + len("Test payload data") // IPv4 header + TCP header + payload
		if ip.Length != uint16(expectedLen) {
			t.Errorf("IPv4 length = %d, want %d", ip.Length, expectedLen)
		}
	} else {
		t.Error("IPv4 layer not found in packet")
	}
}

func TestSerializeErrorHandling(t *testing.T) {
	// Test serialization with an invalid layer configuration
	// This test is a bit tricky because gopacket is quite forgiving
	// We'll create a scenario that might fail in serialization

	// Create an ethernet layer with invalid type for the next layer
	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	// Follow with a non-IPv4 layer when IPv4 is expected
	// This actually won't cause an error in gopacket, so we test that errors are handled
	tcp := &layers.TCP{
		SrcPort: 80,
		DstPort: 12345,
	}

	err, data := Serialize(eth, tcp)
	// This might not actually error, but we're testing the error handling path
	if err != nil {
		// Error path - should return nil data
		if data != nil {
			t.Error("When error occurs, data should be nil")
		}
	} else {
		// Success path - should return data
		if data == nil {
			t.Error("When no error, data should not be nil")
		}
	}
}

func TestSerializeMultiplePackets(t *testing.T) {
	// Test serializing multiple different packet types in sequence
	srcMAC := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	dstMAC := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	packets := []struct {
		name   string
		layers []gopacket.SerializableLayer
	}{
		{
			name: "ARP request",
			layers: []gopacket.SerializableLayer{
				&layers.Ethernet{
					SrcMAC:       srcMAC,
					DstMAC:       dstMAC,
					EthernetType: layers.EthernetTypeARP,
				},
				&layers.ARP{
					AddrType:          layers.LinkTypeEthernet,
					Protocol:          layers.EthernetTypeIPv4,
					HwAddressSize:     6,
					ProtAddressSize:   4,
					Operation:         layers.ARPRequest,
					SourceHwAddress:   srcMAC,
					SourceProtAddress: []byte{192, 168, 1, 100},
					DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
					DstProtAddress:    []byte{192, 168, 1, 1},
				},
			},
		},
		{
			name: "ICMP echo",
			layers: []gopacket.SerializableLayer{
				&layers.Ethernet{
					SrcMAC:       srcMAC,
					DstMAC:       dstMAC,
					EthernetType: layers.EthernetTypeIPv4,
				},
				&layers.IPv4{
					Version:  4,
					Protocol: layers.IPProtocolICMPv4,
					TTL:      64,
					SrcIP:    []byte{192, 168, 1, 100},
					DstIP:    []byte{8, 8, 8, 8},
				},
				&layers.ICMPv4{
					TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
					Id:       1,
					Seq:      1,
				},
				gopacket.Payload([]byte("ping")),
			},
		},
	}

	for _, pkt := range packets {
		t.Run(pkt.name, func(t *testing.T) {
			err, data := Serialize(pkt.layers...)
			if err != nil {
				t.Errorf("Failed to serialize %s: %v", pkt.name, err)
			}
			if len(data) == 0 {
				t.Errorf("Serialized %s has zero length", pkt.name)
			}
		})
	}
}

// Benchmarks
func BenchmarkSerialize(b *testing.B) {
	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolTCP,
		TTL:      64,
		SrcIP:    []byte{192, 168, 1, 1},
		DstIP:    []byte{192, 168, 1, 2},
	}

	tcp := &layers.TCP{
		SrcPort: 12345,
		DstPort: 80,
		Seq:     1000,
		SYN:     true,
		Window:  65535,
	}

	tcp.SetNetworkLayerForChecksum(ip4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Serialize(eth, ip4, tcp)
	}
}

func BenchmarkSerializeWithPayload(b *testing.B) {
	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		TTL:      64,
		SrcIP:    []byte{192, 168, 1, 1},
		DstIP:    []byte{192, 168, 1, 2},
	}

	udp := &layers.UDP{
		SrcPort: 12345,
		DstPort: 53,
	}

	udp.SetNetworkLayerForChecksum(ip4)

	payload := gopacket.Payload(bytes.Repeat([]byte("x"), 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Serialize(eth, ip4, udp, payload)
	}
}
