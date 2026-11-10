package dns_spoof

import (
	"strings"
	"testing"
)

func TestDnsNameToHexString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOk bool
	}{
		{"simple", "google.com", "|06676f6f676c6503636f6d00|", true},
		{"trailing dot", "google.com.", "|06676f6f676c6503636f6d00|", true},
		{"uppercase", "GOOGLE.COM", "|06676f6f676c6503636f6d00|", true},
		{"leading dot", ".google.com", "|06676f6f676c6503636f6d00|", true},
		{"wildcard", "*.google.com", "", false},
		{"empty", "", "", false},
		{"oversized label", strings.Repeat("a", 64) + ".com", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := dnsNameToHexString(tt.input)
			if ok != tt.wantOk {
				t.Errorf("dnsNameToHexString(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("dnsNameToHexString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
