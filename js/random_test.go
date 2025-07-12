package js

import (
	"net"
	"regexp"
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	r := randomPackage{}

	tests := []struct {
		name    string
		size    int
		charset string
	}{
		{
			name:    "alphanumeric",
			size:    10,
			charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:    "numbers only",
			size:    20,
			charset: "0123456789",
		},
		{
			name:    "lowercase letters",
			size:    15,
			charset: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:    "uppercase letters",
			size:    8,
			charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:    "special characters",
			size:    12,
			charset: "!@#$%^&*()_+-=[]{}|;:,.<>?",
		},
		{
			name:    "unicode characters",
			size:    5,
			charset: "αβγδεζηθικλμνξοπρστυφχψω",
		},
		{
			name:    "mixed unicode and ascii",
			size:    10,
			charset: "abc123αβγ",
		},
		{
			name:    "single character",
			size:    100,
			charset: "a",
		},
		{
			name:    "empty size",
			size:    0,
			charset: "abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.String(tt.size, tt.charset)

			// Check length
			if len([]rune(result)) != tt.size {
				t.Errorf("expected length %d, got %d", tt.size, len([]rune(result)))
			}

			// Check that all characters are from the charset
			for _, char := range result {
				if !strings.ContainsRune(tt.charset, char) {
					t.Errorf("character %c not in charset %s", char, tt.charset)
				}
			}
		})
	}
}

func TestRandomStringDistribution(t *testing.T) {
	r := randomPackage{}
	charset := "ab"
	size := 1000

	// Generate many single-character strings
	counts := make(map[rune]int)
	for i := 0; i < size; i++ {
		result := r.String(1, charset)
		if len(result) == 1 {
			counts[rune(result[0])]++
		}
	}

	// Check that both characters appear (very high probability)
	if len(counts) != 2 {
		t.Errorf("expected both characters to appear, got %d unique characters", len(counts))
	}

	// Check distribution is reasonable (not perfect due to randomness)
	for char, count := range counts {
		ratio := float64(count) / float64(size)
		if ratio < 0.3 || ratio > 0.7 {
			t.Errorf("character %c appeared %d times (%.2f%%), expected around 50%%",
				char, count, ratio*100)
		}
	}
}

func TestRandomMac(t *testing.T) {
	r := randomPackage{}
	macRegex := regexp.MustCompile(`^([0-9a-f]{2}:){5}[0-9a-f]{2}$`)

	// Generate multiple MAC addresses
	macs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		mac := r.Mac()

		// Check format
		if !macRegex.MatchString(mac) {
			t.Errorf("invalid MAC format: %s", mac)
		}

		// Check it's a valid MAC
		_, err := net.ParseMAC(mac)
		if err != nil {
			t.Errorf("invalid MAC address: %s, error: %v", mac, err)
		}

		// Store for uniqueness check
		macs[mac] = true
	}

	// Check that we get different MACs (very high probability)
	if len(macs) < 95 {
		t.Errorf("expected at least 95 unique MACs out of 100, got %d", len(macs))
	}
}

func TestRandomMacNormalization(t *testing.T) {
	r := randomPackage{}

	// Generate several MACs and check they're normalized
	for i := 0; i < 10; i++ {
		mac := r.Mac()

		// Check lowercase
		if mac != strings.ToLower(mac) {
			t.Errorf("MAC not normalized to lowercase: %s", mac)
		}

		// Check separator is colon
		if strings.Contains(mac, "-") {
			t.Errorf("MAC contains hyphen instead of colon: %s", mac)
		}

		// Check length
		if len(mac) != 17 { // 6 bytes * 2 chars + 5 colons
			t.Errorf("MAC has wrong length: %s (len=%d)", mac, len(mac))
		}
	}
}

func TestRandomStringEdgeCases(t *testing.T) {
	r := randomPackage{}

	// Test with various edge cases
	tests := []struct {
		name    string
		size    int
		charset string
	}{
		{
			name:    "zero size",
			size:    0,
			charset: "abc",
		},
		{
			name:    "very large size",
			size:    10000,
			charset: "abc",
		},
		{
			name:    "size larger than charset",
			size:    10,
			charset: "ab",
		},
		{
			name:    "single char charset with large size",
			size:    1000,
			charset: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.String(tt.size, tt.charset)

			if len([]rune(result)) != tt.size {
				t.Errorf("expected length %d, got %d", tt.size, len([]rune(result)))
			}

			// Check all characters are from charset
			for _, c := range result {
				if !strings.ContainsRune(tt.charset, c) {
					t.Errorf("character %c not in charset %s", c, tt.charset)
				}
			}
		})
	}
}

func TestRandomStringNegativeSize(t *testing.T) {
	r := randomPackage{}

	// Test that negative size causes panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative size but didn't get one")
		}
	}()

	// This should panic
	_ = r.String(-1, "abc")
}

func TestRandomPackageInstance(t *testing.T) {
	// Test that we can create multiple instances
	r1 := randomPackage{}
	r2 := randomPackage{}

	// Both should work independently
	s1 := r1.String(5, "abc")
	s2 := r2.String(5, "xyz")

	if len(s1) != 5 {
		t.Errorf("r1.String returned wrong length: %d", len(s1))
	}
	if len(s2) != 5 {
		t.Errorf("r2.String returned wrong length: %d", len(s2))
	}

	// Check correct charset usage
	for _, c := range s1 {
		if !strings.ContainsRune("abc", c) {
			t.Errorf("r1 produced character outside charset: %c", c)
		}
	}
	for _, c := range s2 {
		if !strings.ContainsRune("xyz", c) {
			t.Errorf("r2 produced character outside charset: %c", c)
		}
	}
}

func BenchmarkRandomString(b *testing.B) {
	r := randomPackage{}
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b.Run("size-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.String(10, charset)
		}
	})

	b.Run("size-100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.String(100, charset)
		}
	})

	b.Run("size-1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.String(1000, charset)
		}
	})
}

func BenchmarkRandomMac(b *testing.B) {
	r := randomPackage{}

	for i := 0; i < b.N; i++ {
		_ = r.Mac()
	}
}

func BenchmarkRandomStringCharsets(b *testing.B) {
	r := randomPackage{}

	charsets := map[string]string{
		"small":   "abc",
		"medium":  "abcdefghijklmnopqrstuvwxyz",
		"large":   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?",
		"unicode": "αβγδεζηθικλμνξοπρστυφχψωABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}

	for name, charset := range charsets {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = r.String(20, charset)
			}
		})
	}
}
