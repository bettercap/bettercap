package core

import (
	"os"
	"testing"

	"github.com/evilsocket/islazy/fs"
)

func hasInt(a []int, v int) bool {
	for _, n := range a {
		if n == v {
			return true
		}
	}
	return false
}

func sameInts(a []int, b []int, ordered bool) bool {
	if len(a) != len(b) {
		return false
	}

	if ordered {
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
	} else {
		for _, v := range a {
			if !hasInt(b, v) {
				return false
			}
		}
	}

	return true
}

func TestCoreUniqueIntsUnsorted(t *testing.T) {
	var units = []struct {
		from []int
		to   []int
	}{
		{[]int{}, []int{}},
		{[]int{1, 1, 1, 1, 1}, []int{1}},
		{[]int{1, 2, 1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{[]int{4, 3, 4, 3, 2, 2}, []int{4, 3, 2}},
		{[]int{8, 3, 8, 4, 6, 1}, []int{8, 3, 4, 6, 1}},
	}

	for _, u := range units {
		got := UniqueInts(u.from, false)
		if !sameInts(got, u.to, false) {
			t.Fatalf("expected '%v', got '%v'", u.to, got)
		}
	}
}

func TestCoreUniqueIntsSorted(t *testing.T) {
	var units = []struct {
		from []int
		to   []int
	}{
		{[]int{}, []int{}},
		{[]int{1, 1, 1, 1, 1}, []int{1}},
		{[]int{1, 2, 1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{[]int{4, 3, 4, 3, 2, 2}, []int{2, 3, 4}},
		{[]int{8, 3, 8, 4, 6, 1}, []int{1, 3, 4, 6, 8}},
	}

	for _, u := range units {
		got := UniqueInts(u.from, true)
		if !sameInts(got, u.to, true) {
			t.Fatalf("expected '%v', got '%v'", u.to, got)
		}
	}
}

func TestCoreExists(t *testing.T) {
	var units = []struct {
		what   string
		exists bool
	}{
		{".", true},
		{"/", true},
		{"wuuut", false},
		{"/wuuu.t", false},
		{os.Args[0], true},
	}

	for _, u := range units {
		got := fs.Exists(u.what)
		if got != u.exists {
			t.Fatalf("expected '%v', got '%v'", u.exists, got)
		}
	}
}

func TestHasBinary(t *testing.T) {
	tests := []struct {
		name       string
		executable string
		expected   bool
	}{
		{
			name:       "common shell",
			executable: "sh",
			expected:   true,
		},
		{
			name:       "echo command",
			executable: "echo",
			expected:   true,
		},
		{
			name:       "non-existent binary",
			executable: "this-binary-definitely-does-not-exist-12345",
			expected:   false,
		},
		{
			name:       "empty string",
			executable: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasBinary(tt.executable)
			if got != tt.expected {
				t.Errorf("HasBinary(%q) = %v, want %v", tt.executable, got, tt.expected)
			}
		})
	}
}

func TestExec(t *testing.T) {
	tests := []struct {
		name       string
		executable string
		args       []string
		wantError  bool
		contains   string
	}{
		{
			name:       "echo with args",
			executable: "echo",
			args:       []string{"hello", "world"},
			wantError:  false,
			contains:   "hello world",
		},
		{
			name:       "echo empty",
			executable: "echo",
			args:       []string{},
			wantError:  false,
			contains:   "",
		},
		{
			name:       "non-existent command",
			executable: "this-command-does-not-exist-12345",
			args:       []string{},
			wantError:  true,
			contains:   "",
		},
		{
			name:       "true command",
			executable: "true",
			args:       []string{},
			wantError:  false,
			contains:   "",
		},
		{
			name:       "false command",
			executable: "false",
			args:       []string{},
			wantError:  true,
			contains:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip platform-specific commands if not available
			if !HasBinary(tt.executable) && !tt.wantError {
				t.Skipf("%s not found in PATH", tt.executable)
			}

			output, err := Exec(tt.executable, tt.args)

			if tt.wantError {
				if err == nil {
					t.Errorf("Exec(%q, %v) expected error but got none", tt.executable, tt.args)
				}
			} else {
				if err != nil {
					t.Errorf("Exec(%q, %v) unexpected error: %v", tt.executable, tt.args, err)
				}
				if tt.contains != "" && output != tt.contains {
					t.Errorf("Exec(%q, %v) = %q, want %q", tt.executable, tt.args, output, tt.contains)
				}
			}
		})
	}
}

func TestExecWithOutput(t *testing.T) {
	// Test that Exec properly captures and trims output
	if HasBinary("printf") {
		output, err := Exec("printf", []string{"  hello world  \n"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output != "hello world" {
			t.Errorf("expected trimmed output 'hello world', got %q", output)
		}
	}
}

func BenchmarkUniqueInts(b *testing.B) {
	// Create a slice with duplicates
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i % 100 // This creates 10 duplicates of each number 0-99
	}

	b.Run("unsorted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = UniqueInts(input, false)
		}
	})

	b.Run("sorted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = UniqueInts(input, true)
		}
	})
}
