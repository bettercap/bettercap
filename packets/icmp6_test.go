package packets

import (
	"bytes"
	"net"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func TestICMP6Constants(t *testing.T) {
	// Test the multicast constants
	expectedMAC := net.HardwareAddr([]byte{0x33, 0x33, 0x00, 0x00, 0x00, 0x01})
	if !bytes.Equal(macIpv6Multicast, expectedMAC) {
		t.Errorf("macIpv6Multicast = %v, want %v", macIpv6Multicast, expectedMAC)
	}

	expectedIP := net.ParseIP("ff02::1")
	if !ipv6Multicast.Equal(expectedIP) {
		t.Errorf("ipv6Multicast = %v, want %v", ipv6Multicast, expectedIP)
	}
}

func TestICMP6NeighborAdvertisement(t *testing.T) {
	srcHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	srcIP := net.ParseIP("fe80::1")
	dstHW, _ := net.ParseMAC("11:22:33:44:55:66")
	dstIP := net.ParseIP("fe80::2")
	routerIP := net.ParseIP("fe80::3")

	err, data := ICMP6NeighborAdvertisement(srcHW, srcIP, dstHW, dstIP, routerIP)
	if err != nil {
		t.Fatalf("ICMP6NeighborAdvertisement() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ICMP6NeighborAdvertisement() returned empty data")
	}

	// Parse the packet to verify structure
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	// Check Ethernet layer
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		if !bytes.Equal(eth.SrcMAC, srcHW) {
			t.Errorf("Ethernet SrcMAC = %v, want %v", eth.SrcMAC, srcHW)
		}
		if !bytes.Equal(eth.DstMAC, dstHW) {
			t.Errorf("Ethernet DstMAC = %v, want %v", eth.DstMAC, dstHW)
		}
		if eth.EthernetType != layers.EthernetTypeIPv6 {
			t.Errorf("EthernetType = %v, want %v", eth.EthernetType, layers.EthernetTypeIPv6)
		}
	} else {
		t.Error("Packet missing Ethernet layer")
	}

	// Check IPv6 layer
	if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ip := ipLayer.(*layers.IPv6)
		if !ip.SrcIP.Equal(srcIP) {
			t.Errorf("IPv6 SrcIP = %v, want %v", ip.SrcIP, srcIP)
		}
		if !ip.DstIP.Equal(dstIP) {
			t.Errorf("IPv6 DstIP = %v, want %v", ip.DstIP, dstIP)
		}
		if ip.HopLimit != 255 {
			t.Errorf("IPv6 HopLimit = %d, want 255", ip.HopLimit)
		}
		if ip.NextHeader != layers.IPProtocolICMPv6 {
			t.Errorf("IPv6 NextHeader = %v, want %v", ip.NextHeader, layers.IPProtocolICMPv6)
		}
	} else {
		t.Error("Packet missing IPv6 layer")
	}

	// Check ICMPv6 layer
	if icmpLayer := packet.Layer(layers.LayerTypeICMPv6); icmpLayer != nil {
		icmp := icmpLayer.(*layers.ICMPv6)
		expectedType := uint8(layers.ICMPv6TypeNeighborAdvertisement)
		if icmp.TypeCode.Type() != expectedType {
			t.Errorf("ICMPv6 Type = %v, want %v", icmp.TypeCode.Type(), expectedType)
		}
	} else {
		t.Error("Packet missing ICMPv6 layer")
	}

	// Check ICMPv6NeighborAdvertisement layer
	if naLayer := packet.Layer(layers.LayerTypeICMPv6NeighborAdvertisement); naLayer != nil {
		na := naLayer.(*layers.ICMPv6NeighborAdvertisement)
		if !na.TargetAddress.Equal(routerIP) {
			t.Errorf("TargetAddress = %v, want %v", na.TargetAddress, routerIP)
		}
		// Check flags (solicited && override)
		expectedFlags := uint8(0x20 | 0x40)
		if na.Flags != expectedFlags {
			t.Errorf("Flags = %x, want %x", na.Flags, expectedFlags)
		}
		// Check options
		if len(na.Options) != 1 {
			t.Errorf("Options count = %d, want 1", len(na.Options))
		} else {
			opt := na.Options[0]
			if opt.Type != layers.ICMPv6OptTargetAddress {
				t.Errorf("Option Type = %v, want %v", opt.Type, layers.ICMPv6OptTargetAddress)
			}
			if !bytes.Equal(opt.Data, srcHW) {
				t.Errorf("Option Data = %v, want %v", opt.Data, srcHW)
			}
		}
	} else {
		t.Error("Packet missing ICMPv6NeighborAdvertisement layer")
	}
}

func TestICMP6RouterAdvertisement(t *testing.T) {
	ip := net.ParseIP("fe80::1")
	hw, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	prefix := "2001:db8::"
	prefixLength := uint8(64)
	routerLifetime := uint16(1800)

	err, data := ICMP6RouterAdvertisement(ip, hw, prefix, prefixLength, routerLifetime)
	if err != nil {
		t.Fatalf("ICMP6RouterAdvertisement() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ICMP6RouterAdvertisement() returned empty data")
	}

	// Parse the packet to verify structure
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	// Check Ethernet layer
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		if !bytes.Equal(eth.SrcMAC, hw) {
			t.Errorf("Ethernet SrcMAC = %v, want %v", eth.SrcMAC, hw)
		}
		if !bytes.Equal(eth.DstMAC, macIpv6Multicast) {
			t.Errorf("Ethernet DstMAC = %v, want %v", eth.DstMAC, macIpv6Multicast)
		}
		if eth.EthernetType != layers.EthernetTypeIPv6 {
			t.Errorf("EthernetType = %v, want %v", eth.EthernetType, layers.EthernetTypeIPv6)
		}
	} else {
		t.Error("Packet missing Ethernet layer")
	}

	// Check IPv6 layer
	if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ip6 := ipLayer.(*layers.IPv6)
		if !ip6.SrcIP.Equal(ip) {
			t.Errorf("IPv6 SrcIP = %v, want %v", ip6.SrcIP, ip)
		}
		if !ip6.DstIP.Equal(ipv6Multicast) {
			t.Errorf("IPv6 DstIP = %v, want %v", ip6.DstIP, ipv6Multicast)
		}
		if ip6.HopLimit != 255 {
			t.Errorf("IPv6 HopLimit = %d, want 255", ip6.HopLimit)
		}
		if ip6.NextHeader != layers.IPProtocolICMPv6 {
			t.Errorf("IPv6 NextHeader = %v, want %v", ip6.NextHeader, layers.IPProtocolICMPv6)
		}
		if ip6.TrafficClass != 224 {
			t.Errorf("IPv6 TrafficClass = %d, want 224", ip6.TrafficClass)
		}
	} else {
		t.Error("Packet missing IPv6 layer")
	}

	// Check ICMPv6 layer
	if icmpLayer := packet.Layer(layers.LayerTypeICMPv6); icmpLayer != nil {
		icmp := icmpLayer.(*layers.ICMPv6)
		expectedType := uint8(layers.ICMPv6TypeRouterAdvertisement)
		if icmp.TypeCode.Type() != expectedType {
			t.Errorf("ICMPv6 Type = %v, want %v", icmp.TypeCode.Type(), expectedType)
		}
	} else {
		t.Error("Packet missing ICMPv6 layer")
	}

	// Check ICMPv6RouterAdvertisement layer
	if raLayer := packet.Layer(layers.LayerTypeICMPv6RouterAdvertisement); raLayer != nil {
		ra := raLayer.(*layers.ICMPv6RouterAdvertisement)
		if ra.HopLimit != 255 {
			t.Errorf("HopLimit = %d, want 255", ra.HopLimit)
		}
		if ra.Flags != 0x08 {
			t.Errorf("Flags = %x, want 0x08", ra.Flags)
		}
		if ra.RouterLifetime != routerLifetime {
			t.Errorf("RouterLifetime = %d, want %d", ra.RouterLifetime, routerLifetime)
		}
		// Check options - the actual order from the code is SourceAddress, MTU, PrefixInfo
		if len(ra.Options) != 3 {
			t.Errorf("Options count = %d, want 3", len(ra.Options))
		} else {
			// Find each option type
			hasSourceAddr := false
			hasMTU := false
			hasPrefixInfo := false

			for _, opt := range ra.Options {
				switch opt.Type {
				case layers.ICMPv6OptSourceAddress:
					hasSourceAddr = true
					if !bytes.Equal(opt.Data, hw) {
						t.Errorf("SourceAddress option data = %v, want %v", opt.Data, hw)
					}
				case layers.ICMPv6OptMTU:
					hasMTU = true
					expectedMTU := []byte{0x00, 0x00, 0x00, 0x00, 0x05, 0xdc} // 1500
					if !bytes.Equal(opt.Data, expectedMTU) {
						t.Errorf("MTU option data = %v, want %v", opt.Data, expectedMTU)
					}
				case layers.ICMPv6OptPrefixInfo:
					hasPrefixInfo = true
					// Verify prefix length is in the data
					if len(opt.Data) > 0 && opt.Data[0] != prefixLength {
						t.Errorf("PrefixInfo prefix length = %d, want %d", opt.Data[0], prefixLength)
					}
				}
			}

			if !hasSourceAddr {
				t.Error("Missing SourceAddress option")
			}
			if !hasMTU {
				t.Error("Missing MTU option")
			}
			if !hasPrefixInfo {
				t.Error("Missing PrefixInfo option")
			}
		}
	} else {
		t.Error("Packet missing ICMPv6RouterAdvertisement layer")
	}
}

func TestICMP6NeighborAdvertisementWithNilValues(t *testing.T) {
	// Test with nil values - function should handle gracefully
	err, data := ICMP6NeighborAdvertisement(nil, nil, nil, nil, nil)

	// The function likely returns an error or empty data with nil inputs
	if err == nil && len(data) > 0 {
		t.Error("Expected error or empty data with nil values")
	}
}

func TestICMP6RouterAdvertisementWithNilValues(t *testing.T) {
	// Test with nil values - function should handle gracefully
	err, data := ICMP6RouterAdvertisement(nil, nil, "", 0, 0)

	// The function likely returns an error or empty data with nil inputs
	if err == nil && len(data) > 0 {
		t.Error("Expected error or empty data with nil values")
	}
}

func TestICMP6RouterAdvertisementVariousInputs(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		hw             string
		prefix         string
		prefixLength   uint8
		routerLifetime uint16
		shouldError    bool
	}{
		{
			name:           "valid input",
			ip:             "fe80::1",
			hw:             "aa:bb:cc:dd:ee:ff",
			prefix:         "2001:db8::",
			prefixLength:   64,
			routerLifetime: 1800,
			shouldError:    false,
		},
		{
			name:           "zero router lifetime",
			ip:             "fe80::1",
			hw:             "aa:bb:cc:dd:ee:ff",
			prefix:         "2001:db8::",
			prefixLength:   64,
			routerLifetime: 0,
			shouldError:    false,
		},
		{
			name:           "max prefix length",
			ip:             "fe80::1",
			hw:             "aa:bb:cc:dd:ee:ff",
			prefix:         "2001:db8::",
			prefixLength:   128,
			routerLifetime: 1800,
			shouldError:    false,
		},
		{
			name:           "max router lifetime",
			ip:             "fe80::1",
			hw:             "aa:bb:cc:dd:ee:ff",
			prefix:         "2001:db8::",
			prefixLength:   64,
			routerLifetime: 65535,
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			hw, _ := net.ParseMAC(tt.hw)

			err, data := ICMP6RouterAdvertisement(ip, hw, tt.prefix, tt.prefixLength, tt.routerLifetime)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldError && len(data) == 0 {
				t.Error("Expected data but got empty")
			}
		})
	}
}

func TestICMP6NeighborAdvertisementVariousInputs(t *testing.T) {
	tests := []struct {
		name        string
		srcHW       string
		srcIP       string
		dstHW       string
		dstIP       string
		routerIP    string
		shouldError bool
	}{
		{
			name:        "valid IPv6 link-local",
			srcHW:       "aa:bb:cc:dd:ee:ff",
			srcIP:       "fe80::1",
			dstHW:       "11:22:33:44:55:66",
			dstIP:       "fe80::2",
			routerIP:    "fe80::3",
			shouldError: false,
		},
		{
			name:        "valid IPv6 global",
			srcHW:       "aa:bb:cc:dd:ee:ff",
			srcIP:       "2001:db8::1",
			dstHW:       "11:22:33:44:55:66",
			dstIP:       "2001:db8::2",
			routerIP:    "2001:db8::3",
			shouldError: false,
		},
		{
			name:        "broadcast MAC",
			srcHW:       "ff:ff:ff:ff:ff:ff",
			srcIP:       "fe80::1",
			dstHW:       "ff:ff:ff:ff:ff:ff",
			dstIP:       "fe80::2",
			routerIP:    "fe80::3",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcHW, _ := net.ParseMAC(tt.srcHW)
			srcIP := net.ParseIP(tt.srcIP)
			dstHW, _ := net.ParseMAC(tt.dstHW)
			dstIP := net.ParseIP(tt.dstIP)
			routerIP := net.ParseIP(tt.routerIP)

			err, data := ICMP6NeighborAdvertisement(srcHW, srcIP, dstHW, dstIP, routerIP)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldError && len(data) == 0 {
				t.Error("Expected data but got empty")
			}
		})
	}
}

// Benchmarks
func BenchmarkICMP6NeighborAdvertisement(b *testing.B) {
	srcHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	srcIP := net.ParseIP("fe80::1")
	dstHW, _ := net.ParseMAC("11:22:33:44:55:66")
	dstIP := net.ParseIP("fe80::2")
	routerIP := net.ParseIP("fe80::3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ICMP6NeighborAdvertisement(srcHW, srcIP, dstHW, dstIP, routerIP)
	}
}

func BenchmarkICMP6RouterAdvertisement(b *testing.B) {
	ip := net.ParseIP("fe80::1")
	hw, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	prefix := "2001:db8::"
	prefixLength := uint8(64)
	routerLifetime := uint16(1800)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ICMP6RouterAdvertisement(ip, hw, prefix, prefixLength, routerLifetime)
	}
}
