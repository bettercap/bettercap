package firewall

import (
	"reflect"
	"testing"
)

func TestGetCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		r       *Redirection
		enabled bool
		want    []string
	}{
		{
			name: "DNAT without source address",
			r: &Redirection{
				Interface:  "eth0",
				Protocol:   "tcp",
				SrcAddress: "",
				SrcPort:    8080,
				DstAddress: "1.2.3.4",
				DstPort:    9090,
			},
			enabled: true,
			want: []string{
				"-t", "nat",
				"-A", "PREROUTING",
				"-i", "eth0",
				"-p", "tcp",
				"--dport", "8080",
				"-j", "DNAT",
				"--to", "1.2.3.4:9090",
			},
		},
		{
			name: "DNAT with source address",
			r: &Redirection{
				Interface:  "eth0",
				Protocol:   "tcp",
				SrcAddress: "10.0.0.1",
				SrcPort:    8080,
				DstAddress: "1.2.3.4",
				DstPort:    9090,
			},
			enabled: true,
			want: []string{
				"-t", "nat",
				"-A", "PREROUTING",
				"-i", "eth0",
				"-p", "tcp",
				"-d", "10.0.0.1",
				"--dport", "8080",
				"-j", "DNAT",
				"--to", "1.2.3.4:9090",
			},
		},
		{
			name: "drop without match, enabled",
			r: &Redirection{
				Interface: "eth0",
				Protocol:  "udp",
				SrcPort:   53,
				Drop:      true,
			},
			enabled: true,
			want: []string{
				"-I", "FORWARD",
				"-i", "eth0",
				"-p", "udp",
				"--dport", "53",
				"-j", "DROP",
			},
		},
		{
			name: "drop with match, enabled",
			r: &Redirection{
				Interface: "eth0",
				Protocol:  "udp",
				SrcPort:   53,
				Drop:      true,
				DropMatch: "|06676f6f676c6503636f6d00|",
			},
			enabled: true,
			want: []string{
				"-I", "FORWARD",
				"-i", "eth0",
				"-p", "udp",
				"--dport", "53",
				"-m", "string", "--algo", "bm", "--hex-string", "|06676f6f676c6503636f6d00|",
				"-j", "DROP",
			},
		},
		{
			name: "drop disabled uses delete action",
			r: &Redirection{
				Interface: "eth0",
				Protocol:  "udp",
				SrcPort:   53,
				Drop:      true,
				DropMatch: "|06676f6f676c6503636f6d00|",
			},
			enabled: false,
			want: []string{
				"-D", "FORWARD",
				"-i", "eth0",
				"-p", "udp",
				"--dport", "53",
				"-m", "string", "--algo", "bm", "--hex-string", "|06676f6f676c6503636f6d00|",
				"-j", "DROP",
			},
		},
	}

	f := &LinuxFirewall{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.getCommandLine(tt.r, tt.enabled)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCommandLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
