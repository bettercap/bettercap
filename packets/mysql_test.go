package packets

import (
	"bytes"
	"testing"
)

func TestMySQLConstants(t *testing.T) {
	// Test MySQLGreeting
	if len(MySQLGreeting) != 95 {
		t.Errorf("MySQLGreeting length = %d, want 95", len(MySQLGreeting))
	}
	// Check some key bytes in the greeting
	if MySQLGreeting[0] != 0x5b {
		t.Errorf("MySQLGreeting[0] = 0x%02x, want 0x5b", MySQLGreeting[0])
	}
	// Check version string starts at byte 5
	versionBytes := MySQLGreeting[5:12]
	expectedVersion := []byte("5.6.28-")
	if !bytes.Equal(versionBytes, expectedVersion) {
		t.Errorf("MySQL version = %s, want %s", versionBytes, expectedVersion)
	}

	// Test MySQLFirstResponseOK
	if len(MySQLFirstResponseOK) != 11 {
		t.Errorf("MySQLFirstResponseOK length = %d, want 11", len(MySQLFirstResponseOK))
	}
	// Check packet sequence number
	if MySQLFirstResponseOK[3] != 0x02 {
		t.Errorf("MySQLFirstResponseOK sequence = 0x%02x, want 0x02", MySQLFirstResponseOK[3])
	}

	// Test MySQLSecondResponseOK
	if len(MySQLSecondResponseOK) != 11 {
		t.Errorf("MySQLSecondResponseOK length = %d, want 11", len(MySQLSecondResponseOK))
	}
	// Check packet sequence number
	if MySQLSecondResponseOK[3] != 0x04 {
		t.Errorf("MySQLSecondResponseOK sequence = 0x%02x, want 0x04", MySQLSecondResponseOK[3])
	}
}

func TestMySQLGetFile(t *testing.T) {
	tests := []struct {
		name     string
		infile   string
		expected []byte
	}{
		{
			name:   "empty filename",
			infile: "",
			expected: []byte{
				0x01,                   // length + 1
				0x00, 0x00, 0x01, 0xfb, // header
			},
		},
		{
			name:   "short filename",
			infile: "test.txt",
			expected: []byte{
				0x09,                   // length of "test.txt" + 1 = 9
				0x00, 0x00, 0x01, 0xfb, // header
				't', 'e', 's', 't', '.', 't', 'x', 't',
			},
		},
		{
			name:   "path with directory",
			infile: "/etc/passwd",
			expected: []byte{
				0x0c,                   // length of "/etc/passwd" + 1 = 12
				0x00, 0x00, 0x01, 0xfb, // header
				'/', 'e', 't', 'c', '/', 'p', 'a', 's', 's', 'w', 'd',
			},
		},
		{
			name:   "windows path",
			infile: "C:\\Windows\\System32\\config\\sam",
			expected: []byte{
				0x1f,                   // length of path + 1 = 31
				0x00, 0x00, 0x01, 0xfb, // header
				'C', ':', '\\', 'W', 'i', 'n', 'd', 'o', 'w', 's', '\\',
				'S', 'y', 's', 't', 'e', 'm', '3', '2', '\\',
				'c', 'o', 'n', 'f', 'i', 'g', '\\', 's', 'a', 'm',
			},
		},
		{
			name:   "unicode filename",
			infile: "файл.txt",
			expected: func() []byte {
				filename := "файл.txt"
				result := []byte{
					byte(len(filename) + 1),
					0x00, 0x00, 0x01, 0xfb,
				}
				return append(result, []byte(filename)...)
			}(),
		},
		{
			name:   "max length filename",
			infile: string(make([]byte, 254)), // Max that fits in a single byte length
			expected: func() []byte {
				result := []byte{
					0xff, // 254 + 1 = 255
					0x00, 0x00, 0x01, 0xfb,
				}
				return append(result, make([]byte, 254)...)
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MySQLGetFile(tt.infile)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("MySQLGetFile(%q) = %v, want %v", tt.infile, result, tt.expected)
			}
		})
	}
}

func TestMySQLGetFileLength(t *testing.T) {
	// Test that the length byte is correctly calculated
	testCases := []struct {
		filename string
		expected byte
	}{
		{"", 0x01},
		{"a", 0x02},
		{"ab", 0x03},
		{"abc", 0x04},
		{"test.txt", 0x09},
		{string(make([]byte, 100)), 0x65}, // 100 + 1 = 101 = 0x65
		{string(make([]byte, 254)), 0xff}, // 254 + 1 = 255 = 0xff
	}

	for _, tc := range testCases {
		result := MySQLGetFile(tc.filename)
		if result[0] != tc.expected {
			t.Errorf("MySQLGetFile(%q) length byte = 0x%02x, want 0x%02x",
				tc.filename, result[0], tc.expected)
		}
	}
}

func TestMySQLGetFileHeader(t *testing.T) {
	// Test that the header bytes are always the same
	expectedHeader := []byte{0x00, 0x00, 0x01, 0xfb}

	filenames := []string{
		"",
		"test",
		"long_filename_with_many_characters.txt",
		"/path/to/file",
		"C:\\Windows\\file.exe",
	}

	for _, filename := range filenames {
		result := MySQLGetFile(filename)
		if len(result) < 5 {
			t.Errorf("MySQLGetFile(%q) returned packet too short: %d bytes", filename, len(result))
			continue
		}

		header := result[1:5]
		if !bytes.Equal(header, expectedHeader) {
			t.Errorf("MySQLGetFile(%q) header = %v, want %v", filename, header, expectedHeader)
		}
	}
}

func TestMySQLPacketStructure(t *testing.T) {
	// Test the overall packet structure
	filename := "test_file.sql"
	packet := MySQLGetFile(filename)

	// Check minimum packet size (1 byte length + 4 bytes header)
	if len(packet) < 5 {
		t.Fatalf("Packet too short: %d bytes", len(packet))
	}

	// Check that packet length matches expected
	expectedLen := 1 + 4 + len(filename) // length byte + header + filename
	if len(packet) != expectedLen {
		t.Errorf("Packet length = %d, want %d", len(packet), expectedLen)
	}

	// Check that the length byte correctly represents filename length + 1
	if packet[0] != byte(len(filename)+1) {
		t.Errorf("Length byte = %d, want %d", packet[0], len(filename)+1)
	}

	// Check that the filename is correctly appended
	filenameInPacket := string(packet[5:])
	if filenameInPacket != filename {
		t.Errorf("Filename in packet = %q, want %q", filenameInPacket, filename)
	}
}

func TestMySQLGreetingStructure(t *testing.T) {
	// Test specific parts of the MySQL greeting packet
	greeting := MySQLGreeting

	// The greeting should contain "mysql_native_password" at the end
	expectedSuffix := "mysql_native_password"
	suffixStart := len(greeting) - len(expectedSuffix) - 1 // -1 for null terminator
	suffix := string(greeting[suffixStart : suffixStart+len(expectedSuffix)])

	if suffix != expectedSuffix {
		t.Errorf("Greeting suffix = %q, want %q", suffix, expectedSuffix)
	}

	// Check null terminator
	if greeting[len(greeting)-1] != 0x00 {
		t.Error("Greeting should end with null terminator")
	}
}

// Benchmarks
func BenchmarkMySQLGetFile(b *testing.B) {
	filename := "/etc/passwd"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MySQLGetFile(filename)
	}
}

func BenchmarkMySQLGetFileShort(b *testing.B) {
	filename := "a.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MySQLGetFile(filename)
	}
}

func BenchmarkMySQLGetFileLong(b *testing.B) {
	filename := string(make([]byte, 200))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MySQLGetFile(filename)
	}
}
