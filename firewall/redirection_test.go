package firewall

import (
	"testing"
)

func TestNewRedirection(t *testing.T) {
	iface := "eth0"
	proto := "tcp"
	portFrom := 8080
	addrTo := "192.168.1.100"
	portTo := 9090

	r := NewRedirection(iface, proto, portFrom, addrTo, portTo)

	if r == nil {
		t.Fatal("NewRedirection returned nil")
	}

	if r.Interface != iface {
		t.Errorf("expected Interface %s, got %s", iface, r.Interface)
	}

	if r.Protocol != proto {
		t.Errorf("expected Protocol %s, got %s", proto, r.Protocol)
	}

	if r.SrcAddress != "" {
		t.Errorf("expected empty SrcAddress, got %s", r.SrcAddress)
	}

	if r.SrcPort != portFrom {
		t.Errorf("expected SrcPort %d, got %d", portFrom, r.SrcPort)
	}

	if r.DstAddress != addrTo {
		t.Errorf("expected DstAddress %s, got %s", addrTo, r.DstAddress)
	}

	if r.DstPort != portTo {
		t.Errorf("expected DstPort %d, got %d", portTo, r.DstPort)
	}
}

func TestRedirectionString(t *testing.T) {
	tests := []struct {
		name string
		r    Redirection
		want string
	}{
		{
			name: "basic redirection",
			r: Redirection{
				Interface:  "eth0",
				Protocol:   "tcp",
				SrcAddress: "",
				SrcPort:    8080,
				DstAddress: "192.168.1.100",
				DstPort:    9090,
			},
			want: "[eth0] (tcp) :8080 -> 192.168.1.100:9090",
		},
		{
			name: "with source address",
			r: Redirection{
				Interface:  "wlan0",
				Protocol:   "udp",
				SrcAddress: "192.168.1.50",
				SrcPort:    53,
				DstAddress: "8.8.8.8",
				DstPort:    53,
			},
			want: "[wlan0] (udp) 192.168.1.50:53 -> 8.8.8.8:53",
		},
		{
			name: "localhost redirection",
			r: Redirection{
				Interface:  "lo",
				Protocol:   "tcp",
				SrcAddress: "127.0.0.1",
				SrcPort:    80,
				DstAddress: "127.0.0.1",
				DstPort:    8080,
			},
			want: "[lo] (tcp) 127.0.0.1:80 -> 127.0.0.1:8080",
		},
		{
			name: "high port numbers",
			r: Redirection{
				Interface:  "eth1",
				Protocol:   "tcp",
				SrcAddress: "",
				SrcPort:    65535,
				DstAddress: "10.0.0.1",
				DstPort:    65534,
			},
			want: "[eth1] (tcp) :65535 -> 10.0.0.1:65534",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewRedirectionVariousProtocols(t *testing.T) {
	protocols := []string{"tcp", "udp", "icmp", "any"}

	for _, proto := range protocols {
		t.Run(proto, func(t *testing.T) {
			r := NewRedirection("eth0", proto, 1234, "10.0.0.1", 5678)
			if r.Protocol != proto {
				t.Errorf("expected protocol %s, got %s", proto, r.Protocol)
			}
		})
	}
}

func TestNewRedirectionVariousInterfaces(t *testing.T) {
	interfaces := []string{"eth0", "wlan0", "lo", "docker0", "br0", "tun0"}

	for _, iface := range interfaces {
		t.Run(iface, func(t *testing.T) {
			r := NewRedirection(iface, "tcp", 80, "192.168.1.1", 8080)
			if r.Interface != iface {
				t.Errorf("expected interface %s, got %s", iface, r.Interface)
			}
		})
	}
}

func TestRedirectionStringEmptyFields(t *testing.T) {
	tests := []struct {
		name string
		r    Redirection
		want string
	}{
		{
			name: "empty interface",
			r: Redirection{
				Interface:  "",
				Protocol:   "tcp",
				SrcAddress: "",
				SrcPort:    80,
				DstAddress: "192.168.1.1",
				DstPort:    8080,
			},
			want: "[] (tcp) :80 -> 192.168.1.1:8080",
		},
		{
			name: "empty protocol",
			r: Redirection{
				Interface:  "eth0",
				Protocol:   "",
				SrcAddress: "",
				SrcPort:    80,
				DstAddress: "192.168.1.1",
				DstPort:    8080,
			},
			want: "[eth0] () :80 -> 192.168.1.1:8080",
		},
		{
			name: "empty destination",
			r: Redirection{
				Interface:  "eth0",
				Protocol:   "tcp",
				SrcAddress: "",
				SrcPort:    80,
				DstAddress: "",
				DstPort:    8080,
			},
			want: "[eth0] (tcp) :80 -> :8080",
		},
		{
			name: "all empty strings",
			r: Redirection{
				Interface:  "",
				Protocol:   "",
				SrcAddress: "",
				SrcPort:    0,
				DstAddress: "",
				DstPort:    0,
			},
			want: "[] () :0 -> :0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRedirectionStructCopy(t *testing.T) {
	// Test that Redirection can be safely copied
	original := NewRedirection("eth0", "tcp", 80, "192.168.1.1", 8080)
	original.SrcAddress = "10.0.0.1"

	// Create a copy
	copy := *original

	// Modify the copy
	copy.Interface = "wlan0"
	copy.SrcPort = 443

	// Verify original is unchanged
	if original.Interface != "eth0" {
		t.Error("original Interface was modified")
	}
	if original.SrcPort != 80 {
		t.Error("original SrcPort was modified")
	}

	// Verify copy has new values
	if copy.Interface != "wlan0" {
		t.Error("copy Interface was not set correctly")
	}
	if copy.SrcPort != 443 {
		t.Error("copy SrcPort was not set correctly")
	}
}

func BenchmarkNewRedirection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRedirection("eth0", "tcp", 80, "192.168.1.1", 8080)
	}
}

func BenchmarkRedirectionString(b *testing.B) {
	r := Redirection{
		Interface:  "eth0",
		Protocol:   "tcp",
		SrcAddress: "192.168.1.50",
		SrcPort:    8080,
		DstAddress: "192.168.1.100",
		DstPort:    9090,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkRedirectionStringEmpty(b *testing.B) {
	r := Redirection{
		Interface:  "eth0",
		Protocol:   "tcp",
		SrcAddress: "",
		SrcPort:    8080,
		DstAddress: "192.168.1.100",
		DstPort:    9090,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}
