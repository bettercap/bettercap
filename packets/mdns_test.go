package packets

import (
	"bytes"
	"net"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func TestMDNSConstants(t *testing.T) {
	if MDNSPort != 5353 {
		t.Errorf("MDNSPort = %d, want 5353", MDNSPort)
	}

	expectedMac := net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0xfb}
	if !bytes.Equal(MDNSDestMac, expectedMac) {
		t.Errorf("MDNSDestMac = %v, want %v", MDNSDestMac, expectedMac)
	}

	expectedIP := net.ParseIP("224.0.0.251")
	if !MDNSDestIP.Equal(expectedIP) {
		t.Errorf("MDNSDestIP = %v, want %v", MDNSDestIP, expectedIP)
	}
}

func TestNewMDNSProbe(t *testing.T) {
	from := net.ParseIP("192.168.1.100")
	fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

	err, data := NewMDNSProbe(from, fromHW)
	if err != nil {
		t.Errorf("NewMDNSProbe() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("NewMDNSProbe() returned empty data")
	}

	// Parse the packet to verify structure
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	// Check Ethernet layer
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		if !bytes.Equal(eth.SrcMAC, fromHW) {
			t.Errorf("Ethernet SrcMAC = %v, want %v", eth.SrcMAC, fromHW)
		}
		if !bytes.Equal(eth.DstMAC, MDNSDestMac) {
			t.Errorf("Ethernet DstMAC = %v, want %v", eth.DstMAC, MDNSDestMac)
		}
	} else {
		t.Error("Packet missing Ethernet layer")
	}

	// Check IPv4 layer
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip := ipLayer.(*layers.IPv4)
		if !ip.SrcIP.Equal(from) {
			t.Errorf("IPv4 SrcIP = %v, want %v", ip.SrcIP, from)
		}
		if !ip.DstIP.Equal(MDNSDestIP) {
			t.Errorf("IPv4 DstIP = %v, want %v", ip.DstIP, MDNSDestIP)
		}
	} else {
		t.Error("Packet missing IPv4 layer")
	}

	// Check UDP layer
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		if udp.DstPort != MDNSPort {
			t.Errorf("UDP DstPort = %d, want %d", udp.DstPort, MDNSPort)
		}
	} else {
		t.Error("Packet missing UDP layer")
	}

	// The DNS layer is carried as payload in UDP, not a separate layer
	// So we check the UDP payload instead
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		// Verify that the UDP payload contains DNS data
		if len(udp.Payload) == 0 {
			t.Error("UDP payload is empty (should contain DNS data)")
		}
	}
}

func TestMDNSGetMeta(t *testing.T) {
	// Create a mock MDNS packet with various record types
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       MDNSDestMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP("192.168.1.100"),
		DstIP:    MDNSDestIP,
	}

	udp := layers.UDP{
		SrcPort: MDNSPort,
		DstPort: MDNSPort,
	}

	dns := layers.DNS{
		ID:     1,
		QR:     true,
		OpCode: layers.DNSOpCodeQuery,
		Answers: []layers.DNSResourceRecord{
			{
				Name:  []byte("test.local"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
				IP:    net.ParseIP("192.168.1.100"),
			},
			{
				Name:  []byte("test.local"),
				Type:  layers.DNSTypeTXT,
				Class: layers.DNSClassIN,
				TXTs:  [][]byte{[]byte("model=Test Device"), []byte("version=1.0")},
			},
		},
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &udp, &dns)
	if err != nil {
		t.Fatalf("Failed to serialize packet: %v", err)
	}

	packet := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	meta := MDNSGetMeta(packet)
	if meta == nil {
		t.Fatal("MDNSGetMeta() returned nil")
	}

	// TXT records are extracted correctly

	if model, ok := meta["mdns:model"]; !ok || model != "Test Device" {
		t.Errorf("Expected model 'Test Device', got '%v'", model)
	}

	if version, ok := meta["mdns:version"]; !ok || version != "1.0" {
		t.Errorf("Expected version '1.0', got '%v'", version)
	}
}

func TestMDNSGetMetaNonMDNS(t *testing.T) {
	// Create a non-MDNS UDP packet
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP("192.168.1.100"),
		DstIP:    net.ParseIP("192.168.1.200"),
	}

	udp := layers.UDP{
		SrcPort: 12345,
		DstPort: 80,
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &udp)
	if err != nil {
		t.Fatalf("Failed to serialize packet: %v", err)
	}

	packet := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	meta := MDNSGetMeta(packet)
	if meta != nil {
		t.Error("MDNSGetMeta() should return nil for non-MDNS packet")
	}
}

func TestMDNSGetMetaInvalidDNS(t *testing.T) {
	// Create MDNS packet with invalid DNS payload
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       MDNSDestMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP("192.168.1.100"),
		DstIP:    MDNSDestIP,
	}

	udp := layers.UDP{
		SrcPort: MDNSPort,
		DstPort: MDNSPort,
	}

	udp.SetNetworkLayerForChecksum(&ip4)
	udp.Payload = []byte{0x00, 0x01, 0x02, 0x03} // Invalid DNS data

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &udp)
	if err != nil {
		t.Fatalf("Failed to serialize packet: %v", err)
	}

	packet := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	meta := MDNSGetMeta(packet)
	if meta != nil {
		t.Error("MDNSGetMeta() should return nil for invalid DNS data")
	}
}

func TestMDNSGetMetaRecovery(t *testing.T) {
	// Test that panic recovery works
	defer func() {
		if r := recover(); r != nil {
			t.Error("MDNSGetMeta should not panic")
		}
	}()

	// Create a minimal packet that might cause issues
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	meta := MDNSGetMeta(packet)
	if meta != nil {
		t.Error("MDNSGetMeta() should return nil for invalid packet")
	}
}

func TestMDNSGetMetaWithAdditionals(t *testing.T) {
	// Create a mock MDNS packet with additional records
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       MDNSDestMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP("192.168.1.100"),
		DstIP:    MDNSDestIP,
	}

	udp := layers.UDP{
		SrcPort: MDNSPort,
		DstPort: MDNSPort,
	}

	dns := layers.DNS{
		ID:     1,
		QR:     true,
		OpCode: layers.DNSOpCodeQuery,
		Additionals: []layers.DNSResourceRecord{
			{
				Name:  []byte("additional.local"),
				Type:  layers.DNSTypeAAAA,
				Class: layers.DNSClassIN,
				IP:    net.ParseIP("fe80::1"),
			},
		},
		Authorities: []layers.DNSResourceRecord{
			{
				Name:  []byte("authority.local"),
				Type:  layers.DNSTypePTR,
				Class: layers.DNSClassIN,
			},
		},
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &udp, &dns)
	if err != nil {
		t.Fatalf("Failed to serialize packet: %v", err)
	}

	packet := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	meta := MDNSGetMeta(packet)
	if meta == nil {
		t.Fatal("MDNSGetMeta() returned nil")
	}

	if hostname, ok := meta["mdns:hostname"]; !ok || hostname != "additional.local" {
		t.Errorf("Expected hostname 'additional.local', got '%v'", hostname)
	}
}

// Benchmarks
func BenchmarkNewMDNSProbe(b *testing.B) {
	from := net.ParseIP("192.168.1.100")
	fromHW, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewMDNSProbe(from, fromHW)
	}
}

func BenchmarkMDNSGetMeta(b *testing.B) {
	// Create a sample MDNS packet
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		DstMAC:       MDNSDestMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP("192.168.1.100"),
		DstIP:    MDNSDestIP,
	}

	udp := layers.UDP{
		SrcPort: MDNSPort,
		DstPort: MDNSPort,
	}

	dns := layers.DNS{
		ID:     1,
		QR:     true,
		OpCode: layers.DNSOpCodeQuery,
		Answers: []layers.DNSResourceRecord{
			{
				Name:  []byte("test.local"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
				IP:    net.ParseIP("192.168.1.100"),
			},
		},
	}

	udp.SetNetworkLayerForChecksum(&ip4)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	gopacket.SerializeLayers(buf, opts, &eth, &ip4, &udp, &dns)
	packet := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MDNSGetMeta(packet)
	}
}
