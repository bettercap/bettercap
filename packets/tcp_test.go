package packets

import (
	"bytes"
	"net"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func TestNewTCPSyn(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		fromHW      string
		to          string
		toHW        string
		srcPort     int
		dstPort     int
		expectError bool
		expectIPv6  bool
	}{
		{
			name:        "IPv4 TCP SYN",
			from:        "192.168.1.100",
			fromHW:      "aa:bb:cc:dd:ee:ff",
			to:          "192.168.1.200",
			toHW:        "11:22:33:44:55:66",
			srcPort:     12345,
			dstPort:     80,
			expectError: false,
			expectIPv6:  false,
		},
		{
			name:        "IPv6 TCP SYN",
			from:        "2001:db8::1",
			fromHW:      "aa:bb:cc:dd:ee:ff",
			to:          "2001:db8::2",
			toHW:        "11:22:33:44:55:66",
			srcPort:     54321,
			dstPort:     443,
			expectError: false,
			expectIPv6:  true,
		},
		{
			name:        "IPv4 with different ports",
			from:        "10.0.0.1",
			fromHW:      "01:23:45:67:89:ab",
			to:          "10.0.0.2",
			toHW:        "cd:ef:01:23:45:67",
			srcPort:     8080,
			dstPort:     3306,
			expectError: false,
			expectIPv6:  false,
		},
		{
			name:        "IPv6 link-local addresses",
			from:        "fe80::1",
			fromHW:      "aa:bb:cc:dd:ee:ff",
			to:          "fe80::2",
			toHW:        "11:22:33:44:55:66",
			srcPort:     1234,
			dstPort:     5678,
			expectError: false,
			expectIPv6:  true,
		},
		{
			name:        "IPv4 loopback",
			from:        "127.0.0.1",
			fromHW:      "00:00:00:00:00:00",
			to:          "127.0.0.1",
			toHW:        "00:00:00:00:00:00",
			srcPort:     9000,
			dstPort:     9001,
			expectError: false,
			expectIPv6:  false,
		},
		{
			name:        "IPv6 loopback",
			from:        "::1",
			fromHW:      "00:00:00:00:00:00",
			to:          "::1",
			toHW:        "00:00:00:00:00:00",
			srcPort:     9000,
			dstPort:     9001,
			expectError: false,
			expectIPv6:  true,
		},
		{
			name:        "Max port number",
			from:        "192.168.1.1",
			fromHW:      "aa:bb:cc:dd:ee:ff",
			to:          "192.168.1.2",
			toHW:        "11:22:33:44:55:66",
			srcPort:     65535,
			dstPort:     65535,
			expectError: false,
			expectIPv6:  false,
		},
		{
			name:        "Min port number",
			from:        "192.168.1.1",
			fromHW:      "aa:bb:cc:dd:ee:ff",
			to:          "192.168.1.2",
			toHW:        "11:22:33:44:55:66",
			srcPort:     1,
			dstPort:     1,
			expectError: false,
			expectIPv6:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := net.ParseIP(tt.from)
			fromHW, _ := net.ParseMAC(tt.fromHW)
			to := net.ParseIP(tt.to)
			toHW, _ := net.ParseMAC(tt.toHW)

			err, data := NewTCPSyn(from, fromHW, to, toHW, tt.srcPort, tt.dstPort)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				if len(data) == 0 {
					t.Error("Expected data but got empty")
				}

				// Parse the packet to verify structure
				packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

				// Check Ethernet layer
				if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
					eth := ethLayer.(*layers.Ethernet)
					if !bytes.Equal(eth.SrcMAC, fromHW) {
						t.Errorf("Ethernet SrcMAC = %v, want %v", eth.SrcMAC, fromHW)
					}
					if !bytes.Equal(eth.DstMAC, toHW) {
						t.Errorf("Ethernet DstMAC = %v, want %v", eth.DstMAC, toHW)
					}
					expectedType := layers.EthernetTypeIPv4
					if tt.expectIPv6 {
						expectedType = layers.EthernetTypeIPv6
					}
					if eth.EthernetType != expectedType {
						t.Errorf("EthernetType = %v, want %v", eth.EthernetType, expectedType)
					}
				} else {
					t.Error("Packet missing Ethernet layer")
				}

				// Check IP layer
				if tt.expectIPv6 {
					if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
						ip := ipLayer.(*layers.IPv6)
						if !ip.SrcIP.Equal(from) {
							t.Errorf("IPv6 SrcIP = %v, want %v", ip.SrcIP, from)
						}
						if !ip.DstIP.Equal(to) {
							t.Errorf("IPv6 DstIP = %v, want %v", ip.DstIP, to)
						}
						if ip.HopLimit != 64 {
							t.Errorf("IPv6 HopLimit = %d, want 64", ip.HopLimit)
						}
						if ip.NextHeader != layers.IPProtocolTCP {
							t.Errorf("IPv6 NextHeader = %v, want %v", ip.NextHeader, layers.IPProtocolTCP)
						}
					} else {
						t.Error("Packet missing IPv6 layer")
					}
				} else {
					if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
						ip := ipLayer.(*layers.IPv4)
						if !ip.SrcIP.Equal(from) {
							t.Errorf("IPv4 SrcIP = %v, want %v", ip.SrcIP, from)
						}
						if !ip.DstIP.Equal(to) {
							t.Errorf("IPv4 DstIP = %v, want %v", ip.DstIP, to)
						}
						if ip.TTL != 64 {
							t.Errorf("IPv4 TTL = %d, want 64", ip.TTL)
						}
						if ip.Protocol != layers.IPProtocolTCP {
							t.Errorf("IPv4 Protocol = %v, want %v", ip.Protocol, layers.IPProtocolTCP)
						}
					} else {
						t.Error("Packet missing IPv4 layer")
					}
				}

				// Check TCP layer
				if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
					tcp := tcpLayer.(*layers.TCP)
					if tcp.SrcPort != layers.TCPPort(tt.srcPort) {
						t.Errorf("TCP SrcPort = %d, want %d", tcp.SrcPort, tt.srcPort)
					}
					if tcp.DstPort != layers.TCPPort(tt.dstPort) {
						t.Errorf("TCP DstPort = %d, want %d", tcp.DstPort, tt.dstPort)
					}
					if !tcp.SYN {
						t.Error("TCP SYN flag not set")
					}
					// Verify other flags are not set
					if tcp.ACK || tcp.FIN || tcp.RST || tcp.PSH || tcp.URG {
						t.Error("TCP has unexpected flags set")
					}
				} else {
					t.Error("Packet missing TCP layer")
				}
			}
		})
	}
}

func TestNewTCPSynWithNilValues(t *testing.T) {
	// Test with nil IPs - should return an error
	err, data := NewTCPSyn(nil, nil, nil, nil, 12345, 80)
	if err == nil {
		t.Error("Expected error with nil values, but got none")
	}
	if len(data) != 0 {
		t.Error("Expected no data with nil values")
	}
}

func TestNewTCPSynChecksumComputation(t *testing.T) {
	// Test that checksums are computed correctly for both IPv4 and IPv6
	testCases := []struct {
		name   string
		from   string
		to     string
		isIPv6 bool
	}{
		{
			name:   "IPv4 checksum",
			from:   "192.168.1.1",
			to:     "192.168.1.2",
			isIPv6: false,
		},
		{
			name:   "IPv6 checksum",
			from:   "2001:db8::1",
			to:     "2001:db8::2",
			isIPv6: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			from := net.ParseIP(tc.from)
			to := net.ParseIP(tc.to)
			fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
			toHW, _ := net.ParseMAC("11:22:33:44:55:66")

			err, data := NewTCPSyn(from, fromHW, to, toHW, 12345, 80)
			if err != nil {
				t.Fatalf("Failed to create TCP SYN: %v", err)
			}

			// Parse the packet
			packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

			// Verify TCP checksum is non-zero (computed)
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				if tcp.Checksum == 0 {
					t.Error("TCP checksum was not computed")
				}
			} else {
				t.Error("TCP layer not found")
			}

			// For IPv4, also check IP checksum
			if !tc.isIPv6 {
				if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
					ip := ipLayer.(*layers.IPv4)
					if ip.Checksum == 0 {
						t.Error("IPv4 checksum was not computed")
					}
				}
			}
		})
	}
}

func TestNewTCPSynPortRange(t *testing.T) {
	// Test various port numbers including edge cases
	portTests := []struct {
		srcPort int
		dstPort int
	}{
		{0, 0},         // Minimum possible (though 0 is typically reserved)
		{1, 1},         // Minimum valid
		{80, 443},      // Common ports
		{1024, 1025},   // First non-privileged ports
		{32768, 32769}, // Common ephemeral port range start
		{65534, 65535}, // Maximum ports
	}

	from := net.ParseIP("192.168.1.1")
	to := net.ParseIP("192.168.1.2")
	fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	toHW, _ := net.ParseMAC("11:22:33:44:55:66")

	for _, pt := range portTests {
		err, data := NewTCPSyn(from, fromHW, to, toHW, pt.srcPort, pt.dstPort)
		if err != nil {
			t.Errorf("Failed with ports %d->%d: %v", pt.srcPort, pt.dstPort, err)
			continue
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp := tcpLayer.(*layers.TCP)
			if tcp.SrcPort != layers.TCPPort(pt.srcPort) {
				t.Errorf("TCP SrcPort = %d, want %d", tcp.SrcPort, pt.srcPort)
			}
			if tcp.DstPort != layers.TCPPort(pt.dstPort) {
				t.Errorf("TCP DstPort = %d, want %d", tcp.DstPort, pt.dstPort)
			}
		}
	}
}

// Benchmarks
func BenchmarkNewTCPSynIPv4(b *testing.B) {
	from := net.ParseIP("192.168.1.1")
	to := net.ParseIP("192.168.1.2")
	fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	toHW, _ := net.ParseMAC("11:22:33:44:55:66")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTCPSyn(from, fromHW, to, toHW, 12345, 80)
	}
}

func BenchmarkNewTCPSynIPv6(b *testing.B) {
	from := net.ParseIP("2001:db8::1")
	to := net.ParseIP("2001:db8::2")
	fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	toHW, _ := net.ParseMAC("11:22:33:44:55:66")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTCPSyn(from, fromHW, to, toHW, 12345, 80)
	}
}
