package js

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

func TestBtoa(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello world",
			expected: base64.StdEncoding.EncodeToString([]byte("hello world")),
		},
		{
			name:     "empty string",
			input:    "",
			expected: base64.StdEncoding.EncodeToString([]byte("")),
		},
		{
			name:     "special characters",
			input:    "!@#$%^&*()_+-=[]{}|;:,.<>?",
			expected: base64.StdEncoding.EncodeToString([]byte("!@#$%^&*()_+-=[]{}|;:,.<>?")),
		},
		{
			name:     "unicode string",
			input:    "Hello 世界 🌍",
			expected: base64.StdEncoding.EncodeToString([]byte("Hello 世界 🌍")),
		},
		{
			name:     "newlines and tabs",
			input:    "line1\nline2\ttab",
			expected: base64.StdEncoding.EncodeToString([]byte("line1\nline2\ttab")),
		},
		{
			name:     "long string",
			input:    strings.Repeat("a", 1000),
			expected: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 1000))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create call with argument
			arg, _ := vm.ToValue(tt.input)
			call := otto.FunctionCall{
				ArgumentList: []otto.Value{arg},
			}

			result := btoa(call)

			// Check if result is an error
			if result.IsUndefined() {
				t.Fatal("btoa returned undefined")
			}

			// Get string value
			resultStr, err := result.ToString()
			if err != nil {
				t.Fatalf("failed to convert result to string: %v", err)
			}

			if resultStr != tt.expected {
				t.Errorf("btoa(%q) = %q, want %q", tt.input, resultStr, tt.expected)
			}
		})
	}
}

func TestAtob(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:     "simple base64",
			input:    base64.StdEncoding.EncodeToString([]byte("hello world")),
			expected: "hello world",
		},
		{
			name:     "empty base64",
			input:    base64.StdEncoding.EncodeToString([]byte("")),
			expected: "",
		},
		{
			name:     "special characters base64",
			input:    base64.StdEncoding.EncodeToString([]byte("!@#$%^&*()_+-=[]{}|;:,.<>?")),
			expected: "!@#$%^&*()_+-=[]{}|;:,.<>?",
		},
		{
			name:     "unicode base64",
			input:    base64.StdEncoding.EncodeToString([]byte("Hello 世界 🌍")),
			expected: "Hello 世界 🌍",
		},
		{
			name:      "invalid base64",
			input:     "not valid base64!",
			wantError: true,
		},
		{
			name:      "invalid padding",
			input:     "SGVsbG8gV29ybGQ", // Missing padding
			wantError: true,
		},
		{
			name:     "long base64",
			input:    base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 1000))),
			expected: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create call with argument
			arg, _ := vm.ToValue(tt.input)
			call := otto.FunctionCall{
				ArgumentList: []otto.Value{arg},
			}

			result := atob(call)

			// Get string value
			resultStr, err := result.ToString()
			if err != nil && !tt.wantError {
				t.Fatalf("failed to convert result to string: %v", err)
			}

			if tt.wantError {
				// Should return undefined (NullValue) on error
				if !result.IsUndefined() {
					t.Errorf("expected undefined for error case, got %q", resultStr)
				}
			} else {
				if resultStr != tt.expected {
					t.Errorf("atob(%q) = %q, want %q", tt.input, resultStr, tt.expected)
				}
			}
		})
	}
}

func TestGzipCompress(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple string",
			input: "hello world",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "repeated pattern",
			input: strings.Repeat("abcd", 100),
		},
		{
			name:  "random text",
			input: "The quick brown fox jumps over the lazy dog. " + strings.Repeat("Lorem ipsum dolor sit amet. ", 10),
		},
		{
			name:  "unicode text",
			input: "Hello 世界 🌍 " + strings.Repeat("测试数据 ", 50),
		},
		{
			name:  "binary-like data",
			input: string([]byte{0, 1, 2, 3, 255, 254, 253, 252}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create call with argument
			arg, _ := vm.ToValue(tt.input)
			call := otto.FunctionCall{
				ArgumentList: []otto.Value{arg},
			}

			result := gzipCompress(call)

			// Get compressed data
			compressed, err := result.ToString()
			if err != nil {
				t.Fatalf("failed to convert result to string: %v", err)
			}

			// Verify it's actually compressed (for non-empty strings, compressed should be different)
			if tt.input != "" && compressed == tt.input {
				t.Error("compressed data is same as input")
			}

			// Verify gzip header (should start with 0x1f, 0x8b)
			if len(compressed) >= 2 {
				if compressed[0] != 0x1f || compressed[1] != 0x8b {
					t.Error("compressed data doesn't have valid gzip header")
				}
			}

			// Now decompress to verify
			argCompressed, _ := vm.ToValue(compressed)
			callDecompress := otto.FunctionCall{
				ArgumentList: []otto.Value{argCompressed},
			}

			resultDecompressed := gzipDecompress(callDecompress)
			decompressed, err := resultDecompressed.ToString()
			if err != nil {
				t.Fatalf("failed to decompress: %v", err)
			}

			if decompressed != tt.input {
				t.Errorf("round-trip failed: got %q, want %q", decompressed, tt.input)
			}
		})
	}
}

func TestGzipCompressInvalidArgs(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name string
		args []otto.Value
	}{
		{
			name: "no arguments",
			args: []otto.Value{},
		},
		{
			name: "too many arguments",
			args: func() []otto.Value {
				arg1, _ := vm.ToValue("test")
				arg2, _ := vm.ToValue("extra")
				return []otto.Value{arg1, arg2}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := otto.FunctionCall{
				ArgumentList: tt.args,
			}

			result := gzipCompress(call)

			// Should return undefined (NullValue) on error
			if !result.IsUndefined() {
				resultStr, _ := result.ToString()
				t.Errorf("expected undefined for error case, got %q", resultStr)
			}
		})
	}
}

func TestGzipDecompress(t *testing.T) {
	vm := otto.New()

	// First compress some data
	originalData := "This is test data for decompression"
	arg, _ := vm.ToValue(originalData)
	compressCall := otto.FunctionCall{
		ArgumentList: []otto.Value{arg},
	}
	compressedResult := gzipCompress(compressCall)
	compressedData, _ := compressedResult.ToString()

	t.Run("valid decompression", func(t *testing.T) {
		argCompressed, _ := vm.ToValue(compressedData)
		decompressCall := otto.FunctionCall{
			ArgumentList: []otto.Value{argCompressed},
		}

		result := gzipDecompress(decompressCall)
		decompressed, err := result.ToString()
		if err != nil {
			t.Fatalf("failed to convert result to string: %v", err)
		}

		if decompressed != originalData {
			t.Errorf("decompressed data doesn't match original: got %q, want %q", decompressed, originalData)
		}
	})

	t.Run("invalid gzip data", func(t *testing.T) {
		argInvalid, _ := vm.ToValue("not gzip data")
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argInvalid},
		}

		result := gzipDecompress(call)

		// Should return undefined (NullValue) on error
		if !result.IsUndefined() {
			resultStr, _ := result.ToString()
			t.Errorf("expected undefined for error case, got %q", resultStr)
		}
	})

	t.Run("corrupted gzip data", func(t *testing.T) {
		// Create corrupted gzip by taking valid gzip and modifying it
		corruptedData := compressedData[:len(compressedData)/2] + "corrupted"

		argCorrupted, _ := vm.ToValue(corruptedData)
		call := otto.FunctionCall{
			ArgumentList: []otto.Value{argCorrupted},
		}

		result := gzipDecompress(call)

		// Should return undefined (NullValue) on error
		if !result.IsUndefined() {
			resultStr, _ := result.ToString()
			t.Errorf("expected undefined for error case, got %q", resultStr)
		}
	})
}

func TestGzipDecompressInvalidArgs(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name string
		args []otto.Value
	}{
		{
			name: "no arguments",
			args: []otto.Value{},
		},
		{
			name: "too many arguments",
			args: func() []otto.Value {
				arg1, _ := vm.ToValue("test")
				arg2, _ := vm.ToValue("extra")
				return []otto.Value{arg1, arg2}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := otto.FunctionCall{
				ArgumentList: tt.args,
			}

			result := gzipDecompress(call)

			// Should return undefined (NullValue) on error
			if !result.IsUndefined() {
				resultStr, _ := result.ToString()
				t.Errorf("expected undefined for error case, got %q", resultStr)
			}
		})
	}
}

func TestBtoaAtobRoundTrip(t *testing.T) {
	vm := otto.New()

	testStrings := []string{
		"simple",
		"",
		"with spaces and\nnewlines\ttabs",
		"special!@#$%^&*()_+-=[]{}|;:,.<>?",
		"unicode 世界 🌍",
		strings.Repeat("long string ", 100),
	}

	for _, original := range testStrings {
		t.Run(original, func(t *testing.T) {
			// Encode with btoa
			argOriginal, _ := vm.ToValue(original)
			encodeCall := otto.FunctionCall{
				ArgumentList: []otto.Value{argOriginal},
			}

			encoded := btoa(encodeCall)
			encodedStr, _ := encoded.ToString()

			// Decode with atob
			argEncoded, _ := vm.ToValue(encodedStr)
			decodeCall := otto.FunctionCall{
				ArgumentList: []otto.Value{argEncoded},
			}

			decoded := atob(decodeCall)
			decodedStr, _ := decoded.ToString()

			if decodedStr != original {
				t.Errorf("round-trip failed: got %q, want %q", decodedStr, original)
			}
		})
	}
}

func TestGzipCompressDecompressRoundTrip(t *testing.T) {
	vm := otto.New()

	testData := []string{
		"simple",
		"",
		strings.Repeat("repetitive data ", 100),
		"unicode 世界 🌍 " + strings.Repeat("测试 ", 50),
		string([]byte{0, 1, 2, 3, 255, 254, 253, 252}),
	}

	for _, original := range testData {
		t.Run(original, func(t *testing.T) {
			// Compress
			argOriginal, _ := vm.ToValue(original)
			compressCall := otto.FunctionCall{
				ArgumentList: []otto.Value{argOriginal},
			}

			compressed := gzipCompress(compressCall)
			compressedStr, _ := compressed.ToString()

			// Decompress
			argCompressed, _ := vm.ToValue(compressedStr)
			decompressCall := otto.FunctionCall{
				ArgumentList: []otto.Value{argCompressed},
			}

			decompressed := gzipDecompress(decompressCall)
			decompressedStr, _ := decompressed.ToString()

			if decompressedStr != original {
				t.Errorf("round-trip failed: got %q, want %q", decompressedStr, original)
			}
		})
	}
}

func BenchmarkBtoa(b *testing.B) {
	vm := otto.New()
	arg, _ := vm.ToValue("The quick brown fox jumps over the lazy dog")
	call := otto.FunctionCall{
		ArgumentList: []otto.Value{arg},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = btoa(call)
	}
}

func BenchmarkAtob(b *testing.B) {
	vm := otto.New()
	encoded := base64.StdEncoding.EncodeToString([]byte("The quick brown fox jumps over the lazy dog"))
	arg, _ := vm.ToValue(encoded)
	call := otto.FunctionCall{
		ArgumentList: []otto.Value{arg},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atob(call)
	}
}

func BenchmarkGzipCompress(b *testing.B) {
	vm := otto.New()
	data := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10)
	arg, _ := vm.ToValue(data)
	call := otto.FunctionCall{
		ArgumentList: []otto.Value{arg},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gzipCompress(call)
	}
}

func BenchmarkGzipDecompress(b *testing.B) {
	vm := otto.New()

	// First compress some data
	data := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10)
	argData, _ := vm.ToValue(data)
	compressCall := otto.FunctionCall{
		ArgumentList: []otto.Value{argData},
	}
	compressed := gzipCompress(compressCall)
	compressedStr, _ := compressed.ToString()

	// Benchmark decompression
	argCompressed, _ := vm.ToValue(compressedStr)
	decompressCall := otto.FunctionCall{
		ArgumentList: []otto.Value{argCompressed},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gzipDecompress(decompressCall)
	}
}
